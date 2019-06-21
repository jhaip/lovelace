package main

import (
  "fmt"
  "math/rand"
  "os"
  "time"
  zmq "github.com/pebbe/zmq4"
)

func set_id(soc *zmq.Socket) {
	identity := fmt.Sprintf("%04X-%04X", rand.Intn(0x10000), rand.Intn(0x10000))
	soc.SetIdentity(identity)
}

var N int = 100
var i int = N

func main() {
  rand.Seed(time.Now().UnixNano())

  client, _ := zmq.NewSocket(zmq.DEALER)
  defer client.Close()

  set_id(client)
  client.Connect("tcp://localhost:5570")

  start := time.Now()
  client.SendMessage(fmt.Sprintf("hey"))

	for {
    msg, err := client.RecvMessage(0)
		if err == nil {
      id, _ := client.GetIdentity()
      fmt.Println(msg, id);
      // fmt.Println(msg[0], id)
      client.SendMessage(fmt.Sprintf("hey"))
      i -= 1
      // fmt.Printf("%s %v\n", msg, i)
      if i == 0 {
        elapsed := time.Since(start)
        fmt.Printf("TIME  : %s \n", elapsed)
        os.Exit(3)
      }
		}
  }
}
