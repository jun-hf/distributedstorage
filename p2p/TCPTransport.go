package p2p

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
)

type TCPPeer struct {
	net.Conn
	inbound bool
	wg      *sync.WaitGroup
}

func NewTCPPeer(conn net.Conn, inbound bool) *TCPPeer {
	return &TCPPeer{
		Conn:    conn,
		inbound: inbound,
		wg:      &sync.WaitGroup{},
	}
}

func (t *TCPPeer) Done() {
	t.wg.Done()
}

type TCPTransportOpts struct {
	ListenAddr    string
	HandshakeFunc HandshakeFunc
	Decoder       Decoder
	OnPeer        func(Peer) error
}

type TCPTransport struct {
	TCPTransportOpts
	incomingRpc chan RPC
	listener    net.Listener
}

func NewTCPTransport(opts TCPTransportOpts) *TCPTransport {
	return &TCPTransport{
		TCPTransportOpts: opts,
		incomingRpc:      make(chan RPC),
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

func (t *TCPTransport) Dial(addr string) error {
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return err
	}
	go t.handleConnection(conn, false)
	return nil
}

func (t *TCPTransport) Close() error {
	return t.listener.Close()
}

func (t *TCPTransport) Consume() <-chan RPC {
	return t.incomingRpc
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
	fmt.Printf("%v had established connection from %v\n", conn.LocalAddr().String(), conn.RemoteAddr().String())
	for {
		rpc := RPC{}
		rpc.From = peer.RemoteAddr().String()
		err = t.Decoder.Decode(conn, &rpc)
		if err == io.EOF || errors.Is(err, net.ErrClosed) {
			return
		}
		if err != nil {
			continue
		}
		if rpc.Stream {
			peer.wg.Add(1)
			fmt.Println("streaming from:", peer.RemoteAddr().String())
			peer.wg.Wait()
			fmt.Println("stream completed from:", peer.RemoteAddr().String())
			continue
		}
		t.incomingRpc <- rpc
	}
}
