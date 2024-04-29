package p2p

import (
	"encoding/gob"
	"io"
	"net"
)

// Peer is a remote node in the connections
type Peer interface {
	net.Conn
	// Done is called when the receiving peer has
	// finished processing a stream
	Done()
}

// Transport handles any comunications between Peers
// This can be implemeted in TCP, UDP etc.
type Transport interface {
	ListenAndAccept() error
	Dial(string) error
	Close() error
	Consume() <-chan RPC
}

var (
	IncomingMessage = byte(1)
	IncomingStream = byte(2)
)

// RPC is the data passed between Peers
type RPC struct {
	Payload []byte
	Stream bool
	From string
}

// HandshakeFunc is used to shake hand between peers when connecting. 
type HandshakeFunc func(Peer) error

// NoHandshakeFunc can be used if hand shake between 
// Peers is not required.
func NoHandshakeFunc(Peer) error { return nil }

type Decoder interface {
	Decode(io.Reader, *RPC) error
}

type GOBDecoder struct {}

func (g GOBDecoder) Decode(r io.Reader, msg *RPC) error {
	return gob.NewDecoder(r).Decode(msg)
}

type DefaultDecoder struct {}

// Decode implements the Decoder interface, it 
// checks the first byte of r if it is 
// IncomingMessage or IncomingStream
func (dec DefaultDecoder) Decode(r io.Reader, rpc *RPC) error {
	peekBuf := make([]byte, 1)
	if _, err := r.Read(peekBuf); err != nil {
		return err
	}
	
	if peekBuf[0] == IncomingStream {
		rpc.Stream = true
		return nil
	}

	buf := make([]byte, 1028)
	n, err := r.Read(buf)
	if err != nil {
		return err
	}
	rpc.Payload = buf[:n]
	return nil
}