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

type SelectQueryVariable struct {
  Fact int
  Position int
  Equals []SelectQueryVariable
}

type SubscriptionData struct {
  Id    string   `json:"id"`
  Facts []string `json:"facts"`
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
		fmt.Println(id, factid, position, value, typee)
	}
	err = rows.Err()
	checkErr(err)
}

func send_results(publisher *zmq.Socket, source string, id string, results_str string) {
  // results_str = json.dumps(results)
  msg := fmt.Sprintf("%s%s%s", source, id, results_str)
  fmt.Println("Sending ", msg)
  publisher.Send(msg, 0)
}

func claim_fact(db *sql.DB, fact []Term, source string) {
  tx, err := db.Begin()
  checkErr(err)
  stmt, err := tx.Prepare("INSERT INTO facts (factid, position, value, type) VALUES (?,?,?,?)")
  checkErr(err)
  defer stmt.Close()
  _, err = stmt.Exec(next_fact_id, 0, source, "source")
  checkErr(err)
  for i, term := range fact {
    term_type := term.Type
    term_value := term.Value
  	_, err = stmt.Exec(next_fact_id, i+1, term_value, term_type)
  	checkErr(err)
  }
  tx.Commit()
  next_fact_id += 1
}

func parse_fact_string(parser *participle.Parser, fact_string string) []Term {
  fmt.Println("PARSE", fact_string)
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
  fmt.Println(fact_terms)
  return fact_terms
}

func select_facts(db *sql.DB, query [][]Term, get_ids bool, include_types bool) bool {
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
      fmt.Println("skipping variable with no name")
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
      fmt.Println("skipping variable with no name")
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
  fmt.Println(sql)

  rows, err := db.Query(sql)
	checkErr(err)
	defer rows.Close()
  fmt.Println(":::::::")
  repr.Println(rows, repr.Indent("  "), repr.OmitEmpty(true))
	// for rows.Next() {
	// 	var id int
	// 	var factid int
  //   var position int
  //   var value string
  //   var typee string
	// 	err = rows.Scan(&id, &factid, &position, &value, &typee)
	// 	checkErr(err)
	// 	fmt.Println(id, factid, position, value, typee)
	// }
	err = rows.Err()
	checkErr(err)

  return false
}

func update_all_subscriptions(db *sql.DB) {
  query_part := []Term{Term{"variable", "source"}, Term{"text", "subscription"}, Term{"variable", "subscription_id"}, Term{"postfix", ""}}
  query := [][]Term{query_part}
  // TODO: select_facts
  subscriptions := select_facts(db, query, false, false)
  fmt.Println(subscriptions)
  // for _, row := range subscriptions {
  //   source = row[0]
  //   subscription_id = row[1]
  //   // TODO: get_facts_for_subscription
  //   facts := get_facts_for_subscription(source, subscription_id)
  //   // logging.info("FACTS FOR SUBSCRIPTION {} {}".format(source, subscription_id))
  //   // logging.info(facts)
  //   send_results(publisher, source, subscription_id, facts)
  // }
}

func claim(db *sql.DB, parser *participle.Parser, fact_string string, source string) {
  fact := parse_fact_string(parser, fact_string)  // TODO
  claim_fact(db, fact, source)
  print_all(db)
  update_all_subscriptions(db)
}

func subscribe(db *sql.DB, parser *participle.Parser, fact_strings []string, subscription_id string, source string) {
  for i, fact_string := range fact_strings {
    fmt.Println(subscription_id)
    msg := fmt.Sprintf("subscription \"%s\" %v %s", subscription_id, i, fact_string)
    fmt.Println(msg)
    claim(db, parser, msg, source)
  }
}

func main() {
  parser, err := participle.Build(&Fact{})
  checkErr(err)
  // parse_fact_string("#P5 #0 \"This \\\"is\\\" a test\" one \"two\" 0.5 2 1. .99999 1.23e8 $ $X % %Z")
  // repr.Println(fact, repr.Indent("  "), repr.OmitEmpty(true))

  db, err := sql.Open("sqlite3", ":memory:")
	checkErr(err)
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

	for {
		msg, _ := subscriber.Recv(0)
		fmt.Printf("%s\n", msg)
    event_type := msg[0:event_type_len]
    source := msg[event_type_len:(event_type_len + source_len)]
    val := msg[(event_type_len + source_len):]
    if event_type == ".....PING" {
      send_results(publisher, source, val, "")
    } else if event_type == "....CLAIM" {
      claim(db, parser, val, source)
    // } else if event_type == "..RETRACT" {
    //     retract(val)
    // } else if event_type == "...SELECT" {
    //     json_val = json.loads(val)
    //     select(json_val["facts"], json_val["id"], source)
    } else if event_type == "SUBSCRIBE" {
      subscription_data := SubscriptionData{}
      json.Unmarshal([]byte(val), &subscription_data)
      subscribe(db, parser, subscription_data.Facts, subscription_data.Id, source)
    }
  }
}
