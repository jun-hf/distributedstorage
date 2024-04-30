package server

import (
	"fmt"
	"log"
	"sync"

	"github.com/jun-hf/distributedstorage/p2p"
	"github.com/jun-hf/distributedstorage/store"
)

// the goal
// Create a server with test that get write and stream
// Create a server struct Done
// Now create a test for creating a server
// Next, create Start function
// What do I want my start function to do?
// It should dail all the peers from the list of outbound node start the
// transport server and Accept connection

// How can we pass OnPeer
// Think about how does closing affects the peers

type ServerOpts struct {
	Transport p2p.Transport
	Root string
	OutboundServer []string
	TransformPathFunc store.TransformPathFunc
}

type Server struct {
	transport p2p.Transport
	store *store.Store
	quitCh chan struct{}
	outboundServer []string

	mu sync.Mutex
	peers map[string]p2p.Peer
}

func New(opts ServerOpts) *Server {
	store := store.New(store.StoreOpts{
		TransformPathFunc: opts.TransformPathFunc,
		Root: opts.Root,
	})
	return &Server{
		transport: opts.Transport,
		store: store,
		quitCh: make(chan struct{}),
		outboundServer: opts.OutboundServer,
		peers: make(map[string]p2p.Peer),
	}
}


func (s *Server) Start() error {
	if err := s.transport.ListenAndAccept(); err != nil {
		return err
	}
	go s.process()
	return s.dial()
}

func (s *Server) process() {
	defer s.cleanUp()
	for {
		select {
		case rpc := <- s.transport.Consume():
			fmt.Println(rpc)
		case <-s.quitCh:
			return
		}
	}
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