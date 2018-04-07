package main

import (
	//"fmt"
    "github.com/davecgh/go-spew/spew"
	"github.com/sile16/purego/purego"
)

func main() {
	//client := purego.NewClientAPIToken("10.224.112.10", "ca65b5bb-66d3-9420-e4dc-ea67ef2e509d")

	//bad API key test:
	//client := purego.NewClientAPIToken("10.224.112.10", "ca6566d3-9420-e4dc-ea67ef2e509d")

	//bad API, but with backup username,password
	client := purego.NewClientUserPassAPI("10.224.112.10", "pureuser","pureuser","bad-api-key")


	//client := purego.NewClient("10.224.112.10")
	client.LogLevel = 4
	//client.StartSession()
	spew.Dump(client.GetArray())
	//spew.Dump(client.GetVolumes()[0:2])

}
