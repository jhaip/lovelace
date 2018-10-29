package main

import (
	"fmt"
	"strings"
	"testing"
	"time"

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

func TestParse2(t *testing.T) {
	msg := "#1800 \"This \\\"is\\\" a test\" one \"two\" (0.5, 2) @ $X $x2 $ % %Z true false null"
	start := time.Now()
	terms, err := ParseReader("", strings.NewReader(msg))
	timeProcessing := time.Since(start)
	fmt.Printf("processing: %s \n", timeProcessing)
	if err != nil {
		t.Error(err)
	}
	repr.Println(terms, repr.Indent("  "), repr.OmitEmpty(true))
	expected_terms := []Term{
		Term{"id", "1800"},
		Term{"text", "This \"is\" a test"},
		Term{"text", "one"},
		Term{"text", "two"},
		Term{"text", "("},
		Term{"float", "0.500000"},
		Term{"text", ","},
		Term{"integer", "2"},
		Term{"text", ")"},
		Term{"text", "@"},
		Term{"variable", "X"},
		Term{"variable", "x2"},
		Term{"variable", ""},
		Term{"postfix", ""},
		Term{"postfix", "Z"},
		Term{"text", "true"},
		Term{"text", "false"},
		Term{"text", "null"},
	}
	checkTerms(terms.([]Term), expected_terms, t)
}

func TestParse2Numbers(t *testing.T) {
	msg := "0.5 2 -2 1.0 -1.0 0.99999 1.23e8"
	raw_terms, err := ParseReader("", strings.NewReader(msg))
	terms, _ := raw_terms.([]Term)
	if err != nil {
		t.Error(err)
	}
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

func TestParse2Variables(t *testing.T) {
	msg := "$ $X $Y $ one $cat"
	raw_terms, err := ParseReader("", strings.NewReader(msg))
	terms, _ := raw_terms.([]Term)
	if err != nil {
		t.Error(err)
	}
	repr.Println(terms, repr.Indent("  "), repr.OmitEmpty(true))
	expected_terms := []Term{
		Term{"variable", ""},
		Term{"variable", "X"},
		Term{"variable", "Y"},
		Term{"variable", ""},
		Term{"text", "one"},
		Term{"variable", "cat"},
	}
	checkTerms(terms, expected_terms, t)
}
