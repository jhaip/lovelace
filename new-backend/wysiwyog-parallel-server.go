package main

import (
	"database/sql"
	"fmt"
  "encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"log"
  zmq "github.com/pebbe/zmq4"
  "github.com/alecthomas/participle"
  // "github.com/alecthomas/repr"
  "strconv"
	"sync"
  "time"
)

var next_fact_id int = 1
var mutex = &sync.Mutex{}
var subscriber_mutex = &sync.Mutex{}

type Term struct {
	Type string
	Value string
}

type SelectQueryVariable struct {
  Fact int
  Position int
  Equals []SelectQueryVariable
}

type SubscriptionData struct {
  Id    string   `json:"id"`
  Facts []string `json:"facts"`
}

type Subscription struct {
  Source string
  Id string
  Query [][]Term
}

type Subscriptions struct {
  Subscriptions []Subscription
}

// Grammar Start
type Fact struct {
  FactTerms []*FactTerm `{ @@ }`
}

type Postfix struct {
  Postfix string `"%"[@Ident]`
}

type Wildcard struct {
  Wildcard string `"$"[@Ident]`
}

type Number struct {
  Number float64 `@Float`
}

type Integer struct {
  Integer int `@Int`
}

type FactTerm struct {
  Postfix *Postfix `@@ |`
  Wildcard *Wildcard `@@ |`
  Id string `"#"(@Ident|@Int) |`
  String string `@String |`
  Integer *Integer `@@ |`
  Number *Number `@@ |`
  Value string `@Ident`
}
//

func init_db(db *sql.DB) {
  sqlStmt := `
  CREATE TABLE IF NOT EXISTS facts (
    id INTEGER PRIMARY KEY,
    factid INTEGER,
    position INTEGER,
    value,
    type TEXT
  );
	`
	_, err := db.Exec(sqlStmt)
	if err != nil {
		log.Printf("%q: %s\n", err, sqlStmt)
		return
	}
}

func checkErr(err error) {
  if err != nil {
    log.Fatal(err)
  }
}

func print_all(db *sql.DB) {
  rows, err := db.Query("SELECT * FROM facts")
	checkErr(err)
	defer rows.Close()
	for rows.Next() {
		var id int
		var factid int
    var position int
    var value string
    var typee string
		err = rows.Scan(&id, &factid, &position, &value, &typee)
		checkErr(err)
		// fmt.Println(id, factid, position, value, typee)
	}
	err = rows.Err()
	checkErr(err)
}

func send_results(publisher *zmq.Socket, source string, id string, results [][]string) {
  start := time.Now()
  results_json_str, err := json.Marshal(results)
  checkErr(err)
  msg := fmt.Sprintf("%s%s%s", source, id, string(results_json_str))
  // fmt.Println("Sending ", msg)
  publisher.Send(msg, zmq.DONTWAIT)
  timeToSendResults := time.Since(start)
  fmt.Printf("time to send results: %s \n", timeToSendResults)
}

func claim_fact(db *sql.DB, fact []Term) {
  tx, err := db.Begin()
  checkErr(err)
  stmt, err := tx.Prepare("INSERT INTO facts (factid, position, value, type) VALUES (?,?,?,?)")
  checkErr(err)
  defer stmt.Close()
  for i, term := range fact {
    term_type := term.Type
    term_value := term.Value
  	_, err = stmt.Exec(next_fact_id, i, term_value, term_type)
  	checkErr(err)
  }
  tx.Commit()
  next_fact_id += 1
}

