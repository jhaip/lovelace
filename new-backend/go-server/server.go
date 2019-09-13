package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	// "log"
	"math"
	"os"
	"io"
	"io/ioutil"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"

	// "net/http"
	_ "net/http/pprof"

	_ "github.com/mattn/go-sqlite3"
	zmq "github.com/pebbe/zmq4"
	"go.uber.org/zap"
	// "runtime/trace"

	b64 "encoding/base64"

	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"
)

var dbMutex sync.RWMutex
var subscriberMutex sync.RWMutex
var zmqClient sync.Mutex

type Term struct {
	Type  string
	Value []byte
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
	dead           *sync.WaitGroup
	warmed         *sync.WaitGroup
}

type Subscriptions struct {
	Subscriptions []Subscription
}

type Notification struct {
	Source string
	Id     string
	Result string
}

type BatchMessageJSON struct {
	Type string     `json:"type"`
	Fact [][]string `json:"fact"`
}

type BatchMessage struct {
	Type string
	Fact []Term
}

type RoomUpdate struct {
	Type uint8
	Source string
	SubscriptionId string
	Facts [][]Term
}


// initJaeger returns an instance of Jaeger Tracer that samples 100% of traces and logs all spans to stdout.
func initJaeger(service string) (opentracing.Tracer, io.Closer) {
    cfg := &config.Configuration{
        Sampler: &config.SamplerConfig{
            Type:  "const",
            Param: 1,
        },
        Reporter: &config.ReporterConfig{
            LogSpans: false,
        },
    }
    tracer, closer, err := cfg.New(service, config.Logger(jaeger.StdLogger))
    if err != nil {
        panic(fmt.Sprintf("ERROR: cannot init Jaeger: %v\n", err))
    }
    return tracer, closer
}

func checkErr(err error) {
	if err != nil {
		zap.L().Fatal("FATAL ERROR", zap.Error(err))
		panic(err)
	}
}

func makeTimestampMillis() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

func notification_worker(notifications <-chan Notification, client *zmq.Socket) {
	cache := make(map[string]string)
	for notification := range notifications {
		msg := fmt.Sprintf("%s%s%s", notification.Source, notification.Id, notification.Result)
		cache_key := fmt.Sprintf("%s%s", notification.Source, notification.Id)
		cache_value, cache_hit := cache[cache_key]
		if cache_hit == false || cache_value != msg {
			cache[cache_key] = msg
			msgWithTime := fmt.Sprintf("%s%s%v%s", notification.Source, notification.Id, makeTimestampMillis(), notification.Result)
			zmqClient.Lock()
			_, sendErr := client.SendMessage(notification.Source, msgWithTime)
			checkErr(sendErr)
			zmqClient.Unlock()
		}
	}
}

func intToBinary(x int) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(x))
	return b
}

func floatToBinary(x float32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, math.Float32bits(x))
	return b
}

func subscribe_worker(subscription_messages <-chan string,
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
			err := json.Unmarshal([]byte(val), &subscription_data)
			checkErr(err)
			query := make([][]Term, 0)
			batch_messages := make([]BatchMessage, len(subscription_data.Facts))
			for i, fact_string := range subscription_data.Facts {
				fact := parse_fact_string(fact_string)
				query = append(query, fact)
				subscription_fact := append([]Term{
					Term{"text", []byte("subscription")},
					Term{"id", []byte(source)},
					Term{"text", []byte(subscription_data.Id)},
					Term{"integer", intToBinary(i)},
				}, fact...)
				dbMutex.Lock()
				claim(db, Fact{subscription_fact})
				dbMutex.Unlock()
				// prepare a batch message for the new subscription fact
				batch_message_facts := make([]Term, len(subscription_fact))
				for k, subscription_fact_term := range subscription_fact {
					batch_message_facts[k] = Term{subscription_fact_term.Type, subscription_fact_term.Value}
				}
				batch_messages[i] = BatchMessage{"claim", batch_message_facts}
			}
			newSubscription := Subscription{source, subscription_data.Id, query, make(chan []BatchMessage, 1000), &sync.WaitGroup{}, &sync.WaitGroup{}}
			newSubscription.dead.Add(1)
			newSubscription.warmed.Add(1)
			subscriberMutex.Lock()
			(*subscriptions).Subscriptions = append(
				(*subscriptions).Subscriptions,
				newSubscription,
			)
			for _, subscription := range (*subscriptions).Subscriptions {
				subscription.batch_messages <- batch_messages
			}
			subscriberMutex.Unlock()
			go startSubscriber(newSubscription, notifications, copyDatabase(db))
			// subscriptions_notifications <- true // is this still needed?
		}
	}
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

