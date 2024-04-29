package p2p

import "net"

// Peer is a remote node in the connections
type Peer interface {
	net.Conn
}

// Transport handles the any comunications between Peers
type Transport interface {
	Close() error
	Consume() 
}

// Rpc is the data passed between Peers
type Rpc struct {
}

