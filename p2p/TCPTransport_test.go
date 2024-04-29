package p2p

import "testing"

func TestTCPTransport(t *testing.T) {
	opts := TCPTransportOpts{
		ListenAddr: ":8080",
		HandshakeFunc: NoHandshakeFunc,
		Decoder: DefaultDecoder{},
	}
}