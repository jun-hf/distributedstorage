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

	server7000 := CreateServer(":7000", "7000-dir", []string{":8080", ":3030"})
	if err := server7000.Start(); err != nil {
		log.Fatal(err)
	}
	time.Sleep(1 * time.Second) // wait for all the server to initialize
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("item_%+v", i)
		data := fmt.Sprintf("big conten%+v", i)
		n, err := server7000.Store(key, strings.NewReader(data))
		if err != nil {
			fmt.Print(err)
		}
		fmt.Println("Server 7000 stream:", n)
	}
	server7000.Delete("item_1")
	fmt.Println("Sleepinggg 10 seconds")
	time.Sleep(10 * time.Second)
	r, err := server7000.Read("item_1")
	if err != nil {
		fmt.Println("Error", err)
	}

	b, _ := io.ReadAll(r)
	fmt.Println(string(b))
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
