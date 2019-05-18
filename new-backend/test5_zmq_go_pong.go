package main

import (
	"fmt"
  zmq "github.com/pebbe/zmq4"
  "time"
  "os"
)

var N int = 10
var i int = N

func main() {
  // fmt.Println("Connecting to hello world server...")
  publisher, _ := zmq.NewSocket(zmq.PUB)
  defer publisher.Close()
  publisher.Connect("ipc://test-pub-sub-5556.ipc")
  subscriber, _ := zmq.NewSocket(zmq.SUB)
	defer subscriber.Close()
  subscriber.Connect("ipc://test-pub-sub-555.ipc")
  subscriber.SetSubscribe("SUBSCRIBE")

  time.Sleep(1.0 * time.Second)
  start := time.Now()
  publisher.Send("....CLAIM5TEST", zmq.DONTWAIT)
  fmt.Printf("sent")

	for {
		msg, _ := subscriber.Recv(0)
    // subscriber.Recv(0)
		fmt.Printf("%s\n", msg)
    publisher.Send("....CLAIM5TEST", zmq.DONTWAIT)
    i -= 1
    // fmt.Printf("%s %v\n", msg, i)
    if i == 0 {
      elapsed := time.Since(start)
      fmt.Printf("TIME  : %s \n", elapsed)
      os.Exit(3)
    }
  }
}
