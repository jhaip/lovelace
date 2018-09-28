package main

import (
	"database/sql"
	"fmt"
  "encoding/json"
	_ "github.com/mattn/go-sqlite3"
	"log"
  zmq "github.com/pebbe/zmq4"
  "github.com/alecthomas/participle"
  "github.com/alecthomas/repr"
)

var next_fact_id int = 1

type Term struct {
	Type string
	Value string
}

type SubscriptionData struct {
  Id    string   `json:"id"`
  Facts []string `json:"facts"`
}

// Grammar Start
type Fact struct {
  FactTerm []*FactTerm `{ @@ }`
}

type Postfix struct {
  Postfix string `"%"[@Ident]`
}

type Wildcard struct {
  Wildcard string `"$"[@Ident]`
}

type FactTerm struct {
  Postfix *Postfix `@@ |`
  Wildcard *Wildcard `@@ |`
  Id string `"#"(@Ident|@Int) |`
  String string `@String |`
  Integer int `@Int |`
  Number float64 `@Float |`
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

func send_results(publisher *zmq.Socket, source string, id string, results_str string) {
  // results_str = json.dumps(results)
  msg := fmt.Sprintf("%s%s%s", source, id, results_str)
  fmt.Println("Sending ", msg)
  publisher.Send(msg, 0)
}

func claim_fact(db *sql.DB, fact []Term, source string) {
  tx, err := db.Begin()
  if err != nil {
  	log.Fatal(err)
  }
  stmt, err := tx.Prepare("INSERT INTO facts (factid, position, value, type) VALUES (?,?,?,?)")
  if err != nil {
  	log.Fatal(err)
  }
  defer stmt.Close()
  _, err = stmt.Exec(next_fact_id, 0, source, "source")
  if err != nil {
    log.Fatal(err)
  }
  for i, term := range fact {
    term_type := term.Type
    term_value := term.Value
  	_, err = stmt.Exec(next_fact_id, i+1, term_value, term_type)
  	if err != nil {
  		log.Fatal(err)
  	}
  }
  tx.Commit()
  next_fact_id += 1
}

func parse_fact_string(fact_string string) []Term {
  fmt.Println("PARSE", fact_string)
  fact := make([]Term, 0)
  return fact
}

func claim(db *sql.DB, fact_string string, source string) {
  fact := parse_fact_string(fact_string)  // TODO
  claim_fact(db, fact, source)
  // update_all_subscriptions()  // TODO
}

func subscribe(db *sql.DB, fact_strings []string, subscription_id string, source string) {
  for i, fact_string := range fact_strings {
    fmt.Println(subscription_id)
    msg := fmt.Sprintf("subscription \"%s\" %v %s", subscription_id, i, fact_string)
    fmt.Println(msg)
    claim(db, msg, source)
  }
}

func main() {
  db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

  init_db(db)

  fmt.Println("Connecting to hello world server...")
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

  parser, err := participle.Build(&Fact{})
  if err != nil {
		panic(err)
	}
  fact := &Fact{}
  err = parser.ParseString("#P5 #0 \"This \\\"is\\\" a test\" one \"two\" 0.5 1. .99999 1.23e8 $ $X % %Z", fact)
	if err != nil {
		panic(err)
	}
  repr.Println(fact, repr.Indent("  "), repr.OmitEmpty(true))

	for {
		msg, _ := subscriber.Recv(0)
		fmt.Printf("%s\n", msg)
    event_type := msg[0:event_type_len]
    source := msg[event_type_len:(event_type_len + source_len)]
    val := msg[(event_type_len + source_len):]
    if event_type == ".....PING" {
      send_results(publisher, source, val, "")
    // } else if event_type == "....CLAIM" {
    //     claim(val, source)
    // } else if event_type == "..RETRACT" {
    //     retract(val)
    // } else if event_type == "...SELECT" {
    //     json_val = json.loads(val)
    //     select(json_val["facts"], json_val["id"], source)
    } else if event_type == "SUBSCRIBE" {
      subscription_data := SubscriptionData{}
      json.Unmarshal([]byte(val), &subscription_data)
      subscribe(db, subscription_data.Facts, subscription_data.Id, source)
    }
  }
}
