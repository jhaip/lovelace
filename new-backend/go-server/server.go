package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	// "log"
	// "math"
	"os"
	"io"
	"io/ioutil"
	"runtime"
	"sort"
	"strconv"
	"sync"
	"time"
	"bytes"

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

	flatbuffers "github.com/google/flatbuffers/go"
	roomupdate "room/roomupdatefbs"
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
	Result []QueryResult
}

type BatchMessageJSON struct {
	Type string     `json:"type"`
	Fact [][]string `json:"fact"`
}

type BatchMessage struct {
	Type string
	Fact []Term
}

type RoomUpdateType uint8
const (
	PING RoomUpdateType = 0
	CLAIM RoomUpdateType = 1
    RETRACT RoomUpdateType = 2
    SUBSCRIBE RoomUpdateType = 3
    DEATH RoomUpdateType = 4
    SUBSCRIPTION_DEATH RoomUpdateType = 5
)

type RoomUpdate struct {
	Type RoomUpdateType
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

func testRoomUpdateSerialization() []byte {
	builder := flatbuffers.NewBuilder(1024)

	factValue := builder.CreateByteVector([]byte("Hello World!"))

	roomupdate.FactStart(builder)
	roomupdate.FactAddType(builder, roomupdate.FactTypeText)
	roomupdate.FactAddValue(builder, factValue)
	fact := roomupdate.FactEnd(builder)

	updateSource := builder.CreateString("1234")
	updateSubId := builder.CreateString("asdf")
	roomupdate.RoomUpdateStartFactsVector(builder, 1)
	builder.PrependUOffsetT(fact)
	facts := builder.EndVector(1)  // 1 == 1 from roomupdate.RoomUpdateStartFactsVector(builder, 1)

	roomupdate.RoomUpdateStart(builder)
	roomupdate.RoomUpdateAddType(builder, roomupdate.UpdateTypeClaim)
	roomupdate.RoomUpdateAddSource(builder, updateSource)
	roomupdate.RoomUpdateAddSubscriptionId(builder, updateSubId)	
	roomupdate.RoomUpdateAddFacts(builder, facts)
	update := roomupdate.RoomUpdateEnd(builder)

	roomupdate.RoomUpdatesStartUpdatesVector(builder, 1)
	builder.PrependUOffsetT(update)
	updates := builder.EndVector(1)

	roomupdate.RoomUpdatesStart(builder)
	roomupdate.RoomUpdatesAddUpdates(builder, updates)
	full_updates_msg := roomupdate.RoomUpdatesEnd(builder)
	
	builder.Finish(full_updates_msg)
	msg_buf := builder.FinishedBytes()
	return msg_buf
}

func testRoomUpdateDeserialization(data []byte) {
	room_updates_obj := roomupdate.GetRootAsRoomUpdates(data, 0)
	updates_length := room_updates_obj.UpdatesLength()
	fmt.Println(updates_length)
	update := new(roomupdate.RoomUpdate)
	room_updates_obj.Updates(update, 0)
	fmt.Println(update.Type())
	source := string(update.Source())
	sub_id := string(update.SubscriptionId())
	fmt.Println(source)
	fmt.Println(sub_id)
	facts_length := update.FactsLength()
	fmt.Println(facts_length)
	fact := new(roomupdate.Fact)
	update.Facts(fact, 0)
	fmt.Println(fact.Type())
	fmt.Println(string(fact.ValueBytes()))
}

func testRoomUpdateSerializationResult() []byte {
	builder := flatbuffers.NewBuilder(1024)

	resultVariableName := builder.CreateString("X")
	resultValue := builder.CreateByteVector([]byte("blue"))

	roomupdate.RoomResultStart(builder)
	roomupdate.RoomResultAddVariableName(builder, resultVariableName)
	roomupdate.RoomResultAddType(builder, roomupdate.FactTypeText)
	roomupdate.RoomResultAddValue(builder, resultValue)
	room_result := roomupdate.FactEnd(builder)

	roomupdate.ResultSetStartResultsVector(builder, 1)
	builder.PrependUOffsetT(room_result)
	results := builder.EndVector(1)

	roomupdate.ResultSetStart(builder)
	roomupdate.ResultSetAddResults(builder, results)
	result_set := roomupdate.ResultSetEnd(builder)

	roomResponseSource := builder.CreateString("1234")
	roomResponseSubscriptionId := builder.CreateString("asdf")

	roomupdate.RoomResponseStartResultSetsVector(builder, 1)
	builder.PrependUOffsetT(result_set)
	result_sets := builder.EndVector(1)

	roomupdate.RoomResponseStart(builder)
	roomupdate.RoomResponseAddSource(builder, roomResponseSource)
	roomupdate.RoomResponseAddSubscriptionId(builder, roomResponseSubscriptionId)
	roomupdate.RoomResponseAddResultSets(builder, result_sets)
	full_response_msg := roomupdate.RoomResponseEnd(builder)

	builder.Finish(full_response_msg)
	msg_buf := builder.FinishedBytes()
	return msg_buf
}

func testRoomUpdateDeserializationResult(data []byte) {
	room_response_obj := roomupdate.GetRootAsRoomResponse(data, 0)
	fmt.Println(string(room_response_obj.Source()))
	fmt.Println(string(room_response_obj.SubscriptionId()))
	result_sets_length := room_response_obj.ResultSetsLength()
	fmt.Println(result_sets_length)
	result_set := new(roomupdate.ResultSet)
	room_response_obj.ResultSets(result_set, 0)
	results_length := result_set.ResultsLength()
	fmt.Println(results_length)
	result := new(roomupdate.RoomResult)
	result_set.Results(result, 0)
	fmt.Println(string(result.VariableName()))
	fmt.Println(result.Type())
	fmt.Println(string(result.ValueBytes()))
}

func serialize_query_result(notification Notification, timestamp int64) ([]byte, []byte) {
	termTypeToRoomResultTypeEnum := map[string]roomupdate.FactType {
		"id": roomupdate.FactTypeId,
		"text": roomupdate.FactTypeText,
		"integer": roomupdate.FactTypeInteger,
		"float": roomupdate.FactTypeFloat,
		"binary": roomupdate.FactTypeBinary,
	}

	builder := flatbuffers.NewBuilder(1024)

	n_results := len(notification.Result)
	result_sets_tmp := make([]flatbuffers.UOffsetT, n_results)

	for i := 0; i < n_results; i += 1 {
		n_room_results := len(notification.Result[i].Result)
		room_results_tmp := make([]flatbuffers.UOffsetT, n_room_results)

		z := 0
		for variableName, term := range notification.Result[i].Result {
			resultVariableName := builder.CreateString(variableName)
			resultValue := builder.CreateByteVector(term.Value)
			resultValueType := termTypeToRoomResultTypeEnum[term.Type]

			roomupdate.RoomResultStart(builder)
			roomupdate.RoomResultAddVariableName(builder, resultVariableName)
			roomupdate.RoomResultAddType(builder, resultValueType)
			roomupdate.RoomResultAddValue(builder, resultValue)
			room_results_tmp[z] = roomupdate.FactEnd(builder)
			z += 1
		}

		roomupdate.ResultSetStartResultsVector(builder, n_room_results)
		for k := 0; k < n_room_results; k += 1 {
			builder.PrependUOffsetT(room_results_tmp[k])
		}
		results := builder.EndVector(n_room_results)

		roomupdate.ResultSetStart(builder)
		roomupdate.ResultSetAddResults(builder, results)
		result_sets_tmp[i] = roomupdate.ResultSetEnd(builder)
	}
	
	roomupdate.RoomResponseStartResultSetsVector(builder, n_results)
	for i := 0; i < n_results; i += 1 {
		builder.PrependUOffsetT(result_sets_tmp[i])
	}
	result_sets := builder.EndVector(n_results)

	roomResponseSource := builder.CreateString(notification.Source)
	roomResponseSubscriptionId := builder.CreateString(notification.Id)

	roomupdate.RoomResponseStart(builder)
	roomupdate.RoomResponseAddSource(builder, roomResponseSource)
	roomupdate.RoomResponseAddSubscriptionId(builder, roomResponseSubscriptionId)
	roomupdate.RoomResponseAddResultSets(builder, result_sets)
	full_response_msg := roomupdate.RoomResponseEnd(builder)

	builder.Finish(full_response_msg)
	msg_buf := builder.FinishedBytes()

	// for now, don't return a version with a timestamp
	return msg_buf, msg_buf
}

func marshal_query_result(query_results []QueryResult) string {
	encoded_results := make([]map[string][]string, 0)
	for _, query_result := range query_results {
		encoded_result := make(map[string][]string)
		for variable_name, term := range query_result.Result {
			// TODO: eventually support encoding at binary here
			if term.Type == "integer" {
				intValue := int(int32(binary.LittleEndian.Uint32(term.Value)))
				encoded_result[variable_name] = []string{term.Type, strconv.Itoa(intValue)}
			} else if term.Type == "float" {
				floatValue := float64(float32(binary.LittleEndian.Uint32(term.Value)))
				encoded_result[variable_name] = []string{term.Type, strconv.FormatFloat(floatValue, 'f', -1, 32)}
			} else {
				encoded_result[variable_name] = []string{term.Type, string(term.Value[:])}
			}
		}
		encoded_results = append(encoded_results, encoded_result)
	}
	marshalled_results, err := json.Marshal(encoded_results)
	checkErr(err)
	return string(marshalled_results)
}

func notification_worker(notifications <-chan Notification, client *zmq.Socket) {
	cache := make(map[string][]byte)
	for notification := range notifications {
		msg, msgWithTimestamp := serialize_query_result(notification, makeTimestampMillis())
		cache_key := fmt.Sprintf("%s%s", notification.Source, notification.Id)
		cache_value, cache_hit := cache[cache_key]
		if cache_hit == false || !bytes.Equal(cache_value, msg) {
			cache[cache_key] = msg
			zmqClient.Lock()
			_, sendErr := client.SendMessage(notification.Source, msgWithTimestamp)
			checkErr(sendErr)
			zmqClient.Unlock()
		}
	}
}

func intToBinary(x int) []byte {
	// b := make([]byte, 4)
	// binary.LittleEndian.PutUint32(b, uint32(x))
	// return b
	str := strconv.Itoa(x)
	return []byte(str)
}

func floatToBinary(x float32) []byte {
	// b := make([]byte, 4)
	// binary.LittleEndian.PutUint32(b, math.Float32bits(x))
	// return b
	str := strconv.FormatFloat(float64(x), 'f', -1, 32)
	return []byte(str)
}

func subscribe_worker(subscription_messages <-chan RoomUpdate,
	subscriptions_notifications chan<- bool,
	subscriptions *Subscriptions,
	notifications chan<- Notification,
	db *map[string]Fact) {

	for room_update := range subscription_messages {
		batch_messages := make([]BatchMessage, len(room_update.Facts))
		for i, fact := range room_update.Facts {
			subscription_fact := append([]Term{
				Term{"text", []byte("subscription")},
				Term{"id", []byte(room_update.Source)},
				Term{"text", []byte(room_update.SubscriptionId)},
				Term{"integer", intToBinary(i)},
			}, fact...)
			dbMutex.Lock()
			claim(db, Fact{subscription_fact})
			dbMutex.Unlock()
			// prepare a batch message for the new subscription fact
			batch_messages[i] = BatchMessage{"claim", subscription_fact}
		}
		newSubscription := Subscription{
			room_update.Source,
			room_update.SubscriptionId,
			room_update.Facts,
			make(chan []BatchMessage, 1000),
			&sync.WaitGroup{},
			&sync.WaitGroup{},
		}
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

func batch_worker(batch_messages_chan <-chan []BatchMessage, subscriptions_notifications chan<- bool, db *map[string]Fact, subscriptions *Subscriptions) {
	for batch_messages := range batch_messages_chan {
		for _, batch_message := range batch_messages {
			if batch_message.Type == "claim" {
				// claims <- terms
				dbMutex.Lock()
				claim(db, Fact{batch_message.Fact})
				dbMutex.Unlock()
			} else if batch_message.Type == "retract" {
				// retractions <- terms
				dbMutex.Lock()
				retract(db, Fact{batch_message.Fact})
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

func json_term_to_binary_term(term_type, value string) Term {
	if term_type == "integer" {
		intTermValue, err := strconv.Atoi(value)
		checkErr(err)
		return Term{term_type, intToBinary(intTermValue)}
	} else if term_type == "float" {
		floatTermValue64, err := strconv.ParseFloat(value, 32)
		floatTermValue32 := float32(floatTermValue64)
		checkErr(err)
		return Term{term_type, floatToBinary(floatTermValue32)}
	}
	return Term{term_type, []byte(value)}
}


func parse_room_update(source string, data []byte) []RoomUpdate {
	termTypeEnumToString := map[roomupdate.FactType]string{
		roomupdate.FactTypeId: "id",
		roomupdate.FactTypeText: "text",
		roomupdate.FactTypeInteger: "integer",
		roomupdate.FactTypeFloat: "float",
		roomupdate.FactTypeBinary: "binary",
	}
	updateTypeMsgEnumToGoRoomUpdateType := map[roomupdate.UpdateType]RoomUpdateType {
		roomupdate.UpdateTypePing: PING,
		roomupdate.UpdateTypeClaim: CLAIM,
		roomupdate.UpdateTypeRetract: RETRACT,
		roomupdate.UpdateTypeSubscribe: SUBSCRIBE,
		roomupdate.UpdateTypeDeath: DEATH,
		roomupdate.UpdateTypeSubscriptionDeath: SUBSCRIPTION_DEATH,
	}
	// fmt.Println("parse_room_update RAW:")
	// fmt.Println(source)
	// fmt.Println(len(data))
	// fmt.Println(data)
	room_updates_obj := roomupdate.GetRootAsRoomUpdates(data, 0)
	returnUpdates := make([]RoomUpdate, room_updates_obj.UpdatesLength())
	// fmt.Println("N upates:")
	// fmt.Println(room_updates_obj.UpdatesLength())
	for i, _ := range returnUpdates {
		update := new(roomupdate.RoomUpdate)
		room_updates_obj.Updates(update, i)

		updateType, updateTypeExists := updateTypeMsgEnumToGoRoomUpdateType[update.Type()]
		if updateTypeExists == false {
			zap.L().Fatal("unknown update type")
			panic("unknown update type")
		}

		returnUpdateFacts := make([][]Term, 0)

		if updateType == PING {
			// do nothing special
		} else if updateType == SUBSCRIBE {
			for k := 0; k < update.FactsLength(); k += 1 {
				roomUpdateFact := new(roomupdate.Fact)
				update.Facts(roomUpdateFact, k)
				// subscription query parts are encoded as ["text", "$ $ $x ..."]
				fact := parse_fact_string(string(roomUpdateFact.ValueBytes()))
				returnUpdateFacts = append(returnUpdateFacts, fact)
			}
		} else {
			terms := make([]Term, update.FactsLength())
			for k, _ := range terms {
				fact := new(roomupdate.Fact)
				update.Facts(fact, k)
				termType, termTypeExists := termTypeEnumToString[fact.Type()]
				if termTypeExists == false {
					zap.L().Fatal("unknown term type")
					panic("unknown term type")
				}
				terms[k] = Term{termType, fact.ValueBytes()}
			}
			returnUpdateFacts = append(returnUpdateFacts, terms)
		}

		// fmt.Println("SUB ID:")
		// fmt.Println(update.SubscriptionId())
		// fmt.Println("--")
		
		returnUpdates[i] = RoomUpdate{
			updateType,
			string(update.Source()),
			string(update.SubscriptionId()),
			returnUpdateFacts,
		}
	}
	// fmt.Println("PARSED:")
	// fmt.Println(returnUpdates)
	return returnUpdates
}


func parse_room_update_json(source string, msg string) []RoomUpdate {
	// This function can break apart messages into a list,
	// but it should be handed to a subscriber as a batch of messages
	// because a subscriber will only attempt to notify subscribers at the end of a batch
	event_type_len := 9
	source_len := 4
	event_type := msg[0:event_type_len]
	val := msg[(event_type_len + source_len):]
	if event_type == ".....PING" {
		return []RoomUpdate{{PING, source, val, make([][]Term, 0)}}
	} else if event_type == "SUBSCRIBE" {
		subscription_data := SubscriptionData{}
		err := json.Unmarshal([]byte(val), &subscription_data)
		checkErr(err)
		query := make([][]Term, 0)
		for _, fact_string := range subscription_data.Facts {
			fact := parse_fact_string(fact_string)
			query = append(query, fact)
		}
		return []RoomUpdate{{SUBSCRIBE, source, subscription_data.Id, query}}
	} else if event_type == "....BATCH" {
		var batch_messages []BatchMessageJSON
		err := json.Unmarshal([]byte(val), &batch_messages)
		checkErr(err)
		updates := make([]RoomUpdate, len(batch_messages))
		for i, batch_message := range batch_messages {
			terms := make([]Term, len(batch_message.Fact))
			for j, term := range batch_message.Fact {
				terms[j] = json_term_to_binary_term(term[0], term[1])
			}
			if batch_message.Type == "claim" {
				updates[i] = RoomUpdate{CLAIM, source, "", [][]Term{terms}}
			} else if batch_message.Type == "retract" {
				updates[i] = RoomUpdate{RETRACT, source, "", [][]Term{terms}}
			} else if batch_message.Type == "death" {
				// Assume Fact = [["id", "0004"]]
				dying_source := batch_message.Fact[0][1]
				facts := [][]Term{[]Term{Term{"id", []byte(dying_source)}}}
				updates[i] = RoomUpdate{DEATH, source, "", facts}
			} else if batch_message.Type == "subscriptiondeath" {
				// Assume Fact = [["id", "0004"], ["text", ..subscription id..]]
				source := batch_message.Fact[0][1]
				dying_subscription_id := batch_message.Fact[1][1]
				facts := [][]Term{[]Term{
					Term{"id", []byte(source)},
					Term{"text", []byte(dying_subscription_id)},
				}}
				updates[i] = RoomUpdate{SUBSCRIPTION_DEATH, source, "", facts}
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

	subscription_messages := make(chan RoomUpdate, 1000)
	subscriptions_notifications := make(chan bool, 1000)
	notifications := make(chan Notification, 1000)
	batch_messages := make(chan []BatchMessage, 1000)

	go subscribe_worker(subscription_messages, subscriptions_notifications, &subscriptions, notifications, &factDatabase)
	go notification_worker(notifications, client)
	// go debug_database_observer(&factDatabase)
	go batch_worker(batch_messages, subscriptions_notifications, &factDatabase, &subscriptions)

	zap.L().Info("listening...")
	// rootSpan := tracer.StartSpan("run-test")
	// mapc := opentracing.TextMapCarrier(make(map[string]string))
	// err := tracer.Inject(rootSpan.Context(), opentracing.TextMap, mapc)
	// checkErr(err)
	// zap.L().Info(mapc["uber-trace-id"])

	// TODO: switch to RoomUpdateType every where
	roomUpdateTypeToTypeString := map[RoomUpdateType]string{
		CLAIM: "claim",
		RETRACT: "retract",
		DEATH: "death",
		SUBSCRIPTION_DEATH: "subscriptiondeath",
	}

	s := testRoomUpdateSerialization()
	fmt.Println(s)
	testRoomUpdateDeserialization(s)

	s2 := testRoomUpdateSerializationResult()
	fmt.Println(s2)
	testRoomUpdateDeserializationResult(s2)
	
	// go func() {
	// 	time.Sleep(time.Duration(40) * time.Second)
	// 	// rootSpan.Finish()
	// 	// closer.Close()
	// 	panic("time elapsed - ending");
	// }()
	for {
		zmqClient.Lock()
		rawMsg, recvErr := client.RecvMessageBytes(zmq.DONTWAIT)
		if recvErr != nil {
			zmqClient.Unlock()
			time.Sleep(time.Duration(1) * time.Millisecond)
			continue;
		}
		rawMsgId := string(rawMsg[0])
		if (bytes.Equal(rawMsg[1], []byte{9, 254})) {
			fmt.Println("receiving special bytes!")
			_, sendErr := client.SendMessage("1200", []byte{254, 9})
			checkErr(sendErr)
			zmqClient.Unlock()
			time.Sleep(time.Duration(1) * time.Millisecond)
			continue;
		}
		// msg := string(rawMsg[1])
		msg := rawMsg[1]
		zmqClient.Unlock()
		// span := rootSpan.Tracer().StartSpan(
		// 	"zmq-recv-loop",
		// 	opentracing.ChildOf(rootSpan.Context()),
		// )

		updates := parse_room_update(rawMsgId, msg)
		if len(updates) == 1 && updates[0].Type == PING {
			update := updates[0]
			zap.L().Debug("got PING", zap.String("source", update.Source), zap.String("value", update.SubscriptionId))
			// notifications <- Notification{source, val, mapc["uber-trace-id"]}
			notifications <- Notification{update.Source, update.SubscriptionId, make([]QueryResult, 0)}
		} else if len(updates) == 1 && updates[0].Type == SUBSCRIBE {
			subscription_messages <- updates[0]
		} else {
			update_batch_messages := make([]BatchMessage, len(updates))
			for i, update := range updates {
				update_batch_messages[i] = BatchMessage{
					roomUpdateTypeToTypeString[update.Type],
					update.Facts[0], // update.Facts should always have length 1
				}
			}
			batch_messages <- update_batch_messages
		}
		
		/*
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
		*/

		time.Sleep(time.Duration(1) * time.Microsecond)
		// span.Finish()
	}
}
