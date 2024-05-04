package server

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/jun-hf/distributedstorage/cryto"
	"github.com/jun-hf/distributedstorage/p2p"
	"github.com/jun-hf/distributedstorage/store"
)

type ServerOpts struct {
	Transport         p2p.Transport
	Root              string
	OutboundServer    []string
	TransformPathFunc store.TransformPathFunc
	Id string
}

type Server struct {
	transport      p2p.Transport
	store          *store.Store
	quitCh         chan struct{}
	outboundServer []string
	encryptKey     []byte
	id string

	mu    sync.RWMutex
	peers map[string]p2p.Peer
}

func New(opts ServerOpts) *Server {
	store := store.New(store.StoreOpts{
		TransformPathFunc: opts.TransformPathFunc,
		Root:              opts.Root,
	})
	if len(opts.Id) == 0 {
		opts.Id = cryto.UUID()
	}
	return &Server{
		transport:      opts.Transport,
		store:          store,
		quitCh:         make(chan struct{}),
		outboundServer: opts.OutboundServer,
		encryptKey:     cryto.New(),
		peers:          make(map[string]p2p.Peer),
		id: opts.Id,
	}
}

func (s *Server) Start() error {
	if err := s.transport.ListenAndAccept(); err != nil {
		return err
	}
	go s.process()
	return s.dial()
}

func (s *Server) Delete(key string) error {
	if !s.store.Has(s.id, key) {
		return fmt.Errorf("%+v does not exists", key)
	}
	
	err := s.store.Delete(s.id, key)
	if err != nil {
		return err
	}

	msg := &Message{
		Payload: MessageDeleteKey{
			Key: cryto.Hash(key),
			Id: s.id,
		},
	}
	return s.broadcast(msg)
}

func (s *Server) Read(key string) (io.Reader, error) {
	if s.store.Has(s.id, key) {
		log.Printf("Getting key (%v) from local storage", key)
		return s.store.Read(s.id, key)
	}
	msg := &Message{
		Payload: MessageGetFile{
			Key: cryto.Hash(key),
			Id: s.id,
		},
	}
	if err := s.broadcast(msg); err != nil {
		return nil, err
	}

	s.mu.RLock()
	for _, peer := range s.peers {
		// Get the fileSize
		var fileSize int64
		binary.Read(peer, binary.LittleEndian, &fileSize)
		if _, err := s.store.WriteDecrypt(s.encryptKey, s.id, key, io.LimitReader(peer, fileSize)); err != nil {
			return nil, err
		}
		log.Printf("Getting key (%v) from remote storage", key)
		peer.Done()
	}
	s.mu.RUnlock()

	return s.store.Read(s.id, key)
}

// Store the content to the server and also the peers's server
// will return the amount of success store inclusive of the
// success store in the own server.
func (s *Server) Store(key string, data io.Reader) (int64, error) {
	succWrite := int64(0)
	dataBuff := new(bytes.Buffer)
	tee := io.TeeReader(data, dataBuff)
	n, err := s.store.Write(s.id, key, tee)
	if err != nil {
		return 0, err
	}
	succWrite++

	msg := &Message{
		Payload: MessageStoreFile{
			Id: s.id,
			Key:  cryto.Hash(key),
			Size: n + 16,
		},
	}
	if err := s.broadcast(msg); err != nil {
		return succWrite, err
	}
	time.Sleep(500 * time.Millisecond)
	return s.writeStream(dataBuff)
}

func (s *Server) writeStream(r io.Reader) (int64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	peerList := []io.Writer{}
	for _, p := range s.peers {
		peerList = append(peerList, p)
	}
	mw := io.MultiWriter(peerList...)
	_, err := io.Copy(mw, bytes.NewReader([]byte{p2p.IncomingStream}))
	if err != nil {
		return 0, err
	}
	n, err := cryto.CopyEncrypt(s.encryptKey, r, mw)
	return int64(n), err
}

