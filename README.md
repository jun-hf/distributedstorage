# Distributed Content Addressable Storage
A distributed content addressable storage will encryptions.

## Installation
- git clone this repo
- Run `make run`

# Features
- Retrieve data with a key
- Data sending across the network is encrypted
- p2p package has a fully built TCP server
- Integrity of the data is fully secure by encryting all the data that is send across the network.


## Usage
### Create an new server:

`server.New(ServerOpts) *server.Server`, for Transport it takes in anything that implements the p2p.Transport
interface. This package comes with a default tcp server you can start using.

### Storing an key and data

`(*server.Server).Store(string, io.Reader)`, you can store the key and the associated data as
an io.Reader. This will store the content locally, and also encrypt the data and store it to all the remote peers.

### Deleting the key and data 

`(*server.Server).Delete(string)`, will delete the content stored locally and remotely.

### Reading the data

`(*server.Server).Read(string)`, will read the data associated with the given key. If in an event where the data is missing locally. It will ask the content from the remote nodes.

```go
	server8080 := CreateServer(":8080", "8080-dir", []string{})
	if err := server8080.Start(); err != nil {
		log.Fatal(err)
	}

	server3030 := CreateServer(":3030", "3030-dir", []string{":8080"})
	if err := server3030.Start(); err != nil {
		log.Fatal(err)
	}

	time.Sleep(1 * time.Second)
  	// will store the content with the key "Hello"
	n, err := server3030.Store("Hello", strings.NewReader("JIDJISED"))
	if err != nil {
	 	fmt.Print(err)
	 }
	 fmt.Println("Server 3030 stream:", n)

  	// will return a reader for the content in the key
	r, err := server3030.Read("Hello")
	if err != nil {
		log.Fatal(err)
	}
```
