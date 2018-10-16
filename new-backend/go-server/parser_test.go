package main

import (
	"testing"

	"github.com/alecthomas/repr"
)

func TestParse(t *testing.T) {
	parser, _ := make_parser()
	terms := parse_fact_string(parser, "#P5 #0 \"This \\\"is\\\" a test\" one \"two\" 0.5 2 -2 1. -1.0 .99999 1.23e8 $ $X % %Z")
	repr.Println(terms, repr.Indent("  "), repr.OmitEmpty(true))
	expected_terms := []Term{
		Term{"id", "P5"},
		Term{"id", "0"},
		Term{"text", "This \"is\" a test"},
		Term{"text", "one"},
		Term{"text", "two"},
		Term{"float", "0.500000"},
		Term{"integer", "2"},
		Term{"integer", "-2"},
		Term{"float", "1.000000"},
		Term{"float", "-1.000000"},
		Term{"float", "0.999990"},
		Term{"float", "123000000.000000"},
		Term{"variable", ""},
		Term{"variable", "X"},
		Term{"postfix", ""},
		Term{"postfix", "Z"},
	}
	if len(terms) != len(expected_terms) {
		t.Error("Wrong number of terms")
	}
	for i, term := range terms {
		if term.Type != expected_terms[i].Type {
			t.Error("Wrong term type for term ", i)
		}
		if term.Value != expected_terms[i].Value {
			t.Error("Wrong term value for term ", i)
		}
	}
}