func parse_fact_string(parser *participle.Parser, fact_string string) []Term {
  // fmt.Println("PARSE", fact_string)
  fact := &Fact{}
  err := parser.ParseString(fact_string, fact)
	checkErr(err)
  fact_terms := make([]Term, 0)
  for _, fact_term := range (*fact).FactTerms {
      if fact_term.String != "" {
        fact_terms = append(fact_terms, Term{"text", fact_term.String})
      }
      if fact_term.Value != "" {
        fact_terms = append(fact_terms, Term{"text", fact_term.Value})
      }
      if fact_term.Id != "" {
        fact_terms = append(fact_terms, Term{"id", fact_term.Id})
      }
      if fact_term.Integer != nil {
        fact_terms = append(fact_terms, Term{"integer", fmt.Sprintf("%v", (*fact_term.Integer).Integer)})
      }
      if fact_term.Number != nil {
        fact_terms = append(fact_terms, Term{"float", fmt.Sprintf("%f", (*fact_term.Number).Number)})
      }
      if fact_term.Wildcard != nil {
        if fact_term.Wildcard == nil {
          fact_terms = append(fact_terms, Term{"variable", ""})
        } else {
          fact_terms = append(fact_terms, Term{"variable", (*fact_term.Wildcard).Wildcard})
        }
      }
      if fact_term.Postfix != nil {
        if fact_term.Postfix == nil {
          fact_terms = append(fact_terms, Term{"postfix", ""})
        } else {
          fact_terms = append(fact_terms, Term{"postfix", (*fact_term.Postfix).Postfix})
        }
      }
      // repr.Println(fact_term, repr.Indent("  "), repr.OmitEmpty(true))
  }
  // fmt.Println(fact_terms)
  return fact_terms
}

