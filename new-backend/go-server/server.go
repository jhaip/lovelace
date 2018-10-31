package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"sort"

	_ "github.com/mattn/go-sqlite3"
	zmq "github.com/pebbe/zmq4"

	// "github.com/alecthomas/repr"

	"sync"
	"time"

	// "github.com/pkg/profile"
	b64 "encoding/base64"
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

type Notification struct {
	Source string
	Id     string
	Result string
}

type BatchMessage struct {
	Type string     `json:"type"`
	Fact [][]string `json:"fact"`
}

func checkErr(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func notification_worker(notifications <-chan Notification, retractions chan<- []Term) {
	// start := time.Now()
	publisher, _ := zmq.NewSocket(zmq.PUB)
	defer publisher.Close()
	publisher.Bind("tcp://*:5555")
	NO_RESULTS_MESSAGE := "[]"

	cache := make(map[string]string)

	for notification := range notifications {
		start := time.Now()
		msg := fmt.Sprintf("%s%s%s", notification.Source, notification.Id, notification.Result)
		cache_key := fmt.Sprintf("%s%s", notification.Source, notification.Id)
		cache_value, cache_hit := cache[cache_key]
		if cache_hit == false || cache_value != msg {
			// fmt.Printf("&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&&& setting cache: %v = %v\n", cache_key, msg)
			cache[cache_key] = msg

			// Clear all claims by source + subscription ID
			if cache_hit == true && cache_value[len(cache_value)-2:] != NO_RESULTS_MESSAGE {
				// Clear all claims by source + subscription ID
				// retractions <- []Term{Term{"id", notification.Source}, Term{"postfix", ""}}
			}
			if notification.Result != NO_RESULTS_MESSAGE {
				msgWithTime := fmt.Sprintf("%s%s%v%s", notification.Source, notification.Id, makeTimestamp(), notification.Result)
				// fmt.Printf("SENDING: \"%s\"\n", msgWithTime)
				publisher.Send(msgWithTime, zmq.DONTWAIT)
			}
			// } else {
			// 	fmt.Printf("SKIPPING BECAUSE DuPLICATE VALUE %v %v %v\n", cache_hit, cache_value, msg)
		}
		timeToSendResults := time.Since(start)
		fmt.Printf("______send: %s \n", timeToSendResults)
	}
	// timeToSendResults := time.Since(start)
	// fmt.Printf("time to send results: %s \n", timeToSendResults)
}

func marshal_query_result(query_results []QueryResult) string {
	encoded_results := make([]map[string][]string, 0)
	for _, query_result := range query_results {
		encoded_result := make(map[string][]string)
		for variable_name, term := range query_result.Result {
			encoded_result[variable_name] = []string{term.Type, term.Value}
		}
		encoded_results = append(encoded_results, encoded_result)
	}
	// fmt.Println(encoded_results)
	marshalled_results, err := json.Marshal(encoded_results)
	checkErr(err)
	// fmt.Println("MARSHALLD RESULTS:")
	// repr.Println(query_results, repr.Indent("  "), repr.OmitEmpty(true))
	// fmt.Println(string(marshalled_results))
	return string(marshalled_results)
}

func single_subscriber_update(db map[string]Fact, notifications chan<- Notification, subscription Subscription, wg *sync.WaitGroup, i int) {
	start := time.Now()
	// fmt.Println("pre SELECTING %v", subscription.Query)
	query := make([]Fact, len(subscription.Query))
	for i, fact_terms := range subscription.Query {
		query[i] = Fact{fact_terms}
	}
	// fmt.Println("QUERY:")
	// repr.Println(query, repr.Indent("  "), repr.OmitEmpty(true))
	// dbMutex.RLock()
	results := select_facts(db, query)
	fmt.Printf("GOT %v RESULTS", len(results))
	// dbMutex.RUnlock()
	selectDuration := time.Since(start)
	results_as_str := marshal_query_result(results)
	// fmt.Println("DONE SELECTING")
	notifications <- Notification{subscription.Source, subscription.Id, results_as_str}
	// print_all_facts(db)
	wg.Done()
	duration := time.Since(start)
	fmt.Printf("SINGLE SUBSCRIBER DONE %v, select %v, send %v, total %s\n", i, selectDuration, duration-selectDuration, duration)
}

func update_all_subscriptions(db *map[string]Fact, notifications chan<- Notification, subscriptions Subscriptions) {
	dbMutex.RLock()
	dbValue := make(map[string]Fact)
	for k, fact := range *db {
		newTerms := make([]Term, len(fact.Terms))
		for i, t := range fact.Terms {
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
		go single_subscriber_update(dbValue, notifications, subscription, &wg, i)
	}

	// fmt.Println("WAITING FOR ALL THINGS TO END")
	wg.Wait()
	// dbMutex.RUnlock()
	// dbMutex.RLock()
	// repr.Println(*db, repr.Indent("  "), repr.OmitEmpty(true))
	// print_all_facts(dbValue)
	// dbMutex.RUnlock()
	// fmt.Println("done")
}

func subscribe_worker(subscription_messages <-chan string, claims chan<- []Term, subscriptions_notifications chan<- bool, subscriptions *Subscriptions) {
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
				subscription_fact := parse_fact_string(subscription_fact_msg)
				subscription_fact = append([]Term{Term{"text", "subscription"}, Term{"id", source}}, subscription_fact...)
				fmt.Printf("SUB FACTS %v\n", subscription_fact)
				claims <- subscription_fact
				fact := parse_fact_string(fact_string)
				query = append(query, fact)
			}
			(*subscriptions).Subscriptions = append((*subscriptions).Subscriptions, Subscription{source, subscription_data.Id, query})
			subscriptions_notifications <- true
		}
	}
}

