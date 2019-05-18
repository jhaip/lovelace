package main

import (
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"time"
	zmq "github.com/pebbe/zmq4"
)

func RandomString(len int) string {
      bytes := make([]byte, len)
     for i := 0; i < len; i++ {
          bytes[i] = byte(65 + rand.Intn(25))  //A=65 and Z = 65+25
      }
      return string(bytes)
}

func main() {
	N := 10
	MY_ID := os.Args[1]
	F, _ := strconv.Atoi(MY_ID)
	MY_ID_STR := fmt.Sprintf("%04d", F)
  F = F + N
	SUBSCRIPTION_ID_LEN := 36
	fmt.Println(MY_ID)
	fmt.Println(F)
	fmt.Println(MY_ID_STR)

	publisher, _ := zmq.NewSocket(zmq.PUB)
	defer publisher.Close()
	publisher.Connect("tcp://localhost:5556")
	subscriber, _ := zmq.NewSocket(zmq.SUB)
	defer subscriber.Close()
	subscriber.Connect("tcp://localhost:5555")
	subscriber.SetSubscribe(MY_ID_STR)

	init_ping_id := RandomString(SUBSCRIPTION_ID_LEN)
	source_len := 4
	server_send_time_len := 13

	time.Sleep(100.0 * time.Millisecond)

	publisher.Send(fmt.Sprintf(".....PING%s%s", MY_ID_STR, init_ping_id), zmq.DONTWAIT)
	fmt.Println("sent ping")

	for {
		msg, _ := subscriber.Recv(0)
		fmt.Println("Recv")
		fmt.Println(msg)
		id := msg[source_len:(source_len + SUBSCRIPTION_ID_LEN)]
		fmt.Println("ID:")
		fmt.Println(id)
		val := msg[(source_len + SUBSCRIPTION_ID_LEN + server_send_time_len):]
		if id == init_ping_id {
			fmt.Println("server is listening")
			break
		} else {
			fmt.Println(val)
		}
	}

	subscription_id := RandomString(SUBSCRIPTION_ID_LEN)
	sub_msg := fmt.Sprintf("{\"id\": \"%s\", \"facts\": [\"$ test client %s says $x @ $time1\", \"$ test client %d says $y @ $time2\"]}", subscription_id, MY_ID, F)
	publisher.Send(fmt.Sprintf("SUBSCRIBE%s%s", MY_ID_STR, sub_msg), zmq.DONTWAIT)
	fmt.Println("subscribe done")

	time.Sleep(100.0 * time.Millisecond)
	
	currentTimeMs := time.Now().UnixNano() / 1000000
	claim_msg := fmt.Sprintf("[{\"type\": \"claim\", \"fact\": [[\"text\", \"%s\"], [\"text\", \"test\"], [\"text\", \"client\"], [\"integer\", \"%s\"], [\"text\", \"says\"], [\"integer\", \"%s\"], [\"text\", \"@\"], [\"integer\", \"%d\"]]}]", MY_ID_STR, MY_ID, MY_ID, currentTimeMs)
	publisher.Send(fmt.Sprintf("....BATCH%s%s", MY_ID_STR, claim_msg), zmq.DONTWAIT)
	fmt.Println("startup claim done")
		
	for {
		msg, _ := subscriber.Recv(0)
		fmt.Println("Recv")
		fmt.Println(msg)
		id := msg[source_len:(source_len + SUBSCRIPTION_ID_LEN)]
		val := msg[(source_len + SUBSCRIPTION_ID_LEN + server_send_time_len):]
		if id == init_ping_id {
			fmt.Println("server is listening")
			fmt.Println(id)
		} else {
			fmt.Println(val)
		}
		break
	}
}