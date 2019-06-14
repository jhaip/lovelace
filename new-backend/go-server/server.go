package main

import (
	"encoding/json"
	"fmt"
	// "log"
	"os"
	"io"
	"runtime"
	"strconv"
	"sync"
	"time"

	// "net/http"
	_ "net/http/pprof"

	_ "github.com/mattn/go-sqlite3"
	zmq "github.com/pebbe/zmq4"
	"go.uber.org/zap"
	// "runtime/trace"

	opentracing "github.com/opentracing/opentracing-go"
	jaeger "github.com/uber/jaeger-client-go"
	config "github.com/uber/jaeger-client-go/config"
)

var dbMutex sync.RWMutex
var subscriberMutex sync.RWMutex
var zmqClient sync.Mutex

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
	dead           *sync.WaitGroup
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
	// publisher, err := zmq.NewSocket(zmq.PUB)
	// checkErr(err)
	// defer publisher.Close()
	// publisherBindErr := publisher.Bind("tcp://*:5555")
	// checkErr(publisherBindErr)
	cache := make(map[string]string)
	for notification := range notifications {
		// start := time.Now()
		msg := fmt.Sprintf("%s%s%s", notification.Source, notification.Id, notification.Result)
		cache_key := fmt.Sprintf("%s%s", notification.Source, notification.Id)
		cache_value, cache_hit := cache[cache_key]
		if cache_hit == false || cache_value != msg {
			cache[cache_key] = msg
			msgWithTime := fmt.Sprintf("%s%s%v%s", notification.Source, notification.Id, makeTimestampMillis(), notification.Result)
			zmqClient.Lock()
			// _, sendErr := publisher.Send(msgWithTime, zmq.DONTWAIT)
			// checkErr(sendErr)
			client.SendMessage(notification.Source, msgWithTime)
			zmqClient.Unlock()
		} else {
			zap.L().Error(fmt.Sprintf("SKIPPING MESSAGE %s", notification.Source))
		}
		// timeToSendResults := time.Since(start)
		// zap.L().Debug("send notification", zap.Duration("timeToSendResults", timeToSendResults))
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
				subscription_fact := append([]Term{Term{"text", "subscription"}, Term{"id", source}, Term{"text", subscription_data.Id}, Term{"integer", strconv.Itoa(i)}}, fact...)
				dbMutex.Lock()
				claim(db, Fact{subscription_fact})
				dbMutex.Unlock()
				// prepare a batch message for the new subscription fact
				batch_message_facts := make([][]string, len(subscription_fact))
				for k, subscription_fact_term := range subscription_fact {
					batch_message_facts[k] = []string{subscription_fact_term.Type, subscription_fact_term.Value}
				}
				batch_messages[i] = BatchMessage{"claim", batch_message_facts}
			}
			newSubscription := Subscription{source, subscription_data.Id, query, make(chan []BatchMessage, 1000), &sync.WaitGroup{}}
			newSubscription.dead.Add(1)
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

