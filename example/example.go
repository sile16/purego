package main

import (
	"fmt"
	//"github.com/davecgh/go-spew/spew"
	"github.com/sile16/purego/purego"
)

func main() {

	//client := purego.NewClientAPIToken("10.224.112.10", "ca65b5bb-66d3-9420-e4dc-ea67ef2e509d")

	//bad API key test:
	//client := purego.NewClientAPIToken("10.224.112.10", "ca6566d3-9420-e4dc-ea67ef2e509d")

	//bad API, but with backup username,password
	client := purego.NewClientUserPassAPIInsecure("10.224.112.10", "pureuser", "pureuser", "bad-api-key")

	//test concurrency
	//should see only 1 attempt to start session, but also run up to 10 concurrent HTTP requests.
	messages := make(chan string)
	numThreads := 1000
	//launch all 20 threads
	for x := 0; x < numThreads; x++ {
		go func(num int) {
			client.GetArray()
			messages <- fmt.Sprintf("Done:  %d ", num)
		}(x)
	}

	//print out thread results as they complete.
	for x := 0; x < numThreads; x++ {
		fmt.Println(<-messages)
	}

	//client := purego.NewClient("10.224.112.10")
	client.LogLevel = 4
	//client.StartSession()  // this is unessary, because client will start session on first API call.
	//spew.Dump(client.GetArray())
	//spew.Dump(client.GetVolumes()[0:2])
	//spew.Dump(client.GetArray())

}
