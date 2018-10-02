package main

import (
	// "fmt"
  zmq "github.com/pebbe/zmq4"
)

func main() {
  // fmt.Println("Connecting to hello world server...")
  publisher, _ := zmq.NewSocket(zmq.PUB)
  defer publisher.Close()
  publisher.Connect("ipc://test-pub-sub-5555.ipc")
  subscriber, _ := zmq.NewSocket(zmq.SUB)
	defer subscriber.Close()
  subscriber.Connect("ipc://test-pub-sub-5556.ipc")
  subscriber.SetSubscribe("....CLAIM")

	for {
		// msg, _ := subscriber.Recv(0)
    subscriber.Recv(0)
		// fmt.Printf("%s\n", msg)
    publisher.Send("SUBSCRIBE", zmq.DONTWAIT)
  }
}