func parser_worker(unparsed_messages <-chan string, claims chan<- []Term, retractions chan<- []Term) {
	event_type_len := 9
	source_len := 4
	for msg := range unparsed_messages {
		start := time.Now()
		fmt.Printf("SHOULD PARSE MESSAGE: %s\n", msg)
		event_type := msg[0:event_type_len]
		source := msg[event_type_len:(event_type_len + source_len)]
		val := msg[(event_type_len + source_len):]
		if event_type == "....CLAIM" {
			fact := parse_fact_string(val)
			fact = append([]Term{Term{"id", source}}, fact...)
			claims <- fact
		} else if event_type == "..RETRACT" {
			fmt.Println("GOT RETRACT xxxxxxxxx")
			fact := parse_fact_string(val)
			retractions <- fact
		}
		elapsed := time.Since(start)
		log.Printf("______parse: %s \n", elapsed)
	}
}

func claim_worker(claims <-chan []Term, subscriptions_notifications chan<- bool, db *map[string]Fact) {
	for fact_terms := range claims {
		// fmt.Printf("SHOULD CLAIM: %v\n", claim)
		start := time.Now()
		dbMutex.Lock()
		fmt.Println("CLAIMED NEW FACT:")
		fmt.Println(fact_terms)
		claim(db, Fact{fact_terms})
		dbMutex.Unlock()
		elapsed := time.Since(start)
		log.Printf("______ claim down: %s \n", elapsed)
		// fmt.Println("claim done")
		subscriptions_notifications <- true
	}
}

func retract_worker(retractions <-chan []Term, subscriptions_notifications chan<- bool, db *map[string]Fact) {
	for fact_terms := range retractions {
		start := time.Now()
		dbMutex.Lock()
		fmt.Println("RETRACTING!!!")
		fmt.Println(fact_terms)
		fmt.Println(len(*db))
		retract(db, Fact{fact_terms})
		fmt.Println(len(*db))
		// print_all_facts(*db)
		dbMutex.Unlock()
		subscriptions_notifications <- true
		elapsed := time.Since(start)
		log.Printf("______retract: %s \n", elapsed)
	}
}

