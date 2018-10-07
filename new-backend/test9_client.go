package main

import (
	"fmt"
  zmq "github.com/pebbe/zmq4"
  "time"
)

var N int = 1

func main() {
  // fmt.Println("Connecting to hello world server...")
  publisher, _ := zmq.NewSocket(zmq.PUB)
  defer publisher.Close()
  publisher.Connect("ipc://test-pub-sub-5555.ipc")
  subscriber, _ := zmq.NewSocket(zmq.SUB)
	defer subscriber.Close()
  subscriber.Connect("ipc://test-pub-sub-5556.ipc")
  subscriber.SetSubscribe("SUBSCRIBE")

  time.Sleep(1.0 * time.Second)
  start := time.Now()

	for i := 0; i < N; i += 1 {
    publisher.Send("....CLAIM5", zmq.DONTWAIT)
  }

  elapsed := time.Since(start)
  fmt.Printf("TIME  : %s \n", elapsed)
}
