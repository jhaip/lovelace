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

func term_to_string(term Term) string {
  switch term.Type {
  case "source":
    return "#" + term.Value
  case "variable":
    return "$" + term.Value
  case "postfix":
    return "%" + term.Value
  default:
    return term.Value
  }
}

func terms_to_string(terms []Term) string {
  str := ""
  for i, term := range terms {
    // TODO: may want to handle special types like variables or sources in a special way?
    if i > 0 {
      str += " "
    }
    str += term_to_string(term)
  }
  return str
}

func fact_match(A Fact, B Fact, env QueryResult) (bool, QueryResult) {
  A_has_postfix := A.Terms[len(A.Terms) - 1].Type == "postfix"
  if A_has_postfix {
    if len(A.Terms) > len(B.Terms) {
      return false, QueryResult{}
    }
  } else if len(A.Terms) != len(B.Terms) {
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
    if A_term.Type == "postfix" {
      postfix_variable_name := A_term.Value
      if postfix_variable_name != "" {
        new_env.Result[postfix_variable_name] = Term{"text", terms_to_string(B.Terms[i:])}
      }
      break
    }
  }
  return true, new_env
}

func term_match(A Term, B Term, env QueryResult) (bool, QueryResult) {
  if A.Type == "variable" || A.Type == "postfix" {
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
  facts := make([]Fact, 5)
  facts[0] = Fact{[]Term{Term{"source", "1"}, Term{"text", "Man"}, Term{"integer", "5"}, Term{"text", "toes"}}}
  facts[1] = Fact{[]Term{Term{"source", "1"}, Term{"text", "Snake"}, Term{"text", "no"}, Term{"text", "toes"}}}
  facts[2] = Fact{[]Term{Term{"source", "1"}, Term{"text", "Snake"}, Term{"text", "is"}, Term{"text", "red"}}}
  facts[3] = Fact{[]Term{Term{"source", "2"}, Term{"text", "Bird"}, Term{"integer", "3"}, Term{"text", "toes"}}}
  facts[4] = Fact{[]Term{Term{"source", "2"}, Term{"text", "subscription"}, Term{"variable", "X"}, Term{"text", "is"}, Term{"postfix", "Y"}}}
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

  query4 := make([]Fact, 2)
  query4[0] = Fact{[]Term{Term{"source", "1"}, Term{"variable", "X"}, Term{"variable", "Y"}, Term{"text", "toes"}}}
  query4[1] = Fact{[]Term{Term{"source", "1"}, Term{"variable", "X"}, Term{"text", "is"}, Term{"variable", "Z"}}}
  results4 := select_facts(facts, query4)
  fmt.Println("RESULTS 4 - multiple query:\n")
  repr.Println(results4, repr.Indent("  "), repr.OmitEmpty(true))

  query5 := make([]Fact, 1)
  query5[0] = Fact{[]Term{Term{"source", "1"}, Term{"postfix", "X"}}}
  results5 := select_facts(facts, query5)
  fmt.Println("RESULTS 5 - postfix with name:\n")
  repr.Println(results5, repr.Indent("  "), repr.OmitEmpty(true))

  query6 := make([]Fact, 1)
  query6[0] = Fact{[]Term{Term{"source", "1"}, Term{"text", "Man"}, Term{"postfix", ""}}}
  results6 := select_facts(facts, query6)
  fmt.Println("RESULTS 6 - wildcard postfix:\n")
  repr.Println(results6, repr.Indent("  "), repr.OmitEmpty(true))

  query7 := make([]Fact, 1)
  query7[0] = Fact{[]Term{Term{"variable", ""}, Term{"text", "subscription"}, Term{"postfix", "X"}}}
  results7 := select_facts(facts, query7)
  fmt.Println("RESULTS 7 - postfix with names and special types:\n")
  repr.Println(results7, repr.Indent("  "), repr.OmitEmpty(true))

  // TODO: a better way to differentiate no results, vs results but without a name (for exact match)
  // TODO: handle claims
  // TODO: handle subscriptions
  // TODO: a way to detect if a claim will include a part of a subscription?
  // TODO: return in format compatible with rest of code
}