func on_source_death(dying_source string, db *map[string]Fact, subscriptions *Subscriptions) {
	zap.L().Info("SOURCE DEATH - recv", zap.String("source", dying_source))
	// Retract all facts by source and facts about the source's subscriptions
	dbMutex.Lock()
	retract(db, Fact{[]Term{Term{"id", []byte(dying_source)}, Term{"postfix", []byte("")}}})
	retract(db, Fact{[]Term{Term{"text", []byte("subscription")}, Term{"id", []byte(dying_source)}, Term{"postfix", []byte("")}}})
	dbMutex.Unlock()
	subscriberMutex.Lock()
	newSubscriptions := make([]Subscription, 0)
	for _, subscription := range (*subscriptions).Subscriptions {
		if subscription.Source != dying_source {
			newSubscriptions = append(newSubscriptions, subscription)
			batch_messages := []BatchMessage{
				BatchMessage{"retract", []Term{Term{"id", []byte(dying_source)}, Term{"postfix", []byte("")}}},
				BatchMessage{"retract", []Term{Term{"text", []byte("subscription")}, Term{"id", []byte(dying_source)}, Term{"postfix", []byte("")}}},
			}
			subscription.batch_messages <- batch_messages
		} else {
			zap.L().Info("SOURCE DEATH - closing channel", zap.String("source", dying_source))
			waitStart := time.Now()
			// Wait for subscriber to stop sending cache warming messages
			// to itself to avoid error sending on a closed channel.
			subscription.warmed.Wait()
			close(subscription.batch_messages)
			zap.L().Info("SOURCE DEATH - waiting for death signal", zap.String("source", dying_source))
			subscription.dead.Wait()
			waitTimeElapsed := time.Since(waitStart)
			zap.L().Info("SOURCE DEATH - confirmed dead", zap.String("source", dying_source), zap.Duration("timeToClose", waitTimeElapsed))
			// SOMETHING BAD COULD HAPPEN IF A MESSAGE WAS RECEIVED AND SOMEONE TRIED TO
			// ADD A MESSAGE TO THE SUBSCRIPTIONS QUEUE
		}
	}
	(*subscriptions).Subscriptions = newSubscriptions
	subscriberMutex.Unlock()
}

func on_subscription_death(source string, subscriptionId string, db *map[string]Fact, subscriptions *Subscriptions) {
	zap.L().Info("SUBSCRIPTION DEATH - recv", zap.String("source", source), zap.String("subscriptionId", subscriptionId))
	dbMutex.Lock()
	retract(db, Fact{[]Term{
		Term{"text", []byte("subscription")},
		Term{"id", []byte(source)},
		Term{"text", []byte(subscriptionId)},
		Term{"postfix", []byte("")},
	}})
	dbMutex.Unlock()
	subscriberMutex.Lock()
	newSubscriptions := make([]Subscription, 0)
	for _, subscription := range (*subscriptions).Subscriptions {
		if subscription.Id != subscriptionId {
			newSubscriptions = append(newSubscriptions, subscription)
			batch_messages := []BatchMessage{
				BatchMessage{"retract", []Term{
					Term{"text", []byte("subscription")},
					Term{"id", []byte(source)},
					Term{"text", []byte(subscriptionId)},
					Term{"postfix", []byte("")},
				}},
			}
			subscription.batch_messages <- batch_messages
		} else {
			zap.L().Info("SUBSCRIPTION DEATH - closing channel", zap.String("source", source), zap.String("subscriptionId", subscriptionId))
			waitStart := time.Now()
			// Wait for subscriber to stop sending cache warming messages
			// to itself to avoid error sending on a closed channel.
			subscription.warmed.Wait()
			close(subscription.batch_messages)
			zap.L().Info("SUBSCRIPTION DEATH - waiting for death signal", zap.String("source", source), zap.String("subscriptionId", subscriptionId))
			subscription.dead.Wait()
			waitTimeElapsed := time.Since(waitStart)
			zap.L().Info("SUBSCRIPTION DEATH - confirmed dead", zap.String("source", source), zap.String("subscriptionId", subscriptionId), zap.Duration("timeToClose", waitTimeElapsed))
			// SOMETHING BAD COULD HAPPEN IF A MESSAGE WAS RECEIVED AND SOMEONE TRIED TO
			// ADD A MESSAGE TO THE SUBSCRIPTIONS QUEUE
		}
	}
	(*subscriptions).Subscriptions = newSubscriptions
	subscriberMutex.Unlock()
}

