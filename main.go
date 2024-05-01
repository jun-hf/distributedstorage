package main

import (
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/jun-hf/distributedstorage/p2p"
	"github.com/jun-hf/distributedstorage/server"
	"github.com/jun-hf/distributedstorage/store"
)

func main() {
	server8080 := CreateServer(":8080", "8080-dir", []string{})
	if err := server8080.Start(); err != nil {
		log.Fatal(err)
	}

	server3030 := CreateServer(":3030", "3030-dir", []string{":8080"})
	if err := server3030.Start(); err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)
	n, err := server3030.Store("Hello", strings.NewReader("JIDJISED"))
	if err != nil {
		fmt.Print(err)
	}
	fmt.Println("Server 3030 stream:", n)

	r, err := server3030.Read("Hello")
	if err != nil {
		log.Fatal(err)
	}

	b, _ := io.ReadAll(r)
	fmt.Println("Read: ", string(b))
	select {}
}

func CreateServer(listenAddr, root string, outboundServer []string) *server.Server {
	transport := p2p.NewTCPTransport(p2p.TCPTransportOpts{
		ListenAddr:    listenAddr,
		HandshakeFunc: p2p.NoHandshakeFunc,
		Decoder:       p2p.DefaultDecoder{},
	})
	serverOpts := server.ServerOpts{
		Transport:         transport,
		Root:              root,
		OutboundServer:    outboundServer,
		TransformPathFunc: store.SHA1PathTransformFunc,
	}
	s1 := server.New(serverOpts)
	transport.OnPeer = s1.OnPeer
	return s1
}
