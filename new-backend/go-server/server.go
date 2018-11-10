package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"sort"

	_ "github.com/mattn/go-sqlite3"
	zmq "github.com/pebbe/zmq4"

	"sync"
	"time"

	// "github.com/pkg/profile"
	b64 "encoding/base64"

	"go.uber.org/zap"
)

var dbMutex sync.RWMutex
var zmqMutex sync.Mutex

const MeasurementDuration = 5.0 * time.Second

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
	Source         string
	Id             string
	Query          [][]Term
	batch_messages chan []BatchMessage
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

type LatencyMeasurer struct {
	lastMeasuredTime    time.Time
	count               int64
	lastDbLockWaitTime  time.Time
	dbLockWaitTime      time.Duration
	lastActionTime      time.Time
	actionTime          time.Duration
	lastMessageWaitTime time.Time
	messageWaitTime     time.Duration
}

func checkErr(err error) {
	if err != nil {
		// log.Fatal(err)
		zap.L().Fatal("FATAL ERROR", zap.Error(err))
	}
}

func makeTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func makeLatencyMeasurer() LatencyMeasurer {
	return LatencyMeasurer{time.Now(), 0, time.Time{}, 0, time.Time{}, 0, time.Time{}, 0}
}

func preLatencyMeasurePart(subsystemName string, latencyMeasurer LatencyMeasurer) LatencyMeasurer {
	if subsystemName == "db" {
		latencyMeasurer.lastDbLockWaitTime = time.Now()
	} else if subsystemName == "action" {
		latencyMeasurer.lastActionTime = time.Now()
	} else if subsystemName == "messageWait" {
		latencyMeasurer.lastMessageWaitTime = time.Now()
	}
	return latencyMeasurer
}

func postLatencyMeasurePart(subsystemName string, latencyMeasurer LatencyMeasurer) LatencyMeasurer {
	if subsystemName == "db" {
		latencyMeasurer.dbLockWaitTime += time.Since(latencyMeasurer.lastDbLockWaitTime)
	} else if subsystemName == "action" {
		latencyMeasurer.actionTime += time.Since(latencyMeasurer.lastActionTime)
	} else if subsystemName == "messageWait" {
		latencyMeasurer.messageWaitTime += time.Since(latencyMeasurer.lastMessageWaitTime)
	}
	return latencyMeasurer
}

func updateLatencyMeasurer(latencyMeasurer LatencyMeasurer, actionsDone int64, subsystemName string) LatencyMeasurer {
	var count int64
	count = latencyMeasurer.count + actionsDone
	if time.Since(latencyMeasurer.lastMeasuredTime) > MeasurementDuration {
		zap.L().Debug(subsystemName,
			zap.Duration("latency",
				time.Duration(int64(time.Since(latencyMeasurer.lastMeasuredTime))/count)),
			zap.Int64("count", count),
			zap.Duration("dbLockWaitTime",
				time.Duration(int64(latencyMeasurer.dbLockWaitTime)/count)),
			zap.Duration("actionTime",
				time.Duration(int64(latencyMeasurer.actionTime)/count)),
			zap.Duration("messageWaitTime",
				time.Duration(int64(latencyMeasurer.messageWaitTime)/count)),
		)
		return makeLatencyMeasurer()
	}
	latencyMeasurer.count = count
	return latencyMeasurer
}

func notification_worker(notifications <-chan Notification, retractions chan<- []Term) {
	publisher, _ := zmq.NewSocket(zmq.PUB)
	defer publisher.Close()
	publisher.Bind("tcp://*:5555")
	NO_RESULTS_MESSAGE := "[]"
	cache := make(map[string]string)
	latencyMeasurer := makeLatencyMeasurer()
	latencyMeasurer = preLatencyMeasurePart("messageWait", latencyMeasurer)
	for notification := range notifications {
		latencyMeasurer = postLatencyMeasurePart("messageWait", latencyMeasurer)
		latencyMeasurer = preLatencyMeasurePart("action", latencyMeasurer)
		start := time.Now()
		msg := fmt.Sprintf("%s%s%s", notification.Source, notification.Id, notification.Result)
		cache_key := fmt.Sprintf("%s%s", notification.Source, notification.Id)
		cache_value, cache_hit := cache[cache_key]
		if cache_hit == false || cache_value != msg {
			cache[cache_key] = msg
			if notification.Result != NO_RESULTS_MESSAGE {
				msgWithTime := fmt.Sprintf("%s%s%v%s", notification.Source, notification.Id, makeTimestamp(), notification.Result)
				publisher.Send(msgWithTime, zmq.DONTWAIT)
			}
		}
		timeToSendResults := time.Since(start)
		zap.L().Debug("send notification", zap.Duration("timeToSendResults", timeToSendResults))
		latencyMeasurer = postLatencyMeasurePart("action", latencyMeasurer)
		latencyMeasurer = updateLatencyMeasurer(latencyMeasurer, 1, "latency - send notification")
		latencyMeasurer = preLatencyMeasurePart("messageWait", latencyMeasurer)
	}
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
	marshalled_results, err := json.Marshal(encoded_results)
	checkErr(err)
	return string(marshalled_results)
}