func batch_worker(batch_messages <-chan string, subscriptions_notifications chan<- bool, db *map[string]Fact, subscriptions *Subscriptions) {
	event_type_len := 9
	source_len := 4
	for msg := range batch_messages {
		// event_type := msg[0:event_type_len]
		// source := msg[event_type_len:(event_type_len + source_len)]
		val := msg[(event_type_len + source_len):]
		var batch_messages []BatchMessage
		err := json.Unmarshal([]byte(val), &batch_messages)
		if err != nil {
			zap.L().Info("BATCH MESSAGE BODY:")
			zap.L().Info(val)
		}
		checkErr(err)
		for _, batch_message := range batch_messages {
			terms := batch_message.Fact
			if batch_message.Type == "claim" {
				// claims <- terms
				dbMutex.Lock()
				claim(db, Fact{terms})
				dbMutex.Unlock()
			} else if batch_message.Type == "retract" {
				// retractions <- terms
				dbMutex.Lock()
				retract(db, Fact{terms})
				dbMutex.Unlock()
			} else if batch_message.Type == "death" {
				// Assume Fact = [["id", "0004"]]
				dying_source := string(batch_message.Fact[0].Value[:])
				// This a blocking call that does a couple retracts and waits for a goroutine to die
				// There is a potential for slowdown or blocking the whole server if a subscriber won't die
				on_source_death(dying_source, db, subscriptions)
			} else if batch_message.Type == "subscriptiondeath" {
				// Assume Fact = [["id", "0004"], ["text", ..subscription id..]]
				source := string(batch_message.Fact[0].Value[:])
				dying_subscription_id := string(batch_message.Fact[1].Value[:])
				// This a blocking call that does a couple retracts and waits for a goroutine to die
				// There is a potential for slowdown or blocking the whole server if a subscriber won't die
				on_subscription_death(source, dying_subscription_id, db, subscriptions)
			}
		}
		// subscriptions_notifications <- true
		subscriberMutex.RLock()
		for _, subscription := range (*subscriptions).Subscriptions {
			subscription.batch_messages <- batch_messages
		}
		subscriberMutex.RUnlock()
	}
}

func GetBasePath() string {
	envBasePath := os.Getenv("DYNAMIC_ROOT")
	if envBasePath != "" {
		return envBasePath + "/"
	}
	env := "HOME"
	if runtime.GOOS == "windows" {
		env = "USERPROFILE"
	} else if runtime.GOOS == "plan9" {
		env = "home"
	}
	return os.Getenv(env) + "/lovelace/"
}

func NewLogger() (*zap.Logger, error) {
	// cfg := zap.NewProductionConfig()
	cfg := zap.NewDevelopmentConfig()
	cfg.OutputPaths = []string{
		GetBasePath() + "new-backend/go-server/server.log",
	}
	return cfg.Build()
}

func parse_room_update(source string, msg string) []RoomUpdate {
	event_type_len := 9
	source_len := 4
	event_type := msg[0:event_type_len]
	val := msg[(event_type_len + source_len):]
	if event_type == ".....PING" {
		return []RoomUpdate{{0, source, val, make([][]Term, 0)}}
	} else if event_type == "SUBSCRIBE" {
		subscription_data := SubscriptionData{}
		err := json.Unmarshal([]byte(val), &subscription_data)
		checkErr(err)
		query := make([][]Term, 0)
		for _, fact_string := range subscription_data.Facts {
			fact := parse_fact_string(fact_string)
			query = append(query, fact)
		}
		return []RoomUpdate{{3, source, subscription_data.Id, query}}
	} else if event_type == "....BATCH" {
		var batch_messages []BatchMessageJSON
		err := json.Unmarshal([]byte(val), &batch_messages)
		checkErr(err)
		updates := make([]RoomUpdate, len(batch_messages))
		for i, batch_message := range batch_messages {
			terms := make([]Term, len(batch_message.Fact))
			for j, term := range batch_message.Fact {
				term_type := term[0]
				if term_type == "integer" {
					intTermValue, err := strconv.Atoi(term[1])
					checkErr(err)
					terms[j] = Term{term_type, intToBinary(intTermValue)}
				} else if term_type == "float" {
					floatTermValue64, err := strconv.ParseFloat(term[1], 32)
					floatTermValue32 := float32(floatTermValue64)
					checkErr(err)
					terms[j] = Term{term_type, floatToBinary(floatTermValue32)}
				} else {
					terms[j] = Term{term_type, []byte(term[1])}
				}
			}
			if batch_message.Type == "claim" {
				updates[i] = RoomUpdate{1, source, "", [][]Term{terms}}
			} else if batch_message.Type == "retract" {
				updates[i] = RoomUpdate{2, source, "", [][]Term{terms}}
			} else if batch_message.Type == "death" {
				// Assume Fact = [["id", "0004"]]
				dying_source := batch_message.Fact[0][1]
				facts := [][]Term{[]Term{Term{"id", []byte(dying_source)}}}
				updates[i] = RoomUpdate{4, source, "", facts}
			} else if batch_message.Type == "subscriptiondeath" {
				// Assume Fact = [["id", "0004"], ["text", ..subscription id..]]
				source := batch_message.Fact[0][1]
				dying_subscription_id := batch_message.Fact[1][1]
				facts := [][]Term{[]Term{
					Term{"id", []byte(source)},
					Term{"text", []byte(dying_subscription_id)},
				}}
				updates[i] = RoomUpdate{4, source, "", facts}
			}
		}
		return updates
	}
	return make([]RoomUpdate, 0)
}