func select_facts(db *sql.DB, query [][]Term, get_ids bool, include_types bool) [][]string {
  start := time.Now()
  // include_types = true
  // TODO: what is the return type?
  // variable length type and list of results
  variables := make(map[string]SelectQueryVariable)
  postfixes := make(map[string]SelectQueryVariable)
  for ix, x := range query {
    for iy, y := range x {
      if y.Type == "variable" {
        _, yValueInVariables := variables[y.Value]
        if yValueInVariables {
          variables[y.Value] = SelectQueryVariable{ix, iy, append(variables[y.Value].Equals, SelectQueryVariable{ix, iy, make([]SelectQueryVariable, 0)})}
        } else {
          variables[y.Value] = SelectQueryVariable{ix, iy, make([]SelectQueryVariable, 0)}
        }
      } else if y.Type == "postfix" {
        postfixes[y.Value] = SelectQueryVariable{ix, iy, make([]SelectQueryVariable, 0)}
      }
    }
  }
  sql := "SELECT DISTINCT\n"
  for v, variableValue := range variables {
    if v == "" {
      // fmt.Println("skipping variable with no name")
      continue
    }
    if sql != "SELECT DISTINCT\n" {
      sql += ",\n"
    }
    sql += fmt.Sprintf("facts%d_%d.value as \"%v\"", variableValue.Fact, variableValue.Position, v)
    if include_types {
      sql += ",\n"
      sql += fmt.Sprintf("facts%d_%d.type as \"%v_type\"", variableValue.Fact, variableValue.Position, v)
    }
    if get_ids {
      sql += ",\n"
      sql += fmt.Sprintf("facts%d_%d.id as \"other\"", variableValue.Fact, variableValue.Position)
    }
  }
  for v, postfixValue := range postfixes {
    if v == "" {
      // fmt.Println("skipping variable with no name")
      continue
    }
    if sql != "SELECT DISTINCT\n" {
      sql += ",\n"
    }
    sql += fmt.Sprintf("facts%d_%d.value as \"%v\"", postfixValue.Fact, postfixValue.Position, v)
    if include_types {
      sql += ",\n"
      sql += fmt.Sprintf("facts%d_%d.type as \"%v_type\"", postfixValue.Fact, postfixValue.Position, v)
    }
  }
  sql += "\nFROM\n"
  for ix, x := range query {
    for iy, _ := range x {
      if ix != 0 || iy != 0 {
        sql += ",\n"
      }
      sql += fmt.Sprintf("facts as facts%d_%d", ix, iy)
    }
  }
  sql += "\nWHERE\n"
  for ix, x := range query {
    for iy, y := range x {
      sql += fmt.Sprintf("facts%d_0.factid = facts%d_%d.factid AND\n", ix, ix, iy)
      if y.Type == "postfix" {
        sql += fmt.Sprintf("facts%d_%d.position >= %d AND\n", ix, iy, iy)
      } else {
        sql += fmt.Sprintf("facts%d_%d.position = %d AND\n", ix, iy, iy)
      }
      if y.Type != "variable" && y.Type != "postfix" {
        sql += fmt.Sprintf("facts%d_%d.type = '%s' AND\n", ix, iy, y.Type)
      }
      if y.Type == "text" || y.Type == "source" {
        sql += fmt.Sprintf("facts%d_%d.value = '%s' AND\n", ix, iy, y.Value)
      } else if y.Type != "variable" && y.Type != "postfix" {
        sql += fmt.Sprintf("facts%d_%d.value = %s AND\n", ix, iy, y.Value)
      }
    }
  }
  for _, v := range variables {
    for _, k := range v.Equals {
      sql += fmt.Sprintf("facts%d_%d.value = facts%d_%d.value AND\n", v.Fact, v.Position, k.Fact, k.Position)
    }
  }
  if sql[len(sql)-4:] == "AND\n" {
    sql = sql[:len(sql)-4]
  }
  // fmt.Println(sql)
  makeSqlQuery := time.Since(start)
  start2 := time.Now()
  rows, err := db.Query(sql)
  doQuery := time.Since(start2)
  start3 := time.Now()
	checkErr(err)
	defer rows.Close()
  // fmt.Println(":::::::")
  // repr.Println(rows, repr.Indent("  "), repr.OmitEmpty(true))
  // fmt.Println("rows.Columns")
  // fmt.Println(rows.Columns())
  // fmt.Println("rows.ColumnTypes")
  // fmt.Println(rows.ColumnTypes())
  column_types, err := rows.ColumnTypes()
  checkErr(err)
  fmt.Println(column_types[0].Name())
  // fmt.Println(column_types[1].Name())
  result_columns, err := rows.Columns()
  checkErr(err)
  results := make([][]string, 0)
  // NOTE: golang's Next() and Scan() are slow :(, even slower than Python
  // https://github.com/mattn/go-sqlite3/issues/379
	for rows.Next() {
    row_results := make([]string, len(result_columns))
    row_results_pointers := make([]interface{}, len(result_columns))
    for i, _ := range row_results {
      row_results_pointers[i] = &row_results[i]
    }
	// 	var id int
	// 	var factid int
  //   var position int
  //   var value string
  //   var typee string
    // err = rows.Scan(&id, &factid, &position, &value, &typee)
    // var source interface{}
    // var subscription_id interface{}
    // var source string
    // var subscription_id string
		err = rows.Scan(row_results_pointers...)
    // err = rows.Scan(&row_results[0], &row_results[1])
    // err = rows.Scan(&source, &subscription_id)
		checkErr(err)
    // fmt.Println("row_results")
		// fmt.Println(row_results)
    // // fmt.Println(source.(type), subscription_id.(type))
    // msg := fmt.Sprintf("%v %v", source, subscription_id)
    // // fmt.Println(msg)
    // switch subscription_id.(type) {
		// case int:
		// 	// fmt.Printf("int: %d\n", subscription_id.(int))
		// case string:
		// 	// fmt.Printf("string: %s\n", subscription_id.(string))
		// case bool:
		// 	// fmt.Printf("bool: %t\n", subscription_id.(bool))
    // default:
    //   // fmt.Printf("didn't match type :(\n")
		// }
    // // fmt.Println(source, subscription_id)
    // // fmt.Println(id, factid, position, value, typee)
    results = append(results, row_results)
	}
	err = rows.Err()
	checkErr(err)
  // fmt.Println("final results")
  // fmt.Println(results)
  putTogetherResults := time.Since(start3)
  fmt.Printf("_ _ _ _makeSqlQuery  : %s \n", makeSqlQuery)
  fmt.Printf("_ _ _ _doQuery: %s \n", doQuery)
  fmt.Printf("_ _ _ _putTogetherResults     : %s \n", putTogetherResults)
  return results
}

