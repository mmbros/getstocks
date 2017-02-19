// client.go
package main

import (
	"fmt"
	"log"
	"net"
	"net/rpc/jsonrpc"
)

type Args struct {
	X, Y int
}

func main() {

	client, err := net.Dial("tcp", "127.0.0.1:8888")
	if err != nil {
		log.Fatal("dialing:", err)

	}
	// Synchronous call
	var reply int
	c := jsonrpc.NewClient(client)
	err = c.Call("Sessions.Length", nil, &reply)
	if err != nil {
		log.Fatal("arith error:", err)

	}
	fmt.Printf("Result: %d\n", reply)

}