func single_subscriber_update(db map[string]Fact, notifications chan<- Notification, subscription Subscription, wg *sync.WaitGroup, i int) {
	start := time.Now()
	query := make([]Fact, len(subscription.Query))
	for i, fact_terms := range subscription.Query {
		query[i] = Fact{fact_terms}
	}
	results := select_facts(db, query)
	selectDuration := time.Since(start)
	results_as_str := marshal_query_result(results)
	notifications <- Notification{subscription.Source, subscription.Id, results_as_str}
	wg.Done()
	duration := time.Since(start)
	zap.L().Debug("SINGLE SUBSCRIBER DONE",
		zap.Int("sub_index", i),
		zap.String("source", subscription.Source),
		zap.Duration("select", selectDuration),
		zap.Duration("send", duration-selectDuration),
		zap.Duration("total", duration))
}

func update_all_subscriptions(db *map[string]Fact, notifications chan<- Notification, subscriptions Subscriptions, latencyMeasurer LatencyMeasurer) LatencyMeasurer {
	latencyMeasurer = preLatencyMeasurePart("db", latencyMeasurer)
	dbMutex.RLock()
	dbValue := make(map[string]Fact)
	for k, fact := range *db {
		newTerms := make([]Term, len(fact.Terms))
		for i, t := range fact.Terms {
			newTerms[i] = t
		}
		dbValue[k] = Fact{newTerms}
	}
	latencyMeasurer = postLatencyMeasurePart("db", latencyMeasurer)
	dbMutex.RUnlock()
	latencyMeasurer = preLatencyMeasurePart("action", latencyMeasurer)
	var wg sync.WaitGroup
	wg.Add(len(subscriptions.Subscriptions))
	zap.L().Debug("SINGLE SUBSCRIBER -- begin")
	// TODO: there may be a race condition if the contents of subscriptions changes when running this func.
	// How about just passing in a copy of the subscriptions
	for i, subscription := range subscriptions.Subscriptions {
		go single_subscriber_update(dbValue, notifications, subscription, &wg, i)
	}
	wg.Wait()
	zap.L().Debug("SINGLE SUBSCRIBER -- end")
	latencyMeasurer = postLatencyMeasurePart("action", latencyMeasurer)
	return latencyMeasurer
}

func subscribe_worker(subscription_messages <-chan string,
	claims chan<- []Term,
	subscriptions_notifications chan<- bool,
	subscriptions *Subscriptions,
	notifications chan<- Notification,
	db *map[string]Fact) {

	event_type_len := 9
	source_len := 4
	for msg := range subscription_messages {
		zap.L().Debug("SUBSCRIPTION SHOULD PARSE MESSAGE", zap.String("msg", msg))
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
				claims <- subscription_fact
				fact := parse_fact_string(fact_string) // AVOID DOUBLE PARSING!, this work was already done above
				query = append(query, fact)
			}
			newSubscription := Subscription{source, subscription_data.Id, query, make(chan []BatchMessage, 100)}
			(*subscriptions).Subscriptions = append(
				(*subscriptions).Subscriptions,
				newSubscription,
			)
			go startSubscriber(newSubscription, notifications, copyDatabase(db))
			// subscriptions_notifications <- true // is this still needed?
		}
	}
}

func parser_worker(unparsed_messages <-chan string, claims chan<- []Term, retractions chan<- []Term) {
	event_type_len := 9
	source_len := 4
	for msg := range unparsed_messages {
		start := time.Now()
		zap.L().Debug("PARSE MESSAGE", zap.String("msg", msg))
		event_type := msg[0:event_type_len]
		source := msg[event_type_len:(event_type_len + source_len)]
		val := msg[(event_type_len + source_len):]
		if event_type == "....CLAIM" {
			fact := parse_fact_string(val)
			fact = append([]Term{Term{"id", source}}, fact...)
			claims <- fact
		} else if event_type == "..RETRACT" {
			fact := parse_fact_string(val)
			retractions <- fact
		}
		elapsed := time.Since(start)
		zap.L().Debug("parse time", zap.Duration("duration", elapsed))
	}
}

