package p2p

import (
	"testing"
	"github.com/stretchr/testify/assert"
)

func TestTCPTransport(t *testing.T) {
	opts := TCPTransportOpts{
		ListenAddr: ":8080",
		HandshakeFunc: NoHandshakeFunc,
		Decoder: DefaultDecoder{},
	}
	tcp := NewTCPTransport(opts)
	assert.Equal(t, tcp.ListenAddr, ":8080")
	assert.Nil(t, tcp.ListenAndAccept(), )
}