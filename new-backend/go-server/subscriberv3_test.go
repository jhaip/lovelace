package main

import (
	"sync"
	"testing"
	"time"
	"github.com/alecthomas/repr"
)

const CHANNEL_MESSAGE_DELIVERY_TEST_WAIT = time.Duration(10) * time.Millisecond

func makeFactDatabase() map[string]Fact {
	db := make(map[string]Fact)
	claim(&db, Fact{[]Term{Term{"text", "Sensor"}, Term{"text", "is"}, Term{"integer", "5"}}})
	claim(&db, Fact{[]Term{Term{"text", "Sensor"}, Term{"text", "is"}, Term{"text", "low"}}})
	claim(&db, Fact{[]Term{Term{"text", "low"}, Term{"text", "has"}, Term{"integer", "0"}}})
	return db
}

func TestMakeSubscriberV3(t *testing.T) {
	source := "1234"
	subscriptionId := "21dcca0a-ed5e-4593-b92e-fc9f16499cc8"
	query_part1 := []Term{
		Term{"variable", "A"},
		Term{"text", "is"},
		Term{"variable", "B"},
	}
	query_part2 := []Term{
		Term{"variable", "B"},
		Term{"text", "has"},
		Term{"variable", "C"},
	}
	query := [][]Term{query_part1, query_part2}
	subscription := Subscription{source, subscriptionId, query, make(chan []BatchMessage, 1000), &sync.WaitGroup{}, &sync.WaitGroup{}}
	subscription.dead.Add(1)
	subscription.warmed.Add(1)
	notifications := make(chan Notification, 1000)
	go startSubscriber(subscription, notifications, makeFactDatabase())

	time.Sleep(CHANNEL_MESSAGE_DELIVERY_TEST_WAIT)

	messages := make([]BatchMessage, 1)
	messages[0] = BatchMessage{"claim", [][]string{[]string{"text", "Sky"}, []string{"text", "is"}, []string{"text", "low"}}}
	subscription.batch_messages <- messages

	time.Sleep(CHANNEL_MESSAGE_DELIVERY_TEST_WAIT)

	if len(notifications) != 2 {
		t.Error("Wrong count of notifications", len(notifications))
	}
	notification := <-notifications
	repr.Println(notification, repr.Indent("  "), repr.OmitEmpty(true))
	notification2 := <-notifications
	repr.Println(notification2, repr.Indent("  "), repr.OmitEmpty(true))
}
