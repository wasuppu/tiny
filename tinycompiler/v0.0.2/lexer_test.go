package main

import (
	"fmt"
	"testing"
)

func TestLexer(t *testing.T) {
	tests := []struct {
		program        string
		expectedTokens []Token
	}{
		{"{}", []Token{
			{typ: "BEGIN", value: "{"},
			{typ: "END", value: "}"},
		}},
		{"whilea=0;", []Token{
			{typ: "ID", value: "whilea"},
			{typ: "ASSIGN", value: "="},
			{typ: "INTEGER", value: "0"},
			{typ: "SEMICOLON", value: ";"},
		}},

		{"while=0;", []Token{
			{typ: "WHILE", value: "while"},
			{typ: "ASSIGN", value: "="},
			{typ: "INTEGER", value: "0"},
			{typ: "SEMICOLON", value: ";"},
		}},
		{`{
            println(02); // no direct conversion to integer values
            println "hello world";
            // println(3);
        }`, []Token{
			{typ: "BEGIN", value: "{"},
			{typ: "PRINT", value: "println"},
			{typ: "LPAREN", value: "("},
			{typ: "INTEGER", value: "02"},
			{typ: "RPAREN", value: ")"},
			{typ: "SEMICOLON", value: ";"},
			{typ: "PRINT", value: "println"},
			{typ: "STRING", value: "hello world"},
			{typ: "SEMICOLON", value: ";"},
			{typ: "END", value: "}"},
		}},

		{"a = 2;", []Token{
			{typ: "ID", value: "a"},
			{typ: "ASSIGN", value: "="},
			{typ: "INTEGER", value: "2"},
			{typ: "SEMICOLON", value: ";"},
		}},

		{"(-2+3*4)+5/(7-6)%8;", []Token{
			{typ: "LPAREN", value: "("},
			{typ: "MINUS", value: "-"},
			{typ: "INTEGER", value: "2"},
			{typ: "PLUS", value: "+"},
			{typ: "INTEGER", value: "3"},
			{typ: "TIMES", value: "*"},
			{typ: "INTEGER", value: "4"},
			{typ: "RPAREN", value: ")"},
			{typ: "PLUS", value: "+"},
			{typ: "INTEGER", value: "5"},
			{typ: "DIVIDE", value: "/"},
			{typ: "LPAREN", value: "("},
			{typ: "INTEGER", value: "7"},
			{typ: "MINUS", value: "-"},
			{typ: "INTEGER", value: "6"},
			{typ: "RPAREN", value: ")"},
			{typ: "MOD", value: "%"},
			{typ: "INTEGER", value: "8"},
			{typ: "SEMICOLON", value: ";"},
		}},
	}

	for _, tt := range tests {
		tokens := tokenize(tt.program)

		if len(tokens) != len(tt.expectedTokens) {
			fmt.Println(tokens)
			t.Fatalf("expected tokens length: %d, actual length: %d", len(tt.expectedTokens), len(tokens))
		}

		for i := range tokens {
			if tokens[i].typ != tt.expectedTokens[i].typ {

				t.Fatalf("expected tokens type: %s, actual type: %s", tt.expectedTokens[i].typ, tokens[i].typ)
			}
			if tokens[i].value != tt.expectedTokens[i].value {
				t.Fatalf("expected tokens value: %s, actual value: %s", tt.expectedTokens[i].value, tokens[i].value)
			}
		}
	}
}
