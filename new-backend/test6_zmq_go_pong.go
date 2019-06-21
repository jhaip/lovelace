package main

import (
  "fmt"
  "time"
  zmq "github.com/pebbe/zmq4"
)

func main() {
  client, _ := zmq.NewSocket(zmq.ROUTER)
	defer client.Close()
  client.Bind("tcp://*:5570")

  lastId := ""

  go func() {
		time.Sleep(5.0 * time.Second)
    client.SendMessage(lastId, "SPECIAL!")
	}()

	for {
		msg, _ := client.RecvMessage(0)
    fmt.Println(msg)
    client.SendMessage(msg[0], fmt.Sprintf("hey"))
    client.SendMessage(msg[0], fmt.Sprintf("yo"))
    lastId = msg[0]
  }
}
