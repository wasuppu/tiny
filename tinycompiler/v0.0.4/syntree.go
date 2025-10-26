package main

type Type int

const (
	VOID Type = iota
	INT
	BOOL
	STRING
)

var TypeNames = [...]string{"VOID", "INT", "BOOL", "STRING"}

type Function struct {
	name string         // function name, string
	args []Var          // function arguments, list of tuples (name, type)
	vars []Var          // local variables, list of tuples (name, type)
	fun  []Function     // nested functions, list of Function nodes
	body []Statement    // function body, list of statement nodes (Print/Return/Assign/While/IfThenElse/FunCall)
	deco map[string]any // decoration dictionary to be filled by the parser (line number) and by the semantic analyzer (return type, scope id etc)
}

// statements
type Statement interface {
	s()
}

type Print struct {
	expr    Expression
	newline bool
	deco    map[string]any
}

func (s Print) s() {}

type Return struct {
	expr Expression
	deco map[string]any
}

func (s Return) s() {}

type Assign struct {
	name string
	expr Expression
	deco map[string]any
}

func (s Assign) s() {}

type While struct {
	expr Expression
	body []Statement
	deco map[string]any
}

func (s While) s() {}

type IfThenElse struct {
	expr  Expression
	ibody []Statement
	ebody []Statement
	deco  map[string]any
}

func (s IfThenElse) s() {}

// expressions
type Expression interface {
	e()
	getDeco() map[string]any
}

type ArithOp struct {
	op    string
	left  Expression
	right Expression
	deco  map[string]any
}

func (e ArithOp) e()                      {}
func (e ArithOp) getDeco() map[string]any { return e.deco }

type LogicOp struct {
	op    string
	left  Expression
	right Expression
	deco  map[string]any
}

func (e LogicOp) e()                      {}
func (e LogicOp) getDeco() map[string]any { return e.deco }

type Integer struct {
	value int
	deco  map[string]any
}

func (e Integer) e()                      {}
func (e Integer) getDeco() map[string]any { return e.deco }

type Boolean struct {
	value bool
	deco  map[string]any
}

func (e Boolean) e()                      {}
func (e Boolean) getDeco() map[string]any { return e.deco }

type String struct {
	value string
	deco  map[string]any
}

func (e String) e()                      {}
func (e String) getDeco() map[string]any { return e.deco }

type Var struct {
	name string
	deco map[string]any
}

func (e Var) e()                      {}
func (e Var) getDeco() map[string]any { return e.deco }

// depending on the context, a function call can be a statement or an expression
type FunCall struct {
	name string
	args []Expression
	deco map[string]any
}

func (e FunCall) s()                      {}
func (e FunCall) e()                      {}
func (e FunCall) getDeco() map[string]any { return e.deco }