func (s *Server) broadcast(m *Message) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	buff := new(bytes.Buffer)
	if err := gob.NewEncoder(buff).Encode(m); err != nil {
		return err
	}
	for addr, peer := range s.peers {
		if _, err := peer.Write([]byte{p2p.IncomingMessage}); err != nil {
			fmt.Printf("Write to %v failed: %v\n", addr, err)
			continue
		}
		if _, err := peer.Write(buff.Bytes()); err != nil {
			fmt.Printf("Write to %v failed: %v\n", addr, err)
			continue
		}
	}
	return nil
}

func (s *Server) process() {
	defer s.cleanUp()
	for {
		select {
		case rpc := <-s.transport.Consume():
			var msg Message
			if err := gob.NewDecoder(bytes.NewReader(rpc.Payload)).Decode(&msg); err != nil {
				log.Printf("Server (%v) decode error: %v\n", s.store.Root, err)
				continue
			}
			if err := s.handleMessage(msg, rpc.From); err != nil {
				log.Printf("Server (%v) handleMessage error: %v\n", s.store.Root, err)
				continue
			}
		case <-s.quitCh:
			return
		}
	}
}

func (s *Server) handleMessage(m Message, from string) error {
	switch payload := m.Payload.(type) {
	case MessageStoreFile:
		return s.handleMessageStoreFile(payload, from)
	case MessageGetFile:
		return s.handleMessageGetFile(payload, from)
	case MessageDeleteKey:
		return s.handleMessageDelete(payload)
	default:
		log.Println("No suitable Payload type")
		return nil
	}
}

func (s *Server) handleMessageDelete(m MessageDeleteKey) error {
	if !s.store.Has(m.Id, m.Key) {
		return fmt.Errorf("server (%v) do not have key: %v", s.store.Root, m.Key)
	}
	fmt.Println("Id:", m.Id)
	fmt.Println("Key:", m.Key)
	return s.store.Delete(m.Id, m.Key)
}

func (s *Server) handleMessageGetFile(m MessageGetFile, from string) error {
	if !s.store.Has(m.Id, m.Key) {
		return fmt.Errorf("server (%v) do not have key: %v", s.store.Root, m.Key)
	}

	p, err := s.getPeer(from)
	if err != nil {
		return err
	}

	size, err := s.store.FileSize(m.Id, m.Key)
	if err != nil {
		return err
	}

	p.Write([]byte{p2p.IncomingStream})
	// Sending the fileSize first after opening up the stream
	binary.Write(p, binary.LittleEndian, size)
	_, err = s.store.CopyRead(m.Id, m.Key, p)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) handleMessageStoreFile(m MessageStoreFile, from string) error {
	peer, err := s.getPeer(from)
	if err != nil {
		return err
	}
	defer peer.Done()
	fmt.Printf("peer: %v, size: %+v\n", peer.LocalAddr(), m.Size)
	n, err := s.store.Write(m.Id, m.Key, io.LimitReader(peer, m.Size))
	fmt.Println("Done")
	if err != nil {
		return fmt.Errorf("server (%v) write failed %v", s.store.Root, err)
	}
	log.Printf("server (%v) success store %v bytes\n", s.store.Root, n)
	return nil
}

func (s *Server) getPeer(from string) (p2p.Peer, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	peer, ok := s.peers[from]
	if !ok {
		return nil, fmt.Errorf("(%v) peer not in server (%v)'s connected peers", from, s.store.Root)
	}
	return peer, nil
}

func (s *Server) Close() {
	close(s.quitCh)
}

func (s *Server) cleanUp() {
	if err := s.transport.Close(); err != nil {
		log.Println("Error in closing:", err)
	}
	for _, peer := range s.peers {
		if err := peer.Close(); err != nil {
			log.Println("Error in closing peer:", err)
		}
	}
	log.Println("Server shutdown:", s.store.Root)
}

func (s *Server) OnPeer(p p2p.Peer) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.peers[p.RemoteAddr().String()] = p
	return nil
}

func (s *Server) dial() error {
	if len(s.outboundServer) == 0 {
		return nil
	}
	for _, addr := range s.outboundServer {
		if err := s.transport.Dial(addr); err != nil {
			log.Printf("server (%v) failed to dial %v: %v\n", s.store.Root, addr, err)
			continue
		}
	}
	return nil
}

func init() {
	gob.Register(MessageStoreFile{})
	gob.Register(MessageGetFile{})
	gob.Register(MessageDeleteKey{})
}
