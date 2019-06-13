package main

import (
	"fmt"
  zmq "github.com/pebbe/zmq4"
)

func main() {
  client, _ := zmq.NewSocket(zmq.ROUTER)
	defer client.Close()
  client.Bind("tcp://*:5570")

	for {
		msg, _ := client.RecvMessage(0)
    fmt.Println(msg)
    client.SendMessage(msg[0], fmt.Sprintf("hey"))
  }
}
