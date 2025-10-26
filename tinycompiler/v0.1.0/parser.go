package main

import (
	"fmt"
	"strconv"
)

type Grammer struct {
	nonterminal string
	production  []string
	constructor func([]any) any
}

var Grammars = []Grammer{
	{
		"fun",
		[]string{"fun_type", "ID", "LPAREN", "param_list", "RPAREN", "BEGIN", "var_list", "fun_list", "statement_list", "END"},
		func(p []any) any {
			return Function{
				p[1].(Token).value,
				p[3].([]Var),
				p[6].([]Var),
				p[7].([]Function),
				p[8].([]Statement),
				map[string]any{"type": p[0], "lineno": p[1].(Token).lineno, "label": p[1].(Token).value + "_" + newLabel()}}
		},
	},
	{
		"var",
		[]string{"TYPE", "ID"},
		func(p []any) any {
			typ := BOOL
			if p[0].(Token).value == "int" {
				typ = INT
			}
			return Var{
				p[1].(Token).value,
				map[string]any{"type": typ, "lineno": p[0].(Token).lineno},
			}
		},
	},
	{
		"param_list",
		[]string{"var"},
		func(p []any) any {
			r := make([]Var, len(p))
			for i := range p {
				r[i] = p[i].(Var)
			}
			return r
		},
	},
	{
		"param_list",
		[]string{},
		func(p []any) any {
			r := make([]Var, len(p))
			for i := range p {
				r[i] = p[i].(Var)
			}
			return r
		},
	},
	{
		"param_list",
		[]string{"param_list", "COMMA", "var"},
		func(p []any) any {
			return append(p[0].([]Var), p[2].(Var))
		},
	},
	{
		"fun_type",
		[]string{"TYPE"},
		func(p []any) any {
			typ := BOOL
			if p[0].(Token).value == "int" {
				typ = INT
			}
			return typ
		},
	},
	{
		"fun_type",
		[]string{},
		func(p []any) any {
			return VOID
		},
	},
	{
		"var_list",
		[]string{"var_list", "var", "SEMICOLON"},
		func(p []any) any {
			return append(p[0].([]Var), p[1].(Var))
		},
	},
	{
		"var_list",
		[]string{},
		func(p []any) any {
			r := make([]Var, len(p))
			for i := range p {
				r[i] = p[i].(Var)
			}
			return r
		},
	},
	{
		"fun_list",
		[]string{"fun_list", "fun"},
		func(p []any) any {
			return append(p[0].([]Function), p[1].(Function))
		},
	},
	{
		"fun_list",
		[]string{},
		func(p []any) any {
			r := make([]Function, len(p))
			for i := range p {
				r[i] = p[i].(Function)
			}
			return r
		},
	},
	{
		"statement_list",
		[]string{"statement_list", "statement"},
		func(p []any) any {
			return append(p[0].(([]Statement)), p[1].(Statement))
		},
	},
	{
		"statement_list",
		[]string{},
		func(p []any) any {
			r := make([]Statement, len(p))
			for i := range p {
				r[i] = p[i].(Statement)
			}
			return r
		},
	},
	{
		"statement",
		[]string{"ID", "LPAREN", "arg_list", "RPAREN", "SEMICOLON"},
		func(p []any) any {
			return FunCall{
				p[0].(Token).value,
				p[2].([]Expression),
				map[string]any{"lineno": p[0].(Token).lineno},
			}
		},
	},
	{
		"statement",
		[]string{"ID", "ASSIGN", "expr", "SEMICOLON"},
		func(p []any) any {
			return Assign{
				p[0].(Token).value,
				p[2].(Expression),
				map[string]any{"lineno": p[0].(Token).lineno},
			}
		},
	},
	{
		"statement",
		[]string{"RETURN", "expr", "SEMICOLON"},
		func(p []any) any {
			return Return{
				p[1].(Expression),
				map[string]any{"lineno": p[0].(Token).lineno},
			}
		},
	},
	{
		"statement",
		[]string{"RETURN", "SEMICOLON"},
		func(p []any) any {
			return Return{
				nil,
				map[string]any{"lineno": p[0].(Token).lineno},
			}
		},
	},
	{
		"statement",
		[]string{"PRINT", "expr", "SEMICOLON"},
		func(p []any) any {
			return Print{
				p[1].(Expression),
				p[0].(Token).value == "println",
				map[string]any{"lineno": p[0].(Token).lineno},
			}
		},
	},
	{
		"statement",
		[]string{"IF", "expr", "BEGIN", "statement_list", "END", "else_statement"},
		func(p []any) any {
			return IfThenElse{
				p[1].(Expression),
				p[3].([]Statement),
				p[5].([]Statement),
				map[string]any{"lineno": p[0].(Token).lineno},
			}
		},
	},
	{
		"else_statement",
		[]string{"ELSE", "BEGIN", "statement_list", "END"},
		func(p []any) any {
			return p[2].([]Statement)
		},
	},
	{
		"else_statement",
		[]string{},
		func(p []any) any {
			r := make([]Statement, len(p))
			for i := range p {
				r[i] = p[i].(Statement)
			}
			return r
		},
	},
	{
		"statement",
		[]string{"WHILE", "expr", "BEGIN", "statement_list", "END"},
		func(p []any) any {
			return While{
				p[1].(Expression),
				p[3].([]Statement),
				map[string]any{"lineno": p[0].(Token).lineno},
			}
		},
	},
	{
		"arg_list",
		[]string{"expr"},
		func(p []any) any {
			r := make([]Expression, len(p))
			for i := range p {
				r[i] = p[i].(Expression)
			}
			return r
		},
	},
	{
		"arg_list",
		[]string{"arg_list", "COMMA", "expr"},
		func(p []any) any {
			return append(p[0].([]Expression), p[2].(Expression))
		},
	},
	{
		"arg_list",
		[]string{},
		func(p []any) any {
			r := make([]Expression, len(p))
			for i := range p {
				r[i] = p[i].(Expression)
			}
			return r
		},
	},
	{
		"expr",
		[]string{"conjunction"},
		func(p []any) any {
			return p[0].(Expression)
		},
	},
	{
		"expr",
		[]string{"expr", "OR", "conjunction"},
		func(p []any) any {
			return LogicOp{
				p[1].(Token).value,
				p[0].(Expression),
				p[2].(Expression),
				map[string]any{"lineno": p[1].(Token).lineno, "type": BOOL},
			}
		},
	},
	{
		"expr",
		[]string{"STRING"},
		func(p []any) any {
			return String{
				p[0].(Token).value,
				map[string]any{"lineno": p[0].(Token).lineno, "type": STRING, "label": newLabel()},
			}
		},
	},
	{
		"conjunction",
		[]string{"literal"},
		func(p []any) any {
			return p[0].(Expression)
		},
	},
	{
		"conjunction",
		[]string{"conjunction", "AND", "literal"},
		func(p []any) any {
			return LogicOp{
				p[1].(Token).value,
				p[0].(Expression),
				p[2].(Expression),
				map[string]any{"lineno": p[1].(Token).lineno, "type": BOOL},
			}
		},
	},
	{
		"literal",
		[]string{"comparand"},
		func(p []any) any {
			return p[0].(Expression)
		},
	},
	{
		"literal",
		[]string{"NOT", "comparand"},
		func(p []any) any {
			return LogicOp{
				"==",
				Boolean{
					false,
					map[string]any{"type": BOOL},
				},
				p[1].(Expression),
				map[string]any{"lineno": p[0].(Token).lineno, "type": BOOL},
			}
		},
	},
	{
		"comparand",
		[]string{"addend"},
		func(p []any) any {
			return p[0].(Expression)
		},
	},
	{
		"comparand",
		[]string{"addend", "COMP", "addend"},
		func(p []any) any {
			return LogicOp{
				p[1].(Token).value,
				p[0].(Expression),
				p[2].(Expression),
				map[string]any{"lineno": p[1].(Token).lineno, "type": BOOL},
			}
		},
	},
	{
		"addend",
		[]string{"term"},
		func(p []any) any {
			return p[0].(Expression)
		},
	},
	{
		"addend",
		[]string{"addend", "MINUS", "term"},
		func(p []any) any {
			return ArithOp{
				p[1].(Token).value,
				p[0].(Expression),
				p[2].(Expression),
				map[string]any{"lineno": p[1].(Token).lineno, "type": INT},
			}
		},
	},
	{
		"addend",
		[]string{"addend", "PLUS", "term"},
		func(p []any) any {
			return ArithOp{
				p[1].(Token).value,
				p[0].(Expression),
				p[2].(Expression),
				map[string]any{"lineno": p[1].(Token).lineno, "type": INT},
			}
		},
	},
	{
		"term",
		[]string{"factor"},
		func(p []any) any {
			return p[0].(Expression)
		},
	},
	{
		"term",
		[]string{"term", "MOD", "factor"},
		func(p []any) any {
			return ArithOp{
				p[1].(Token).value,
				p[0].(Expression),
				p[2].(Expression),
				map[string]any{"lineno": p[1].(Token).lineno, "type": INT},
			}
		},
	},
	{
		"term",
		[]string{"term", "DIVIDE", "factor"},
		func(p []any) any {
			return ArithOp{
				p[1].(Token).value,
				p[0].(Expression),
				p[2].(Expression),
				map[string]any{"lineno": p[1].(Token).lineno, "type": INT},
			}
		},
	},
	{
		"term",
		[]string{"term", "TIMES", "factor"},
		func(p []any) any {
			return ArithOp{
				p[1].(Token).value,
				p[0].(Expression),
				p[2].(Expression),
				map[string]any{"lineno": p[1].(Token).lineno, "type": INT},
			}
		},
	},
	{
		"factor",
		[]string{"atom"},
		func(p []any) any {
			return p[0].(Expression)
		},
	},
	{
		"factor",
		[]string{"PLUS", "atom"},
		func(p []any) any {
			return p[1].(Expression)
		},
	},
	{
		"factor",
		[]string{"MINUS", "atom"},
		func(p []any) any {
			return ArithOp{
				"-",
				Integer{
					0,
					map[string]any{"type": INT},
				},
				p[1].(Expression),
				map[string]any{"lineno": p[0].(Token).lineno, "type": INT},
			}
		},
	},
	{
		"atom",
		[]string{"BOOLEAN"},
		func(p []any) any {
			return Boolean{
				p[0].(Token).value == "true",
				map[string]any{"lineno": p[0].(Token).lineno, "type": BOOL},
			}
		},
	},
	{
		"atom",
		[]string{"INTEGER"},
		func(p []any) any {
			n, _ := strconv.Atoi(p[0].(Token).value)
			return Integer{
				n,
				map[string]any{"lineno": p[0].(Token).lineno, "type": INT},
			}
		},
	},
	{
		"atom",
		[]string{"ID", "LPAREN", "arg_list", "RPAREN"},
		func(p []any) any {
			return FunCall{
				p[0].(Token).value,
				p[2].([]Expression),
				map[string]any{"lineno": p[0].(Token).lineno},
			}
		},
	},
	{
		"atom",
		[]string{"ID"},
		func(p []any) any {
			return Var{
				p[0].(Token).value,
				map[string]any{"lineno": p[0].(Token).lineno},
			}
		},
	},
	{
		"atom",
		[]string{"LPAREN", "expr", "RPAREN"},
		func(p []any) any {
			return p[1].(Expression)
		},
	},
}

