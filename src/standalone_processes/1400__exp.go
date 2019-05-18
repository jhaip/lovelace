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
	SUBSCRIPTION_ID_LEN := 16
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

	time.Sleep(10.0 * time.Millisecond)

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
	
	// # time.sleep(10)
    // logging.error("sending #1")
    // currentTimeMs = int(round(time.time() * 1000))
    // claims = []
    // claims.append({"type": "claim", "fact": [
    //     ["text", get_my_id_str()],
    //     ["text", "test"],
    //     ["text", "client"],
    //     ["integer", MY_ID],
    //     ["text", "says"],
    //     ["integer", MY_ID],
    //     ["text", "@"],
    //     ["integer", str(currentTimeMs)]
    // ]})
    // batch(claims)
    // print("1300 claim")
}