package main

import (
  "fmt"
  "github.com/alecthomas/repr"
)

type Term struct {
	Type string
	Value string
}

type Fact struct {
  Terms []Term
}

type QueryResult struct {
  Result map[string]Term
}

// TODO: handle postfix
func fact_match(A Fact, B Fact, env QueryResult) (bool, QueryResult) {
  if len(A.Terms) != len(B.Terms) {
    return false, QueryResult{}
  }
  new_env := QueryResult{map[string]Term{}}
  for k,v := range env.Result {
    new_env.Result[k] = v
  }
  for i, A_term := range A.Terms {
    B_term := B.Terms[i]
    did_match, tmp_new_env := term_match(A_term, B_term, new_env)
    if did_match == false {
      return false, QueryResult{}
    }
    new_env = tmp_new_env
  }
  return true, new_env
}

// TODO: handle postfix
func term_match(A Term, B Term, env QueryResult) (bool, QueryResult) {
  if A.Type == "variable" {
    variable_name := A.Value
    // "Wilcard" matches all but doesn't have a result
    if variable_name == "" {
      return true, env
    }
    _, variable_name_in_result := env.Result[variable_name]
    if variable_name_in_result {
      return term_match(env.Result[variable_name], B, env)
    } else {
      new_env := QueryResult{map[string]Term{}}
      for k,v := range env.Result {
        new_env.Result[k] = v
      }
      new_env.Result[variable_name] = B
      return true, new_env
    }
  } else if A.Type == B.Type && A.Value == B.Value {
    return true, env
  }
  return false, QueryResult{}
}

func collect_solutions(facts []Fact, query []Fact, env QueryResult) []QueryResult {
  if len(query) == 0 {
    return []QueryResult{env}
  }
  first_query_fact := query[0]
  solutions := make([]QueryResult, 0)
  for _, fact := range facts {
    did_match, new_env := fact_match(first_query_fact, fact, env)
    if did_match {
      solutions = append(solutions, collect_solutions(facts, query[1:], new_env)...)
    }
  }
  return solutions
}

func select_facts(facts []Fact, query []Fact) []QueryResult {
  empty_env := QueryResult{map[string]Term{}}
  return collect_solutions(facts, query, empty_env)
}

func main() {
  facts := make([]Fact, 3)
  facts[0] = Fact{[]Term{Term{"source", "1"}, Term{"text", "Man"}, Term{"integer", "5"}, Term{"text", "toes"}}}
  facts[1] = Fact{[]Term{Term{"source", "1"}, Term{"text", "Snake"}, Term{"text", "no"}, Term{"text", "toes"}}}
  facts[2] = Fact{[]Term{Term{"source", "1"}, Term{"text", "Snake"}, Term{"text", "is"}, Term{"text", "red"}}}
  query1 := make([]Fact, 1)
  query1[0] = Fact{[]Term{Term{"variable", ""}, Term{"variable", "X"}, Term{"variable", "Y"}, Term{"text", "toes"}}}
  results1 := select_facts(facts, query1)
  // fmt.Println(results1)
  fmt.Println("RESULTS 1 - several matches:\n")
  repr.Println(results1, repr.Indent("  "), repr.OmitEmpty(true))

  query2 := make([]Fact, 1)
  query2[0] = Fact{[]Term{Term{"source", "100"}, Term{"variable", "X"}, Term{"variable", "Y"}, Term{"text", "toes"}}}
  results2 := select_facts(facts, query2)
  fmt.Println("RESULTS 2 - no matches:\n")
  repr.Println(results2, repr.Indent("  "), repr.OmitEmpty(true))

  query3 := make([]Fact, 1)
  query3[0] = Fact{[]Term{Term{"source", "1"}, Term{"text", "Man"}, Term{"integer", "5"}, Term{"text", "toes"}}}
  results3 := select_facts(facts, query3)
  fmt.Println("RESULTS 3 - exact match:\n")
  repr.Println(results3, repr.Indent("  "), repr.OmitEmpty(true))

  // TODO: handle postfix
  // TODO: a better way to differentiate no results, vs results but without a name (for exact match)
  // TODO: handle claims
  // TODO: handle subscriptions
  // TODO: a way to detect if a claim will include a part of a subscription?
  // TODO: test claims with multiple parts and reused variables
}
