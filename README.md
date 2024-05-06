# Distributed Content Addressable Storage
A distributed content addressable file storage. It can store the data to the server with a key to access the data. The server will also stream the encrypted key and data to the rest of the connected peers to store it, making our system secure. In the event of failure of the local server where the content is loss, the local server can still reach out to the connected peers to restore the key and data making it fault tolerant.

## System Design
![image](https://github.com/jun-hf/distributedstorage/assets/86782267/8b193a27-2c87-41ed-9695-4206c4503bf6)

When you call `server.New(ServerOpts)*Server` it will return a server pointer. You need to pass in a `ServerOpts` into the function. I want to highlight the 2 important field in `ServerOpts` , which is `Transport` and `OutboundServer` . `Transport` is an interface that implements the `p2p.Transport` , in this repo I have already build a tcp transport that is ready to use in `p2p` . Next, the field `OutboundServer` takes in a list of ports to be connected. If you look at the diagram, you can see that the server :`7000` is able to connect to `:8080` , `:3030`

## Installation
- git clone this repo
- Run `make run`

# Features
- Retrieve data with a key
- Data sending across the network is encrypted
- p2p package has a fully built TCP server
- Integrity of the data is fully secure.



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
	for i := 0; i < 4; i++ {
		key := fmt.Sprintf("item_%+v", i)
		data := fmt.Sprintf("big conten%+v", i)
		n, err := server7000.Store(key, strings.NewReader(data))
		if err != nil {
			fmt.Print(err)
		}
		fmt.Println("Server 7000 stream:", n)
	}
	time.Sleep(5 * time.Second)
	server7000.Delete("item_1")
}
```