func notify_subscribers_worker(notify_subscribers <-chan bool, subscriber_worker_finished chan<- bool, db *map[string]Fact, notifications chan<- Notification, subscriptions *Subscriptions) {
	// TODO: passing in subscriptions is probably not safe because it can be written in the other goroutine
	// db_copy := *db
	for range notify_subscribers {
		fmt.Println("inside notify_subscribers_worker loop")
		start := time.Now()
		update_all_subscriptions(db, notifications, *subscriptions)
		updateSubscribersTime := time.Since(start)
		fmt.Printf("update all subscribers time: %s \n", updateSubscribersTime)
		log.Printf("______notify-subscribers: %s \n", updateSubscribersTime)
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

func debug_database_observer(db *map[string]Fact) {
	for {
		dbMutex.RLock()
		dbAsSstring := []byte("\033[H\033[2J") // clear terminal output on MacOS
		dbAsBase64Strings := ""
		var keys []string
		dbCopy := make(map[string]Fact)
		for k, fact := range *db {
			keys = append(keys, k)
			dbCopy[k] = fact
		}
		dbMutex.RUnlock()
		sort.Strings(keys)
		for _, fact_string := range keys {
			dbAsSstring = append(dbAsSstring, []byte(fact_string)...)
			dbAsSstring = append(dbAsSstring, '\n')
			dbAsBase64Strings += "["
			for i, term := range dbCopy[fact_string].Terms {
				if i > 0 {
					dbAsBase64Strings += ","
				}
				if term.Type == "text" {
					dbAsBase64Strings += fmt.Sprintf("[\"%s\", \"%s\"]", term.Type, b64.StdEncoding.EncodeToString([]byte(term.Value)))
				} else {
					dbAsBase64Strings += fmt.Sprintf("[\"%s\", \"%v\"]", term.Type, term.Value)
				}
			}
			dbAsBase64Strings += "]\n"
		}
		err := ioutil.WriteFile("./db_view.txt", dbAsSstring, 0644)
		checkErr(err)
		err2 := ioutil.WriteFile("./db_view_base64.txt", []byte(dbAsBase64Strings), 0644)
		checkErr(err2)
		time.Sleep(1.0 * time.Second)
	}
}

func batch_worker(batch_messages <-chan string, claims chan<- []Term, retractions chan<- []Term, subscriptions_notifications chan<- bool, db *map[string]Fact) {
	event_type_len := 9
	source_len := 4
	for msg := range batch_messages {
		// fmt.Printf("SHOULD PARSE MESSAGE: %s\n", msg)
		// event_type := msg[0:event_type_len]
		// source := msg[event_type_len:(event_type_len + source_len)]
		val := msg[(event_type_len + source_len):]
		// [["CLAIM", [["TEXT", "Hello"], ["INTEGER", "5"]]], ["RETRACT", [["VARIABLE", ""], ["INTEGER", "5"]]]]
		/*
			type BatchMessage struct {
				Type string     `json:"type"`
				Fact [][]string `json:"fact"`
			}
		*/
		var batch_messages []BatchMessage
		err := json.Unmarshal([]byte(val), &batch_messages)
		checkErr(err)
		// fmt.Println(batch_messages)
		for _, batch_message := range batch_messages {
			terms := make([]Term, len(batch_message.Fact))
			for j, term := range batch_message.Fact {
				terms[j] = Term{term[0], term[1]}
			}
			if batch_message.Type == "claim" {
				// claims <- terms
				dbMutex.Lock()
				claim(db, Fact{terms})
				dbMutex.Unlock()
			} else if batch_message.Type == "retract" {
				// retractions <- terms
				dbMutex.Lock()
				// fmt.Println(terms)
				// fmt.Println(len(*db))
				retract(db, Fact{terms})
				// fmt.Println(len(*db))
				// print_all_facts(*db)
				dbMutex.Unlock()
			}
		}
		subscriptions_notifications <- true
	}
}

func main() {
	// defer profile.Start().Stop()

	factDatabase := make_fact_database()

	subscriptions := Subscriptions{}

	// fmt.Println("Connecting to hello world server...")
	subscriber, _ := zmq.NewSocket(zmq.SUB)
	defer subscriber.Close()
	subscriber.Bind("tcp://*:5556")
	subscriber.SetSubscribe(".....PING")
	subscriber.SetSubscribe("....CLAIM")
	subscriber.SetSubscribe("....BATCH")
	subscriber.SetSubscribe("...SELECT")
	subscriber.SetSubscribe("..RETRACT")
	subscriber.SetSubscribe("SUBSCRIBE")

	event_type_len := 9
	source_len := 4

	unparsed_messages := make(chan string, 100)
	subscription_messages := make(chan string, 100)
	claims := make(chan []Term, 100)
	retractions := make(chan []Term, 100)
	subscriptions_notifications := make(chan bool, 100)
	subscriber_worker_finished := make(chan bool, 99)
	notify_subscribers := make(chan bool, 99)
	notifications := make(chan Notification, 1000)
	batch_messages := make(chan string, 100)

	go parser_worker(unparsed_messages, claims, retractions)
	go subscribe_worker(subscription_messages, claims, subscriptions_notifications, &subscriptions)
	go claim_worker(claims, subscriptions_notifications, &factDatabase)
	go retract_worker(retractions, subscriptions_notifications, &factDatabase)
	go notify_subscribers_worker(notify_subscribers, subscriber_worker_finished, &factDatabase, notifications, &subscriptions)
	go debounce_subscriber_worker(subscriptions_notifications, subscriber_worker_finished, notify_subscribers)
	go notification_worker(notifications, retractions)
	go debug_database_observer(&factDatabase)
	go batch_worker(batch_messages, claims, retractions, subscriptions_notifications, &factDatabase)

	go func() {
		for {
			fmt.Println("kick it!")
			subscriptions_notifications <- true
			time.Sleep(1.0 * time.Second)
		}
	}()

	for {
		msg, _ := subscriber.Recv(0)
		// fmt.Printf("%s\n", msg)
		event_type := msg[0:event_type_len]
		source := msg[event_type_len:(event_type_len + source_len)]
		val := msg[(event_type_len + source_len):]
		if event_type == ".....PING" {
			fmt.Println("GOT PING!!!!!!!!!!!!!!!")
			fmt.Println(source)
			fmt.Println(val)
			notifications <- Notification{source, val, ""}
		} else if event_type == "....CLAIM" {
			unparsed_messages <- msg
			// start := time.Now()
			// claim(db, parser, publisher, &subscriptions, val, source)
			// timeProcessing := time.Since(start)
			// fmt.Printf("processing: %s \n", timeProcessing)
		} else if event_type == "..RETRACT" {
			fmt.Println("GOT RETRACT preee xxxxxxxxx")
			unparsed_messages <- msg
			// } else if event_type == "...SELECT" {
			//     json_val = json.loads(val)
			//     select(json_val["facts"], json_val["id"], source)
		} else if event_type == "SUBSCRIBE" {
			subscription_messages <- msg
		} else if event_type == "....BATCH" {
			batch_messages <- msg
		}
		time.Sleep(1.0 * time.Microsecond)
	}
}
