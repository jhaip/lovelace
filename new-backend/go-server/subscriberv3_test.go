package main

import (
	"encoding/json"
	"sync"
	"testing"
	"time"
	"reflect"
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

func parseNotificationResult(notification Notification, t *testing.T) []map[string][]string {
	encoded_results := make([]map[string][]string, 0)
	err := json.Unmarshal([]byte(notification.Result), &encoded_results)
	if err != nil {
		t.Error("Error parsing notification result", notification)
	}
	return encoded_results
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
	go startSubscriberV3(subscription, notifications, makeFactDatabase())

	time.Sleep(CHANNEL_MESSAGE_DELIVERY_TEST_WAIT)

	messages := make([]BatchMessage, 1)
	messages[0] = BatchMessage{"claim", [][]string{[]string{"text", "Sky"}, []string{"text", "is"}, []string{"text", "low"}}}
	subscription.batch_messages <- messages

	time.Sleep(CHANNEL_MESSAGE_DELIVERY_TEST_WAIT)

	if len(notifications) != 2 {
		t.Error("Wrong count of notifications", len(notifications))
	}

	notification := <-notifications

	encoded_results := parseNotificationResult(notification, t)
	expectedResult := make([]map[string][]string, 1)
	expectedResult[0] = make(map[string][]string)
	expectedResult[0]["A"] = []string{"text", "Sensor"}
	expectedResult[0]["B"] = []string{"text", "low"}
	expectedResult[0]["C"] = []string{"integer", "0"}
	if !reflect.DeepEqual(expectedResult, encoded_results) {
		t.Error("Wrong notification result", expectedResult, encoded_results)
	}
	// repr.Println(notification, repr.Indent("  "), repr.OmitEmpty(true))
	// repr.Println(encoded_results, repr.Indent("  "), repr.OmitEmpty(true))

	notification2 := <-notifications
	encoded_results2 := parseNotificationResult(notification2, t)
	expectedResult2 := make([]map[string][]string, 2)
	expectedResult2[0] = make(map[string][]string)
	expectedResult2[0]["A"] = []string{"text", "Sensor"}
	expectedResult2[0]["B"] = []string{"text", "low"}
	expectedResult2[0]["C"] = []string{"integer", "0"}
	expectedResult2[1] = make(map[string][]string)
	expectedResult2[1]["A"] = []string{"text", "Sky"}
	expectedResult2[1]["B"] = []string{"text", "low"}
	expectedResult2[1]["C"] = []string{"integer", "0"}
	if !reflect.DeepEqual(expectedResult2, encoded_results2) {
		t.Error("Wrong notification result", expectedResult2, encoded_results2)
	}
	// repr.Println(notification2, repr.Indent("  "), repr.OmitEmpty(true))

	// Test that a repeated claim does not change the result:
	subscription.batch_messages <- messages

	time.Sleep(CHANNEL_MESSAGE_DELIVERY_TEST_WAIT)

	if len(notifications) != 0 {
		// notification3 := <-notifications
		// encoded_results3 := parseNotificationResult(notification3, t)
		// repr.Println(expectedResult2, repr.Indent("  "), repr.OmitEmpty(true))
		// repr.Println(encoded_results3, repr.Indent("  "), repr.OmitEmpty(true))
		// if !reflect.DeepEqual(expectedResult2, encoded_results3) {
		// 	t.Error("Wrong notification result", expectedResult2, encoded_results3)
		// }
		repr.Println("Wrong count of notifications")
		t.Error("Wrong count of notifications", len(notifications))
	}
}