func on_source_death(dying_source string, db *map[string]Fact, subscriptions *Subscriptions) {
	zap.L().Info("SOURCE DEATH - recv", zap.String("source", dying_source))
	// Retract all facts by source and facts about the source's subscriptions
	dbMutex.Lock()
	retract(db, Fact{[]Term{Term{"id", dying_source}, Term{"postfix", ""}}})
	retract(db, Fact{[]Term{Term{"text", "subscription"}, Term{"id", dying_source}, Term{"postfix", ""}}})
	dbMutex.Unlock()
	subscriberMutex.Lock()
	newSubscriptions := make([]Subscription, 0)
	for _, subscription := range (*subscriptions).Subscriptions {
		if subscription.Source != dying_source {
			newSubscriptions = append(newSubscriptions, subscription)
			batch_messages := []BatchMessage{
				BatchMessage{"retract", [][]string{[]string{"id", dying_source}, []string{"postfix", ""}}},
				BatchMessage{"retract", [][]string{[]string{"text", "subscription"}, []string{"id", dying_source}, []string{"postfix", ""}}},
			}
			subscription.batch_messages <- batch_messages
		} else {
			zap.L().Info("SOURCE DEATH - closing channel", zap.String("source", dying_source))
			close(subscription.batch_messages)
			zap.L().Info("SOURCE DEATH - waiting for death signal", zap.String("source", dying_source))
			subscription.dead.Wait()
			zap.L().Info("SOURCE DEATH - confirmed dead", zap.String("source", dying_source))
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
				retract(db, Fact{terms})
				dbMutex.Unlock()
			} else if batch_message.Type == "death" {
				// Assume Fact = [["id", "0004"]]
				dying_source := batch_message.Fact[0][1]
				// This a blocking call that does a couple retracts and waits for a goroutine to die
				// There is a potential for slowdown or blocking the whole server if a subscriber won't die
				on_source_death(dying_source, db, subscriptions)
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

func main() {
	// tracer, closer := initJaeger("room-service")
	// defer closer.Close()
	// opentracing.SetGlobalTracer(tracer)
	// f, err := os.Create("trace.out")
	// if err != nil {
	// 	panic(err)
	// }
	// defer f.Close()

	// err = trace.Start(f)
	// if err != nil {
	// 	panic(err)
	// }
	// defer trace.Stop()

	// programStartTime := time.Now()

	// defer profile.Start().Stop()
	logger, loggerCreateError := NewLogger() // zap.NewDevelopment() // NewLogger()
	checkErr(loggerCreateError)
	zap.ReplaceGlobals(logger)

	runtime.SetMutexProfileFraction(5)

	// go func() {
	// 	log.Println(http.ListenAndServe("localhost:6060", nil))
	// }()

	factDatabase := make_fact_database()

	subscriptions := Subscriptions{}

	zap.L().Info("Connecting to ZMQ")
	// subscriber, zmqSubscribeError := zmq.NewSocket(zmq.SUB)
	// checkErr(zmqSubscribeError)
	// subscriber.SetRcvhwm(100000)
	// defer subscriber.Close()
	
	// var subSetFilterErr error
	// subSetFilterErr = subscriber.SetSubscribe(".....PING")
	// checkErr(subSetFilterErr)
	// subSetFilterErr = subscriber.SetSubscribe("....BATCH")
	// checkErr(subSetFilterErr)
	// subSetFilterErr = subscriber.SetSubscribe("SUBSCRIBE")
	// checkErr(subSetFilterErr)

	// subBindErr := subscriber.Bind("tcp://*:5556")
	// checkErr(subBindErr)
	client, zmqCreationErr := zmq.NewSocket(zmq.ROUTER)
	checkErr(zmqCreationErr)
	defer client.Close()
    client.Bind("tcp://*:5570")

	event_type_len := 9
	source_len := 4

	subscription_messages := make(chan string, 1000)
	subscriptions_notifications := make(chan bool, 1000)
	notifications := make(chan Notification, 1000)
	batch_messages := make(chan string, 1000)

	go subscribe_worker(subscription_messages, subscriptions_notifications, &subscriptions, notifications, &factDatabase)
	go notification_worker(notifications, client)
	go batch_worker(batch_messages, subscriptions_notifications, &factDatabase, &subscriptions)

	// zap.L().Info("pre-recv")
	// subscriber.Recv(0)
	// zap.L().Info("post-recv")

	// time.Sleep(time.Duration(2) * time.Second)
	zap.L().Info("listening...")
	// rootSpan := tracer.StartSpan("run-test")
	// mapc := opentracing.TextMapCarrier(make(map[string]string))
	// err := tracer.Inject(rootSpan.Context(), opentracing.TextMap, mapc)
	// checkErr(err)
	// zap.L().Info(mapc["uber-trace-id"])
	// for k, v := range mapc { 
	// 	zap.L().Info(k)
	// 	zap.L().Info(v)
	// }
	
	go func() {
		time.Sleep(time.Duration(40) * time.Second)
		// rootSpan.Finish()
		// closer.Close()
		panic("time elapsed - ending");
	}()
	for {
		zap.L().Info("pre-recv")  // TODO: remove
		// msg, recvErr := subscriber.Recv(0)
		zmqClient.Lock()
		rawMsg, recvErr := client.RecvMessage(0)
		// zap.L().Info("post-recv")
		checkErr(recvErr)
		zap.L().Info(fmt.Sprintf("%s %s", rawMsg[0], rawMsg[1])) // TODO: remove
		rawMsgId := rawMsg[0]
		msg := rawMsg[1]
		zmqClient.Unlock()
		// span := rootSpan.Tracer().StartSpan(
		// 	"zmq-recv-loop",
		// 	opentracing.ChildOf(rootSpan.Context()),
		// )
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
		// sleepSpan := span.Tracer().StartSpan(
		// 	"zmq-recv-loop-sleep",
		// 	opentracing.ChildOf(span.Context()),
		// )
		time.Sleep(time.Duration(1) * time.Microsecond)
		// sleepSpan.Finish()
		// span.Finish()
		// delta := time.Now().Sub(programStartTime)
		// if (delta.Seconds() > 20) {
		// 	zap.L().Debug("time elapsed -- ending")
		// 	break;
		// } else {
		// 	zap.L().Debug("not done yet")
		// }
	}
}