func get_facts_for_subscription(db *sql.DB, source string, subscription_id string) [][]string {
  start := time.Now()
  selection_query_part := []Term{Term{"source", source}, Term{"text", "subscription"}, Term{"text", subscription_id}, Term{"variable", "part"}, Term{"postfix", "X"}}
  selection_query := [][]Term{selection_query_part}
  // fmt.Println("SELECTION QUERY::::::")
  start2 := time.Now()
  r := select_facts(db, selection_query, false, true)
  selectFactsTime := time.Since(start2)
	// fmt.Println("SELECTION QUERY RESULTS -------!!!!!!!!!")
	// fmt.Println(r)
  query := make([][]Term, 0)
  for _, row := range r {
		// fmt.Println("SELECTION QUERY RESULTS ----------------------")
		// fmt.Println(row)
    subscription_part, err := strconv.Atoi(row[0])
    checkErr(err)
    if subscription_part >= len(query) {
      query = append(query, make([]Term, 0))
    }
    query[subscription_part] = append(query[subscription_part], Term{row[3], row[2]})
  }
  // fmt.Println("GET FACTS QUERY::::::")
  start3 := time.Now()
  results := select_facts(db, query, false, false)
  getResultsTime := time.Since(start3)
  getFactsTotalTime := time.Since(start)
  fmt.Printf("__selectFactsTime: %s \n", selectFactsTime)
  fmt.Printf("__getResultsTime: %s \n", getResultsTime)
  fmt.Printf("getFactsTotalTime: %s \n", getFactsTotalTime)
  return results
}

func single_subscriber_update(db *sql.DB, publisher *zmq.Socket, subscription Subscription, wg *sync.WaitGroup, i int) {
	start := time.Now()
	// fmt.Println("pre SELECTING %v", subscription.Query)
	results := select_facts(db, subscription.Query, false, false)
	// fmt.Println("DONE SELECTING")
	send_results(publisher, subscription.Source, subscription.Id, results)
	wg.Done()
	duration := time.Since(start)
	fmt.Printf("SINGLE SUBSCRIBER DONE %v -- %s\n", i, duration)
}

func update_all_subscriptions(db *sql.DB, publisher *zmq.Socket, subscriptions Subscriptions) {
	mutex.Lock()
	var wg sync.WaitGroup
	wg.Add(len(subscriptions.Subscriptions))
	// TODO: there may be a race condition if the contents of subscriptions changes when running this func.
	// How about just passing in a copy of the subscriptions

  for i, subscription := range subscriptions.Subscriptions {
		go single_subscriber_update(db, publisher, subscription, &wg, i)
  }
	// fmt.Println("WAITING FOR ALL THINGS TO END")
	wg.Wait()
	mutex.Unlock()
	// fmt.Println("done")
}

func claim(db *sql.DB, parser *participle.Parser, publisher *zmq.Socket, subscriptions *Subscriptions, fact_string string, source string) {
  start := time.Now()
  fact := parse_fact_string(parser, fact_string)
  parseTime := time.Since(start)
  start2 := time.Now()
	// fact []Term
	fact = append([]Term{Term{"source", source}}, fact...)
  claim_fact(db, fact)
  claimTime := time.Since(start2)
  // print_all(db)
  start3 := time.Now()
  update_all_subscriptions(db, publisher, *subscriptions)
  updateSubscribersTime := time.Since(start3)
  fmt.Printf("...parse: %s \n", parseTime)
  fmt.Printf("....claim: %s \n", claimTime)
  fmt.Printf(".....update: %s \n", updateSubscribersTime)
  fmt.Printf("subscription: %v \n", (*subscriptions).Subscriptions)
}

