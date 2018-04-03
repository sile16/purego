package main

import "github.com/sile16/purego/purego"

func main() {
	client := purego.NewClientAPIToken("10.224.112.10", "ca65b5bb-66d3-9420-e4dc-ea67ef2e509d")
	client.StartSession()
	client.GetArray()

}