// func claim_worker(claims <-chan []Term, subscriptions_notifications chan<- bool, db *map[string]Fact) {
// 	for fact_terms := range claims {
// 		start := time.Now()
// 		dbMutex.Lock()
// 		claim(db, Fact{fact_terms})
// 		dbMutex.Unlock()
// 		elapsed := time.Since(start)
// 		zap.L().Debug("claim time", zap.Duration("duration", elapsed))
// 		subscriptions_notifications <- true
// 	}
// }

// func retract_worker(retractions <-chan []Term, subscriptions_notifications chan<- bool, db *map[string]Fact) {
// 	for fact_terms := range retractions {
// 		start := time.Now()
// 		dbMutex.Lock()
// 		retract(db, Fact{fact_terms})
// 		// print_all_facts(*db)
// 		dbMutex.Unlock()
// 		subscriptions_notifications <- true
// 		elapsed := time.Since(start)
// 		zap.L().Debug("retract time", zap.Duration("duration", elapsed))
// 	}
// }

func notify_subscribers_worker(notify_subscribers <-chan bool, subscriber_worker_finished chan<- bool, db *map[string]Fact, notifications chan<- Notification, subscriptions *Subscriptions) {
	// TODO: passing in subscriptions is probably not safe because it can be written in the other goroutine
	latencyMeasurer := makeLatencyMeasurer()
	latencyMeasurer = preLatencyMeasurePart("messageWait", latencyMeasurer)
	for range notify_subscribers {
		latencyMeasurer = postLatencyMeasurePart("messageWait", latencyMeasurer)
		start := time.Now()
		latencyMeasurer = update_all_subscriptions(db, notifications, *subscriptions, latencyMeasurer)
		updateSubscribersTime := time.Since(start)
		zap.L().Debug("notify subscribers time", zap.Duration("duration", updateSubscribersTime))
		subscriber_worker_finished <- true
		latencyMeasurer = updateLatencyMeasurer(latencyMeasurer, 1, "latency - notify subscribers")
		latencyMeasurer = preLatencyMeasurePart("messageWait", latencyMeasurer)
	}
}

func debounce_subscriber_worker(subscriptions_notifications <-chan bool, subscriber_worker_finished <-chan bool, notify_subscribers chan<- bool) {
	claim_waiting := false
	worker_is_free := true
	// TODO: don't use "worker_is_free", but have the notify_subscibers drain the channel?
	go func() {
		for range subscriptions_notifications {
			if worker_is_free {
				worker_is_free = false
				zap.L().Debug("notifying subscriber worker")
				notify_subscribers <- true
			} else {
				zap.L().Debug("debouncing subscription notification becasue worker is busy")
				claim_waiting = true
			}
		}
	}()
	go func() {
		for range subscriber_worker_finished {
			zap.L().Debug("subscriber_worker_finished")
			worker_is_free = true
			if claim_waiting {
				zap.L().Debug("subscriber_worker_finished - running again to catch up to claims")
				claim_waiting = false
				notify_subscribers <- true
			}
		}
	}()
}

func copyDatabase(db *map[string]Fact) map[string]Fact {
	dbCopy := make(map[string]Fact)
	dbMutex.RLock()
	for k, fact := range *db {
		dbCopy[k] = Fact{make([]Term, len(fact.Terms))}
		for i, term := range fact.Terms {
			dbCopy[k].Terms[i] = Term{term.Type, term.Value}
		}
	}
	dbMutex.RUnlock()
	return dbCopy
}

func debug_database_observer(db *map[string]Fact) {
	for {
		dbCopy := copyDatabase(db)
		dbAsSstring := []byte("\033[H\033[2J") // clear terminal output on MacOS
		dbAsBase64Strings := ""
		var keys []string
		for k, _ := range dbCopy {
			keys = append(keys, k)
		}
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
		dbAsBase64Strings += fmt.Sprintf("[[\"id\", \"0\"], [\"text\", \"%s\"]]\n", b64.StdEncoding.EncodeToString([]byte(time.Now().String())))
		err := ioutil.WriteFile("./db_view.txt", dbAsSstring, 0644)
		checkErr(err)
		err2 := ioutil.WriteFile("./db_view_base64.txt", []byte(dbAsBase64Strings), 0644)
		checkErr(err2)
		time.Sleep(1.0 * time.Second)
	}
}

