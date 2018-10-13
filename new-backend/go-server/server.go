package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/alecthomas/participle"
	_ "github.com/mattn/go-sqlite3"
	zmq "github.com/pebbe/zmq4"

	// "github.com/alecthomas/repr"

	"sync"
	"time"
	// "github.com/pkg/profile"
)

var dbMutex sync.RWMutex
var zmqMutex sync.Mutex

type Term struct {
	Type  string
	Value string
}

type SelectQueryVariable struct {
	Fact     int
	Position int
	Equals   []SelectQueryVariable
}

type SubscriptionData struct {
	Id    string   `json:"id"`
	Facts []string `json:"facts"`
}

type Subscription struct {
	Source string
	Id     string
	Query  [][]Term
}

type Subscriptions struct {
	Subscriptions []Subscription
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func send_results(publisher *zmq.Socket, source string, id string, results [][]string) {
	// start := time.Now()
	results_json_str, err := json.Marshal(results)
	checkErr(err)
	msg := fmt.Sprintf("%s%s%s", source, id, string(results_json_str))
	// fmt.Println("Sending ", msg)
	zmqMutex.Lock()
	publisher.Send(msg, zmq.DONTWAIT)
	// fmt.Println(msg)
	zmqMutex.Unlock()
	// timeToSendResults := time.Since(start)
	// fmt.Printf("time to send results: %s \n", timeToSendResults)
}

func single_subscriber_update(db map[string]Fact, publisher *zmq.Socket, subscription Subscription, wg *sync.WaitGroup, i int) {
	start := time.Now()
	// fmt.Println("pre SELECTING %v", subscription.Query)
	query := make([]Fact, len(subscription.Query))
	for i, fact_terms := range subscription.Query {
		query[i] = Fact{fact_terms}
	}
	// dbMutex.RLock()
	results := select_facts(db, query)
	// dbMutex.RUnlock()
	selectDuration := time.Since(start)
	results_as_str := [][]string{}
	for _, result := range results {
		result_as_str := []string{}
		for _, result_term := range result.Result {
			result_as_str = append(result_as_str, result_term.Value)
		}
		results_as_str = append(results_as_str, result_as_str)
	}
	// fmt.Println("DONE SELECTING")
	send_results(publisher, subscription.Source, subscription.Id, results_as_str)
	wg.Done()
	duration := time.Since(start)
	fmt.Printf("SINGLE SUBSCRIBER DONE %v, select %v, send %v, total %s\n", i, selectDuration, duration-selectDuration, duration)
}

func update_all_subscriptions(db *map[string]Fact, publisher *zmq.Socket, subscriptions Subscriptions) {
	dbMutex.RLock()
	dbValue := make(map[string]Fact)
	for k, v := range *db {
		newTerms := make([]Term, len(v.Terms))
		for i, t := range newTerms {
			newTerms[i] = t
		}
		dbValue[k] = Fact{newTerms}
	}
	dbMutex.RUnlock()
	var wg sync.WaitGroup
	wg.Add(len(subscriptions.Subscriptions))
	// TODO: there may be a race condition if the contents of subscriptions changes when running this func.
	// How about just passing in a copy of the subscriptions
	for i, subscription := range subscriptions.Subscriptions {
		go single_subscriber_update(dbValue, publisher, subscription, &wg, i)
	}

	// fmt.Println("WAITING FOR ALL THINGS TO END")
	wg.Wait()
	// dbMutex.RUnlock()
	// dbMutex.RLock()
	// print_all_facts(*db)
	// dbMutex.RUnlock()
	// fmt.Println("done")
}

func subscribe_worker(subscription_messages <-chan string, claims chan<- []Term, subscriptions_notifications chan<- bool, parser *participle.Parser, publisher *zmq.Socket, subscriptions *Subscriptions) {
	event_type_len := 9
	source_len := 4
	for msg := range subscription_messages {
		fmt.Printf("SUBSCRIPTION SHOULD PARSE MESSAGE: %s\n", msg)
		event_type := msg[0:event_type_len]
		source := msg[event_type_len:(event_type_len + source_len)]
		val := msg[(event_type_len + source_len):]
		if event_type == "SUBSCRIBE" {
			subscription_data := SubscriptionData{}
			json.Unmarshal([]byte(val), &subscription_data)
			query := make([][]Term, 0)
			for i, fact_string := range subscription_data.Facts {
				subscription_fact_msg := fmt.Sprintf("subscription \"%s\" %v %s", subscription_data.Id, i, fact_string)
				subscription_fact := parse_fact_string(parser, subscription_fact_msg)
				subscription_fact = append([]Term{Term{"source", source}}, subscription_fact...)
				fmt.Printf("SUB FACTS %v\n", subscription_fact)
				claims <- subscription_fact
				fact := parse_fact_string(parser, fact_string)
				query = append(query, fact)
			}
			(*subscriptions).Subscriptions = append((*subscriptions).Subscriptions, Subscription{source, subscription_data.Id, query})
			subscriptions_notifications <- true // update_all_subscriptions(db, publisher, subscriptions)
		}
	}
}

func parser_worker(unparsed_messages <-chan string, claims chan<- []Term, parser *participle.Parser) {
	event_type_len := 9
	source_len := 4
	for msg := range unparsed_messages {
		fmt.Printf("SHOULD PARSE MESSAGE: %s\n", msg)
		event_type := msg[0:event_type_len]
		source := msg[event_type_len:(event_type_len + source_len)]
		val := msg[(event_type_len + source_len):]
		if event_type == "....CLAIM" {
			fact := parse_fact_string(parser, val)
			fact = append([]Term{Term{"source", source}}, fact...)
			claims <- fact
		}
	}
}

func claim_worker(claims <-chan []Term, subscriptions_notifications chan<- bool, db *map[string]Fact) {
	for fact_terms := range claims {
		fmt.Printf("SHOULD CLAIM: %v\n", claim)
		dbMutex.Lock()
		claim(db, Fact{fact_terms})
		dbMutex.Unlock()
		fmt.Println("claim done")
		subscriptions_notifications <- true // update_all_subscriptions(db, publisher, subscriptions)
	}
}

func notify_subscribers_worker(notify_subscribers <-chan bool, subscriber_worker_finished chan<- bool, db *map[string]Fact, publisher *zmq.Socket, subscriptions *Subscriptions) {
	// TODO: passing in subscriptions is probably not safe because it can be written in the other goroutine
	// db_copy := *db
	for range notify_subscribers {
		fmt.Println("inside notify_subscribers_worker loop")
		start := time.Now()
		update_all_subscriptions(db, publisher, *subscriptions)
		updateSubscribersTime := time.Since(start)
		fmt.Printf("update all subscribers time: %s \n", updateSubscribersTime)
		subscriber_worker_finished <- true
	}
}

func debounce_subscriber_worker(subscriptions_notifications <-chan bool, subscriber_worker_finished <-chan bool, notify_subscribers chan<- bool) {
	claim_waiting := false
	worker_is_free := true
	go func() {
		for range subscriptions_notifications {
			if worker_is_free {
				worker_is_free = false
				fmt.Println("notifying subscriber worker")
				notify_subscribers <- true
			} else {
				fmt.Println("(-) debouncing subscription notification becasue worker is busy")
				claim_waiting = true
			}
		}
	}()
	go func() {
		for range subscriber_worker_finished {
			fmt.Println("subscriber_worker_finished")
			worker_is_free = true
			if claim_waiting {
				fmt.Println("subscriber_worker_finished - running again to catch up to claims")
				claim_waiting = false
				notify_subscribers <- true
			}
		}
	}()
}

func main() {
	// defer profile.Start().Stop()
	parser, err := make_parser()
	checkErr(err)

	factDatabase := make_fact_database()

	subscriptions := Subscriptions{}

	// fmt.Println("Connecting to hello world server...")
	publisher, _ := zmq.NewSocket(zmq.PUB)
	defer publisher.Close()
	publisher.Bind("tcp://*:5555")
	subscriber, _ := zmq.NewSocket(zmq.SUB)
	defer subscriber.Close()
	subscriber.Bind("tcp://*:5556")
	subscriber.SetSubscribe(".....PING")
	subscriber.SetSubscribe("....CLAIM")
	subscriber.SetSubscribe("...SELECT")
	subscriber.SetSubscribe("..RETRACT")
	subscriber.SetSubscribe("SUBSCRIBE")

	event_type_len := 9
	source_len := 4

	unparsed_messages := make(chan string, 100)
	subscription_messages := make(chan string, 100)
	claims := make(chan []Term, 100)
	subscriptions_notifications := make(chan bool, 100)
	subscriber_worker_finished := make(chan bool, 99)
	notify_subscribers := make(chan bool, 99)

	go parser_worker(unparsed_messages, claims, parser)
	go subscribe_worker(subscription_messages, claims, subscriptions_notifications, parser, publisher, &subscriptions)
	go claim_worker(claims, subscriptions_notifications, &factDatabase)
	go notify_subscribers_worker(notify_subscribers, subscriber_worker_finished, &factDatabase, publisher, &subscriptions)
	go debounce_subscriber_worker(subscriptions_notifications, subscriber_worker_finished, notify_subscribers)

	for {
		// zmqMutex.Lock()
		msg, _ := subscriber.Recv(0)
		// zmqMutex.Unlock()
		// fmt.Printf("%s\n", msg)
		event_type := msg[0:event_type_len]
		source := msg[event_type_len:(event_type_len + source_len)]
		val := msg[(event_type_len + source_len):]
		if event_type == ".....PING" {
			fmt.Println("GOT PING!!!!!!!!!!!!!!!")
			fmt.Println(source)
			fmt.Println(val)
			send_results(publisher, source, val, make([][]string, 0))
		} else if event_type == "....CLAIM" {
			unparsed_messages <- msg
			// start := time.Now()
			// claim(db, parser, publisher, &subscriptions, val, source)
			// timeProcessing := time.Since(start)
			// fmt.Printf("processing: %s \n", timeProcessing)
			// } else if event_type == "..RETRACT" {
			//     retract(val)
			// } else if event_type == "...SELECT" {
			//     json_val = json.loads(val)
			//     select(json_val["facts"], json_val["id"], source)
		} else if event_type == "SUBSCRIBE" {
			subscription_messages <- msg
		}
		time.Sleep(1.0 * time.Microsecond)
	}
}
