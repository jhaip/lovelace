package main

import (
	"testing"

	"github.com/alecthomas/repr"
)

func init_fact_map() map[string]Fact {
	factMap := make_fact_database()
	fact0 := Fact{[]Term{Term{"source", "1"}, Term{"text", "Man"}, Term{"integer", "5"}, Term{"text", "toes"}}}
	fact1 := Fact{[]Term{Term{"source", "1"}, Term{"text", "Snake"}, Term{"text", "no"}, Term{"text", "toes"}}}
	fact2 := Fact{[]Term{Term{"source", "1"}, Term{"text", "Snake"}, Term{"text", "is"}, Term{"text", "red"}}}
	fact3 := Fact{[]Term{Term{"source", "2"}, Term{"text", "Bird"}, Term{"integer", "3"}, Term{"text", "toes"}}}
	fact4 := Fact{[]Term{Term{"source", "2"}, Term{"text", "subscription"}, Term{"variable", "X"}, Term{"text", "is"}, Term{"postfix", "Y"}}}
	fact5 := Fact{[]Term{Term{"source", "0001"}, Term{"text", "man"}, Term{"integer", "1"}, Term{"text", "has"}, Term{"integer", "95"}, Term{"text", "toes"}}}
	factMap[fact_to_string(fact0)] = fact0
	factMap[fact_to_string(fact1)] = fact1
	factMap[fact_to_string(fact2)] = fact2
	factMap[fact_to_string(fact3)] = fact3
	factMap[fact_to_string(fact4)] = fact4
	factMap[fact_to_string(fact5)] = fact5
	return factMap
}

func TestQueryBasicSeveralMatches(t *testing.T) {
	factMap := init_fact_map()
	query1 := make([]Fact, 1)
	query1[0] = Fact{[]Term{Term{"variable", ""}, Term{"variable", "X"}, Term{"variable", "Y"}, Term{"text", "toes"}}}
	results1 := select_facts(factMap, query1)
	repr.Println(results1, repr.Indent("  "), repr.OmitEmpty(true))
}

func TestQueryNoMatches(t *testing.T) {
	factMap := init_fact_map()
	query2 := make([]Fact, 1)
	query2[0] = Fact{[]Term{Term{"source", "100"}, Term{"variable", "X"}, Term{"variable", "Y"}, Term{"text", "toes"}}}
	results2 := select_facts(factMap, query2)
	repr.Println(results2, repr.Indent("  "), repr.OmitEmpty(true))
}

func TestQueryExactMatch(t *testing.T) {
	factMap := init_fact_map()
	query3 := make([]Fact, 1)
	query3[0] = Fact{[]Term{Term{"source", "1"}, Term{"text", "Man"}, Term{"integer", "5"}, Term{"text", "toes"}}}
	results3 := select_facts(factMap, query3)
	repr.Println(results3, repr.Indent("  "), repr.OmitEmpty(true))
}

func TestQueryMultiplePartQuery(t *testing.T) {
	factMap := init_fact_map()
	query4 := make([]Fact, 2)
	query4[0] = Fact{[]Term{Term{"source", "1"}, Term{"variable", "X"}, Term{"variable", "Y"}, Term{"text", "toes"}}}
	query4[1] = Fact{[]Term{Term{"source", "1"}, Term{"variable", "X"}, Term{"text", "is"}, Term{"variable", "Z"}}}
	results4 := select_facts(factMap, query4)
	repr.Println(results4, repr.Indent("  "), repr.OmitEmpty(true))
}

func TestQueryPostfixWithName(t *testing.T) {
	factMap := init_fact_map()
	query5 := make([]Fact, 1)
	query5[0] = Fact{[]Term{Term{"source", "1"}, Term{"postfix", "X"}}}
	results5 := select_facts(factMap, query5)
	repr.Println(results5, repr.Indent("  "), repr.OmitEmpty(true))
}

func TestQueryWildcardPostfix(t *testing.T) {
	factMap := init_fact_map()
	query6 := make([]Fact, 1)
	query6[0] = Fact{[]Term{Term{"source", "1"}, Term{"text", "Man"}, Term{"postfix", ""}}}
	results6 := select_facts(factMap, query6)
	repr.Println(results6, repr.Indent("  "), repr.OmitEmpty(true))
}

func TestQueryPostfixWithNamesAndSpecialTypes(t *testing.T) {
	factMap := init_fact_map()
	query7 := make([]Fact, 1)
	query7[0] = Fact{[]Term{Term{"variable", ""}, Term{"text", "subscription"}, Term{"postfix", "X"}}}
	results7 := select_facts(factMap, query7)
	repr.Println(results7, repr.Indent("  "), repr.OmitEmpty(true))
}

func TestQueryVariablesAndWildcards(t *testing.T) {
	factMap := init_fact_map()
	query2 := make([]Fact, 1)
	query2[0] = Fact{[]Term{Term{"variable", ""}, Term{"variable", "X"}, Term{"integer", "1"}, Term{"text", "has"}, Term{"variable", "Y"}, Term{"text", "toes"}}}
	results2 := select_facts(factMap, query2)
	repr.Println(results2, repr.Indent("  "), repr.OmitEmpty(true))
}

func TestRetractExactMatch(t *testing.T) {
	factMap := init_fact_map()
	originalLen := len(factMap)
	factQuery := Fact{[]Term{Term{"source", "1"}, Term{"text", "Man"}, Term{"integer", "5"}, Term{"text", "toes"}}}
	retract(&factMap, factQuery)
	if len(factMap) != originalLen-1 {
		t.Error("Fact was not removed")
	}
}

func TestRetractWithWilcard(t *testing.T) {
	factMap := init_fact_map()
	originalLen := len(factMap)
	factQuery := Fact{[]Term{Term{"source", "1"}, Term{"text", "Man"}, Term{"variable", ""}, Term{"text", "toes"}}}
	retract(&factMap, factQuery)
	if len(factMap) != originalLen-1 {
		t.Error("Fact was not removed")
	}
}

func TestRetractAllFromSource(t *testing.T) {
	factMap := init_fact_map()
	originalLen := len(factMap)
	factQuery := Fact{[]Term{Term{"source", "1"}, Term{"postfix", ""}}}
	retract(&factMap, factQuery)
	if len(factMap) != originalLen-3 {
		t.Error("Fact was not removed")
	}
}

func TestClaim(t *testing.T) {
	factMap := init_fact_map()
	originalLen := len(factMap)
	fact := Fact{[]Term{Term{"source", "10"}, Term{"text", "Word"}, Term{"integer", "50"}}}
	claim(&factMap, fact)
	claim(&factMap, fact) // Redundant claim should have no effect
	if len(factMap) != originalLen+1 {
		t.Error("Fact was not removed")
	}
}