func subscribe(db *sql.DB, parser *participle.Parser, publisher *zmq.Socket, subscriptions *Subscriptions, fact_strings []string, subscription_id string, source string) {
  query := make([][]Term, 0)
  for i, fact_string := range fact_strings {
    // fmt.Println(subscription_id)
    msg := fmt.Sprintf("subscription \"%s\" %v %s", subscription_id, i, fact_string)
    // fmt.Println(msg)
    claim(db, parser, publisher, subscriptions, msg, source)
    fact := parse_fact_string(parser, fact_string)
    query = append(query, fact)
  }
  (*subscriptions).Subscriptions = append((*subscriptions).Subscriptions, Subscription{source, subscription_id, query})
  // fmt.Println("ADDED SUBSCRIPTOIN!")
  // fmt.Printf("len: %v", len((*subscriptions).Subscriptions))
  update_all_subscriptions(db, publisher, *subscriptions)
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
			subscriber_mutex.Lock()
		  (*subscriptions).Subscriptions = append((*subscriptions).Subscriptions, Subscription{source, subscription_data.Id, query})
			subscriber_mutex.Unlock()
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

func claim_worker(claims <-chan []Term, subscriptions_notifications chan<- bool, db *sql.DB) {
	for fact := range claims {
		fmt.Printf("SHOULD CLAIM: %v\n", claim)
		mutex.Lock()
		claim_fact(db, fact)
		mutex.Unlock()
		fmt.Println("claim done")
	  subscriptions_notifications <- true // update_all_subscriptions(db, publisher, subscriptions)
	}
}

func notify_subscribers_worker(notify_subscribers <-chan bool, subscriber_worker_finished chan<- bool, db *sql.DB, publisher *zmq.Socket, subscriptions *Subscriptions) {
	// TODO: passing in subscriptions is probably not safe because it can be written in the other goroutine
	for range notify_subscribers {
		fmt.Println("inside notify_subscribers_worker loop")
		start := time.Now()
		subscriber_mutex.Lock()
		update_all_subscriptions(db, publisher, *subscriptions)
		subscriber_mutex.Unlock()
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
  parser, err := participle.Build(&Fact{})
  checkErr(err)
  // parse_fact_string("#P5 #0 \"This \\\"is\\\" a test\" one \"two\" 0.5 2 1. .99999 1.23e8 $ $X % %Z")
  // repr.Println(fact, repr.Indent("  "), repr.OmitEmpty(true))

  // :memory:
	// "file::memory:?mode=memory&cache=shared"
  db, err := sql.Open("sqlite3", "file:memdb1?mode=memory&cache=shared&_busy_timeout=9999999")
	// db.SetMaxOpenConns(1)
	checkErr(err)
	defer db.Close()
	// db_readonly, err := sql.Open("sqlite3", "file:memdb1?mode=memory&cache=shared&_busy_timeout=9999999")
	// db_readonly.SetMaxOpenConns(1)
	// defer db_readonly.Close()

  init_db(db)

  subscriptions := Subscriptions{}

  // fmt.Println("Connecting to hello world server...")
  publisher, _ := zmq.NewSocket(zmq.PUB)
  defer publisher.Close()
  publisher.Connect("tcp://localhost:5555")
  subscriber, _ := zmq.NewSocket(zmq.SUB)
	defer subscriber.Close()
  subscriber.Connect("tcp://localhost:5556")
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
	go claim_worker(claims, subscriptions_notifications, db)
	go notify_subscribers_worker(notify_subscribers, subscriber_worker_finished, db, publisher, &subscriptions)
	go debounce_subscriber_worker(subscriptions_notifications, subscriber_worker_finished, notify_subscribers)

	for {
		msg, _ := subscriber.Recv(0)
		// fmt.Printf("%s\n", msg)
    event_type := msg[0:event_type_len]
    source := msg[event_type_len:(event_type_len + source_len)]
    val := msg[(event_type_len + source_len):]
    if event_type == ".....PING" {
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
      // subscription_data := SubscriptionData{}
      // json.Unmarshal([]byte(val), &subscription_data)
      // subscribe(db, parser, publisher, &subscriptions, subscription_data.Facts, subscription_data.Id, source)
    }
  }
}
