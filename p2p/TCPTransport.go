package p2p

import (
	"errors"
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
		log.Println("Closing peer connection:", conn.RemoteAddr().String())
		conn.Close()
	}()

	peer := NewTCPPeer(conn, inbound)
}