func batch_worker(batch_messages <-chan string, claims chan<- []Term, retractions chan<- []Term, subscriptions_notifications chan<- bool, db *map[string]Fact, subscriptions *Subscriptions) {
	event_type_len := 9
	source_len := 4
	latencyMeasurer := makeLatencyMeasurer()
	latencyMeasurer = preLatencyMeasurePart("messageWait", latencyMeasurer)
	for msg := range batch_messages {
		latencyMeasurer = postLatencyMeasurePart("messageWait", latencyMeasurer)
		// event_type := msg[0:event_type_len]
		// source := msg[event_type_len:(event_type_len + source_len)]
		val := msg[(event_type_len + source_len):]
		var batch_messages []BatchMessage
		err := json.Unmarshal([]byte(val), &batch_messages)
		checkErr(err)
		for _, batch_message := range batch_messages {
			terms := make([]Term, len(batch_message.Fact))
			for j, term := range batch_message.Fact {
				terms[j] = Term{term[0], term[1]}
			}
			if batch_message.Type == "claim" {
				// claims <- terms
				latencyMeasurer = preLatencyMeasurePart("db", latencyMeasurer)
				dbMutex.Lock()
				latencyMeasurer = postLatencyMeasurePart("db", latencyMeasurer)
				latencyMeasurer = preLatencyMeasurePart("action", latencyMeasurer)
				claim(db, Fact{terms})
				latencyMeasurer = postLatencyMeasurePart("action", latencyMeasurer)
				dbMutex.Unlock()
			} else if batch_message.Type == "retract" {
				// retractions <- terms
				latencyMeasurer = preLatencyMeasurePart("db", latencyMeasurer)
				dbMutex.Lock()
				latencyMeasurer = postLatencyMeasurePart("db", latencyMeasurer)
				latencyMeasurer = preLatencyMeasurePart("action", latencyMeasurer)
				retract(db, Fact{terms})
				latencyMeasurer = postLatencyMeasurePart("action", latencyMeasurer)
				dbMutex.Unlock()
			}
		}
		// subscriptions_notifications <- true
		for _, subscription := range (*subscriptions).Subscriptions {
			subscription.batch_messages <- batch_messages
		}
		latencyMeasurer = updateLatencyMeasurer(latencyMeasurer, 1, "latency - batch")
		latencyMeasurer = preLatencyMeasurePart("messageWait", latencyMeasurer)
	}
}

func NewLogger() (*zap.Logger, error) {
	// cfg := zap.NewProductionConfig()
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{
		"/Users/jhaip/Code/lovelace/new-backend/go-server/server.log",
	}
	return cfg.Build()
}

func main() {
	// defer profile.Start().Stop()
	logger, _ := NewLogger() // zap.NewDevelopment() // NewLogger()
	zap.ReplaceGlobals(logger)

	factDatabase := make_fact_database()

	subscriptions := Subscriptions{}

	zap.L().Info("Connecting to ZMQ")
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
	go subscribe_worker(subscription_messages, claims, subscriptions_notifications, &subscriptions, notifications, &factDatabase)
	// go claim_worker(claims, subscriptions_notifications, &factDatabase)
	// go retract_worker(retractions, subscriptions_notifications, &factDatabase)
	go notify_subscribers_worker(notify_subscribers, subscriber_worker_finished, &factDatabase, notifications, &subscriptions)
	go debounce_subscriber_worker(subscriptions_notifications, subscriber_worker_finished, notify_subscribers)
	go notification_worker(notifications, retractions)
	go debug_database_observer(&factDatabase)
	go batch_worker(batch_messages, claims, retractions, subscriptions_notifications, &factDatabase, &subscriptions)

	// go func() {
	// 	for {
	// 		fmt.Println("kick it!")
	// 		subscriptions_notifications <- true
	// 		time.Sleep(1.0 * time.Second)
	// 	}
	// }()

	zap.L().Info("listening...")
	for {
		msg, _ := subscriber.Recv(0)
		event_type := msg[0:event_type_len]
		source := msg[event_type_len:(event_type_len + source_len)]
		val := msg[(event_type_len + source_len):]
		if event_type == ".....PING" {
			zap.L().Debug("got PING", zap.String("source", source), zap.String("value", val))
			notifications <- Notification{source, val, ""}
		} else if event_type == "....CLAIM" {
			unparsed_messages <- msg
		} else if event_type == "..RETRACT" {
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