type ParseState struct {
	rule  int // index of the parse rule in the grammar
	dot   int // index of next symbol in the rule (dot position)
	start int // we saw this many tokens when we started the rule
	// these two members are not necessary for the recogninzer, but are handy to retrieve a parse path
	token int         //  we saw this many tokens up to the current dot position.
	prev  *ParseState // parent parse state pointer
}

func (t ParseState) isInitialized() bool {
	return t != ParseState{}
}

func (s ParseState) nextSymbol() string {
	prod := Grammars[s.rule].production
	if s.dot < len(prod) {
		return prod[s.dot]
	} else {
		return ""
	}
}

func (s ParseState) eq(o ParseState) bool {
	return s.rule == o.rule && s.dot == o.dot && s.start == o.start // NB no self.token, no self.prev
}

type WendParser struct {
	seen []Token
}

func (p *WendParser) recognize(tokens []Token) ParseState {
	charts := [][]ParseState{{{rule: 0, dot: 0, start: 0}}}

	appendState := func(i int, state ParseState) {
		if len(charts) == i {
			charts = append(charts, []ParseState{})
		}
		in := false
		for _, s := range charts[i] {
			if state.eq(s) {
				in = true
			}
		}
		if !in {
			charts[i] = append(charts[i], state)
		}
	}

	idx := 0
	for (len(p.seen) == 0 || p.seen[len(p.seen)-1].isInitialized()) && (idx < len(tokens)) {
		idx++
		p.seen = append(p.seen, tokens[len(p.seen)])
		pos := len(p.seen) - 1
		i := 0
		for i < len(charts[pos]) {
			state := charts[pos][i]
			symbol := state.nextSymbol()
			if symbol == "" {
				for _, item := range charts[state.start] {
					if item.nextSymbol() == Grammars[state.rule].nonterminal {
						appendState(pos, ParseState{item.rule, item.dot + 1, item.start, pos, &state})
					}
				}

			} else if Tokens[symbol] {
				if p.seen[len(p.seen)-1].isInitialized() && symbol == p.seen[len(p.seen)-1].typ {
					appendState(pos+1, ParseState{rule: state.rule, dot: state.dot + 1, start: state.start, token: pos + 1, prev: &state})
				}
			} else {
				for i, rule := range Grammars {
					if rule.nonterminal == symbol {
						appendState(pos, ParseState{rule: i, dot: 0, start: pos, token: pos, prev: &state})
					}
				}
			}
			i++
		}

		if p.seen[len(p.seen)-1].isInitialized() && len(charts) == pos+1 {
			panic(fmt.Sprintf("Syntax error at line %d, token=%s", p.seen[pos].lineno, p.seen[pos].typ))
		}
	}

	cur := []ParseState{}
	for _, state := range charts[len(charts)-1] {
		if state.rule == 0 && state.dot == len(Grammars[0].production) && state.start == 0 {
			cur = append(cur, state)
		}
	}
	if len(cur) == 0 {
		panic("Syntax error: unexpected EOF")
	}
	return cur[0]
}

func (p *WendParser) buildSyntree(rule ParseState) any {
	production := []ParseState{}
	for rule.isInitialized() {
		if rule.nextSymbol() == "" {
			production = append(production, rule)
		}
		rule = *rule.prev
	}

	stack := []any{}
	token := 0
	for i := len(production) - 1; i >= 0; i-- {
		rule := production[i]
		for _, t := range p.seen[token:rule.token] {
			stack = append(stack, t)
		}
		token = rule.token
		chomp := len(Grammars[rule.rule].production)
		chew := []any{}
		if chomp > 0 {
			chew = stack[len(stack)-chomp:]
			stack = stack[:len(stack)-chomp]
		}
		stack = append(stack, Grammars[rule.rule].constructor(chew))
	}
	return stack[0]
}

func (p *WendParser) Parse(tokens []Token) any {
	return p.buildSyntree(p.recognize(tokens))
}
