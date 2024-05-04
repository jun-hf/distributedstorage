# Distributed Content Addressable Storage
A distributed content addressable storage will encryptions.

## Installation
- git clone this repo
- Run `make run`

# Features
- Retrieve data with a key
- Data sending across the network is encrypted
- p2p package has a fully built TCP server
- Integrity of the data is fully secure by checking the checksum of every incoming read of the data.


## Usage

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
