package main

import (
	"fmt"

	"github.com/alecthomas/participle"
)

// type Term struct {
// 	Type  string
// 	Value string
// }

// Grammar Start
type ParsedFact struct {
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
	// NegativeNumber float64 `"-"@Float`
}

type Integer struct {
	Integer int `@Int`
	// NegativeInteger int `"-"@Int`
}

type FactTerm struct {
	Postfix  *Postfix  `@@ |`
	Wildcard *Wildcard `@@ |`
	Id       string    `"#"(@Ident|@Int) |`
	Integer  *Integer  `@@ |`
	Number   *Number   `@@ |`
	String   string    `@String |`
	Value    string    `@Ident`
}

//

// func checkErr(err error) {
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// }

func make_parser() (*participle.Parser, error) {
	return participle.Build(&ParsedFact{})
}

func parse_fact_string(parser *participle.Parser, fact_string string) []Term {
	// fmt.Println("PARSE", fact_string)
	fact := &ParsedFact{}
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
			var val int
			// if (*fact_term.Integer).NegativeInteger < 0 {
			// 	val = (*fact_term.Integer).NegativeInteger
			// } else {
			// 	val = (*fact_term.Integer).Integer
			// }
			val = (*fact_term.Integer).Integer
			fact_terms = append(fact_terms, Term{"integer", fmt.Sprintf("%v", val)})
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