func main() {
	// tracer, closer := initJaeger("room-service")
	// defer closer.Close()
	// opentracing.SetGlobalTracer(tracer)

	logger, loggerCreateError := NewLogger() // zap.NewDevelopment()
	checkErr(loggerCreateError)
	zap.ReplaceGlobals(logger)

	runtime.SetMutexProfileFraction(5)

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	factDatabase := make_fact_database()

	subscriptions := Subscriptions{}
	
	client, zmqCreationErr := zmq.NewSocket(zmq.ROUTER)
	checkErr(zmqCreationErr)
	defer client.Close()
	client.Bind("tcp://*:5570")
	zap.L().Info("Connecting to ZMQ")

	event_type_len := 9
	source_len := 4

	subscription_messages := make(chan string, 1000)
	subscriptions_notifications := make(chan bool, 1000)
	notifications := make(chan Notification, 1000)
	batch_messages := make(chan string, 1000)

	go subscribe_worker(subscription_messages, subscriptions_notifications, &subscriptions, notifications, &factDatabase)
	go notification_worker(notifications, client)
	go debug_database_observer(&factDatabase)
	go batch_worker(batch_messages, subscriptions_notifications, &factDatabase, &subscriptions)

	zap.L().Info("listening...")
	// rootSpan := tracer.StartSpan("run-test")
	// mapc := opentracing.TextMapCarrier(make(map[string]string))
	// err := tracer.Inject(rootSpan.Context(), opentracing.TextMap, mapc)
	// checkErr(err)
	// zap.L().Info(mapc["uber-trace-id"])
	
	// go func() {
	// 	time.Sleep(time.Duration(40) * time.Second)
	// 	// rootSpan.Finish()
	// 	// closer.Close()
	// 	panic("time elapsed - ending");
	// }()
	for {
		zmqClient.Lock()
		rawMsg, recvErr := client.RecvMessage(zmq.DONTWAIT)
		if recvErr != nil {
			zmqClient.Unlock()
			time.Sleep(time.Duration(1) * time.Millisecond)
			continue;
		}
		rawMsgId := rawMsg[0]
		msg := rawMsg[1]
		zmqClient.Unlock()
		// span := rootSpan.Tracer().StartSpan(
		// 	"zmq-recv-loop",
		// 	opentracing.ChildOf(rootSpan.Context()),
		// )

		// updates := parse_room_update(rawMsgId, msg)
		// for _, update := range updates {
		// 	if update.Type == 0 {
		// 		zap.L().Debug("got PING", zap.String("source", update.Source), zap.String("value", update.SubscriptionId))
		// 		// notifications <- Notification{source, val, mapc["uber-trace-id"]}
		// 		notifications <- Notification{update.Source, update.SubscriptionId, ""}
		// 	} else if update.Type == 3 {
		// 		subscription_messages <- update
		// 	} else {
		// 		batch_messages <- update
		// 	}
		// }

		event_type := msg[0:event_type_len]
		// source := msg[event_type_len:(event_type_len + source_len)]
		source := rawMsgId
		val := msg[(event_type_len + source_len):]
		if event_type == ".....PING" {
			zap.L().Debug("got PING", zap.String("source", source), zap.String("value", val))
			// notifications <- Notification{source, val, mapc["uber-trace-id"]}
			notifications <- Notification{source, val, ""}
		} else if event_type == "SUBSCRIBE" {
			subscription_messages <- msg
		} else if event_type == "....BATCH" {
			batch_messages <- msg
		}
		time.Sleep(time.Duration(1) * time.Microsecond)
		// span.Finish()
	}
}
