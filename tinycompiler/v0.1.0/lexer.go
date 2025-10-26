package main

import (
	"fmt"
	"unicode"
)

var (
	Keywords   = map[string]string{"true": "BOOLEAN", "false": "BOOLEAN", "print": "PRINT", "println": "PRINT", "int": "TYPE", "bool": "TYPE", "if": "IF", "else": "ELSE", "while": "WHILE", "return": "RETURN"}
	DoubleChar = map[string]string{"==": "COMP", "<=": "COMP", ">=": "COMP", "!=": "COMP", "&&": "AND", "||": "OR"}
	SingleChar = map[string]string{"=": "ASSIGN", "<": "COMP", ">": "COMP", "!": "NOT", "+": "PLUS", "-": "MINUS", "/": "DIVIDE", "*": "TIMES", "%": "MOD", "(": "LPAREN", ")": "RPAREN", "{": "BEGIN", "}": "END", ";": "SEMICOLON", ",": "COMMA", ":": "COLON"}
	Tokens     = map[string]bool{"ID": true, "STRING": true, "INTEGER": true}
)

func init() {
	for _, v := range Keywords {
		Tokens[v] = true
	}
	for _, v := range DoubleChar {
		Tokens[v] = true
	}
	for _, v := range SingleChar {
		Tokens[v] = true
	}
}

type Token struct {
	typ    string
	value  string
	lineno int
}

func (t Token) isInitialized() bool {
	return t != Token{}
}

func (t Token) String() string {
	return fmt.Sprintf("Token(type=%s, value=%s, lineno=%d)", t.typ, t.value, t.lineno)
}

func tokenize(text string) []Token {
	tokens := []Token{}
	lineno, idx, state, accum := 0, 0, 0, ""

	for idx < len(text) {
		sym1 := byte(' ') // current symbol
		if idx < len(text) {
			sym1 = text[idx]
		}
		sym2 := byte(' ') // next symbol
		if idx < len(text)-1 {
			sym2 = text[idx+1]
		}

		switch state {
		case 0: // start scanning a new token
			if sym1 == '/' && sym2 == '/' { // start a comment scan
				state = 1
			} else if unicode.IsDigit(rune(sym1)) { // start a number scan
				state = 2
				accum += string(sym1)
			} else if sym1 == '"' { // start a string scan
				state = 3
			} else if unicode.IsLetter(rune(sym1)) || sym1 == '_' { // start a word scan
				state = 4
				accum += string(sym1)
			} else if typ, ok := DoubleChar[string(sym1)+string(sym2)]; ok { // emit two-character token
				tokens = append(tokens, Token{typ, string(sym1) + string(sym2), lineno})
				idx++
			} else if typ, ok := SingleChar[string(sym1)]; ok { // emit one-character token
				tokens = append(tokens, Token{typ, string(sym1), lineno})
			} else if sym1 != '\r' && sym1 != '\t' && sym1 != ' ' && sym1 != '\n' { // ignore whitespace
				panic(fmt.Sprintf("Lexical error: illegal character %q at line %d", sym1, lineno))
			}
		case 2: // scanning a number
			if unicode.IsDigit(rune(sym1)) { // is next character a digit?
				accum += string(sym1) // if yes, continue
			} else {
				tokens = append(tokens, Token{"INTEGER", accum, lineno}) // otherwise, emit number token
				idx--
				state, accum = 0, "" // start new scan
			}
		case 3: // scanning a string, check next character
			if sym1 != '"' || accum != "" && accum[len(accum)-1] == '\\' { // if not quote mark (or if escaped quote mark),
				accum += string(sym1) // continue the scan
			} else {
				tokens = append(tokens, Token{"STRING", accum, lineno}) // otherwise, emit number token
				state, accum = 0, ""                                    // start new scan
			}
		case 4: // scanning a word, check next character
			if unicode.IsLetter(rune(sym1)) || sym1 == '_' || unicode.IsDigit(rune(sym1)) { // still word?
				accum += string(sym1) //  if yes, continue
			} else { // otherwise the scan stops, we have a word
				typ, ok := Keywords[accum]
				if ok { // is the word reserved?
					tokens = append(tokens, Token{typ, accum, lineno}) // if yes, keyword
				} else {
					tokens = append(tokens, Token{"ID", accum, lineno}) // identifier otherwise
				}
				idx--
				state, accum = 0, "" // start new scan
			}
		}

		if sym1 == '\n' {
			lineno++
			if state == 1 { // if comment, start new scan
				state, accum = 0, ""
			}
		}
		idx++
	}

	if state != 0 {
		fmt.Println(state, accum)
		panic("Lexical error: unexpected EOF")
	}
	return tokens
}
