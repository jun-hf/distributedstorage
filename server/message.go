package server

// Message is the only sturct sent across the connections,
// everything needs to be embeded in Payload field
type Message struct {
	Payload any
}

// MessageStoreFile is the message sent to peers to
// notify the metadata of the file
type MessageStoreFile struct {
	Key string 
	Size int64
}

// MessageGetFile is the message to get the file
// with the Key
type MessageGetFile struct {
	Key string
}