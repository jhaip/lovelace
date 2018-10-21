package main

import (
	"testing"

	"github.com/alecthomas/repr"
)

func checkTerms(terms, expected_terms []Term, t *testing.T) {
	if len(terms) != len(expected_terms) {
		t.Error("Wrong number of terms")
	}
	for i, term := range terms {
		if term.Type != expected_terms[i].Type {
			t.Error("Wrong term type for term ", i, "-- expected", expected_terms[i].Type, expected_terms[i].Value, "!=", term.Type, term.Value)
		}
		if term.Value != expected_terms[i].Value {
			t.Error("Wrong term value for term ", i, "-- expected", expected_terms[i].Type, expected_terms[i].Value, "!=", term.Type, term.Value)
		}
	}
}

func TestParse(t *testing.T) {
	parser, _ := make_parser()
	terms := parse_fact_string(parser, "#P5 #0 \"This \\\"is\\\" a test\" one \"two\" 0.5 2$ $X % %Z")
	repr.Println(terms, repr.Indent("  "), repr.OmitEmpty(true))
	expected_terms := []Term{
		Term{"id", "P5"},
		Term{"id", "0"},
		Term{"text", "This \"is\" a test"},
		Term{"text", "one"},
		Term{"text", "two"},
		Term{"float", "0.500000"},
		Term{"integer", "2"},
		Term{"variable", ""},
		Term{"variable", "X"},
		Term{"postfix", ""},
		Term{"postfix", "Z"},
	}
	checkTerms(terms, expected_terms, t)
}

func TestParseNumbers(t *testing.T) {
	parser, _ := make_parser()
	terms := parse_fact_string(parser, "0.5 2 -2 1. -1.0 .99999 1.23e8")
	repr.Println(terms, repr.Indent("  "), repr.OmitEmpty(true))
	expected_terms := []Term{
		Term{"float", "0.500000"},
		Term{"integer", "2"},
		Term{"integer", "-2"},
		Term{"float", "1.000000"},
		Term{"float", "-1.000000"},
		Term{"float", "0.999990"},
		Term{"float", "123000000.000000"},
	}
	checkTerms(terms, expected_terms, t)
}

func TestParseVariables(t *testing.T) {
	parser, _ := make_parser()
	terms := parse_fact_string(parser, "$ $X $Y $ one1 $1")
	repr.Println(terms, repr.Indent("  "), repr.OmitEmpty(true))
	expected_terms := []Term{
		Term{"variable", ""},
		Term{"variable", "X"},
		Term{"variable", "Y"},
		Term{"variable", ""},
		Term{"text", "one1"},
		Term{"variable", "1"},
	}
	checkTerms(terms, expected_terms, t)
}

func TestParseWithNonWordCharacters(t *testing.T) {
	parser, _ := make_parser()
	terms := parse_fact_string(parser, "#1800 paper 39 at TL (3.0, 0.9) @ 123949583")
	repr.Println(terms, repr.Indent("  "), repr.OmitEmpty(true))
	expected_terms := []Term{
		Term{"id", "1800"},
		Term{"text", "paper"},
		Term{"integer", "39"},
		Term{"text", "at"},
		Term{"text", "TL"},
		Term{"text", "("},
		Term{"float", "3.000000"},
		Term{"text", ","},
		Term{"float", "0.900000"},
		Term{"text", ")"},
		Term{"text", "@"},
		Term{"integer", "123949583"},
	}
	checkTerms(terms, expected_terms, t)
}
