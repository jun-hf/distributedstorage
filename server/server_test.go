package server

import (
	"testing"

	"github.com/go-playground/locales/root"
	"github.com/jun-hf/distributedstorage/p2p"
	"github.com/jun-hf/distributedstorage/store"
	"github.com/stretchr/testify/assert"
)

func TestServer(t *testing.T) {
	server8080 := CreateServer(":8080", "8080-dir", []string{})
}

func CreateServer(listenAddr, root string, outboundServer []string) *Server{
	transport := p2p.NewTCPTransport(p2p.TCPTransportOpts{
		ListenAddr: listenAddr,
		HandshakeFunc: p2p.NoHandshakeFunc,
		Decoder: p2p.DefaultDecoder{},
	})
	serverOpts := ServerOpts{
		Transport: transport,
		Root: root,
		OutboundServer: outboundServer,
		TransformPathFunc: store.SHA1PathTransformFunc,
	}
	s1 := New(serverOpts)
	  
}
