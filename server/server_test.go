package server

import (
	"testing"

	"github.com/jun-hf/distributedstorage/p2p"
	"github.com/jun-hf/distributedstorage/store"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	server8080 := CreateServer(":8080", "8080-dir", []string{})
	assert.Nil(t, server8080.Start())

	server3030 := CreateServer(":3030", "3030-dir", []string{":8080"})
	assert.Nil(t, server3030.Start())
}

func CreateServer(listenAddr, root string, outboundServer []string) *Server {
	transport := p2p.NewTCPTransport(p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NoHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	})
	serverOpts := ServerOpts{
		Transport:         transport,
		Root:              root,
		OutboundServer:    outboundServer,
		TransformPathFunc: store.SHA1PathTransformFunc,
	}
	s1 := New(serverOpts)
	transport.OnPeer = s1.OnPeer
	return s1
}
