package p2p

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
)

type TCPPeer struct {
	net.Conn
	inbound bool
	wg *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, inbound bool) *TCPPeer {
	return &TCPPeer{
		Conn: conn,
		inbound: inbound,
		wg: &sync.WaitGroup{},
	}
}

func (t *TCPPeer) Done() {
	t.wg.Done()
}

type TCPTransportOpts struct {
	ListenAddr string
	HandshakeFunc HandshakeFunc
	Decoder Decoder
	OnPeer func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	incomingRpc chan RPC
	listener net.Listener
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		incomingRpc: make(chan RPC),
	}
}

func (t *TCPTransport) ListenAndAccept() error {
	var err error
	t.listener, err = net.Listen("tcp", t.ListenAddr)
	if err != nil {
		return err
	}

	go t.acceptLoop()

	log.Println("Started tcp transport at:", t.listener.Addr())
	return nil
}

func (t *TCPTransport) acceptLoop() {
	for {
		conn, err := t.listener.Accept()
		if errors.Is(err, net.ErrClosed) {
			log.Printf("Closing TCP server %v\n", t.ListenAddr)
			return
		}
		if err != nil {
			log.Printf("TCP server %v failed to accept connection: %v\n", t.ListenAddr, err)
			continue
		}
		go t.handleConnection(conn, true)
	}
}

func (t *TCPTransport) handleConnection(conn net.Conn, inbound bool) {
	var err error
	defer func() {
		log.Printf("Closing peer (%v) connection: %v\n", conn.RemoteAddr().String(), err)
		conn.Close()
	}()

	peer := NewTCPPeer(conn, inbound)
	if err = t.HandshakeFunc(peer); err != nil {
		return
	}

	if t.OnPeer != nil {
		if err = t.OnPeer(peer); err != nil {
			return
		}
	}

	for {
		rpc := RPC{}
		err := t.Decoder.Decode(conn, &rpc)
		if err != nil {
			return
		}
		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Println("streaming from:", peer.RemoteAddr().String())
			peer.wg.Done()
			continue
		}
		rpc.From = peer.RemoteAddr().Network()
		t.incomingRpc <- rpc
	}
}