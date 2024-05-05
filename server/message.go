package server

// Message is the only sturct sent across the connections,
// everything needs to be embeded in Payload field
type Message struct {
	Payload any
}

// MessageStoreFile is the message sent to peers to
// notify the metadata of the file
type MessageStoreFile struct {
	Id string
	Key  string
	Size int64
}

// MessageGetFile is the message to get the file
// with the Key
type MessageGetFile struct {
	Key string
	Id string
}

// MessageDeleteKey is the message send
// to peer to delete the relevant key and Id
type MessageDeleteKey struct {
	Key string 
	Id string
}

// MessageFetch is the message send to
// peers to Fetch back all the content that belongs to the 
// sender content
type MessageFetch struct {
	Id string
}