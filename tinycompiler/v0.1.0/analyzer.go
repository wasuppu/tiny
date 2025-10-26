package main

import (
	"fmt"
)

var counter = 0

func curLabel() string {
	return fmt.Sprintf("uniqstr%d", counter)
}

func newLabel() string {
	counter++
	return fmt.Sprintf("uniqstr%d", counter)
}

func buildSymtable(ast any) {
	fun, ok := ast.(Function)
	if !ok || fun.name != "main" || fun.deco["type"] != VOID || len(fun.args) > 0 {
		panic("Cannot find a valid entry point")
	}
	symtable := newSymbolTable()
	symtable.addFun(fun.name, []Type{}, fun.deco)
	fun.deco["strings"] = make(map[string]string)
	processScope(&fun, symtable)
	processScope(&fun, symtable)
	fun.deco["scopeCnt"] = symtable.scopeCnt
}

func processScope(fun *Function, symtable *SymbolTable) {
	symtable.pushScope(&fun.deco)

	for _, v := range fun.args { // process function arguments
		symtable.addVar(v.name, &v.deco)
	}

	for _, v := range fun.vars { // process local variables
		symtable.addVar(v.name, &v.deco)
	}

	for _, f := range fun.fun { // process nested functions: first add function symbols to the table
		argtypes := []Type{}
		for _, arg := range f.args {
			argtypes = append(argtypes, arg.deco["type"].(Type))
		}
		symtable.addFun(f.name, argtypes, f.deco)
	}

	for i := range fun.fun { // then process nested function bodies
		processScope(&fun.fun[i], symtable)
	}

	for i := range fun.body { // process the list of statements
		fun.body[i] = processStat(fun.body[i], symtable)
	}

	symtable.popScope()
}

func processStat(n Statement, symtable *SymbolTable) Statement {
	switch e := n.(type) {
	case Print:
		e.expr = processExpr(e.expr, symtable)
		return e
	case Return:
		if e.expr == nil {
			return nil
		}
		e.expr = processExpr(e.expr, symtable)
		if (*symtable.retStack[len(symtable.retStack)-1])["type"] != e.expr.getDeco()["type"] {
			panic(fmt.Sprintf("Incompatible types in return statement, line %s", e.deco["lineno"]))
		}
		return e
	case Assign:
		e.expr = processExpr(e.expr, symtable)
		deco := symtable.findVar(e.name)
		if len(e.deco) == 0 {
			e.deco = make(map[string]any)
		}
		for k := range deco {
			e.deco[k] = deco[k]
		}

		if e.deco["type"] != e.expr.getDeco()["type"] {
			panic(fmt.Sprintf("Incompatible types in assignment statement, line %s", e.deco["lineno"]))
		}
		return e
	case FunCall: // no type checking is necessary
		e = processExpr(e, symtable).(FunCall)
		return e
	case While:
		e.expr = processExpr(e.expr, symtable)
		if e.expr.getDeco()["type"].(Type) != BOOL {
			panic(fmt.Sprintf("Non-boolean expression in while statement, line %s", e.deco["lineno"]))
		}
		for i := range e.body {
			e.body[i] = processStat(e.body[i], symtable)
		}
		return e
	case IfThenElse:
		e.expr = processExpr(e.expr, symtable)
		if e.expr.getDeco()["type"].(Type) != BOOL {
			panic(fmt.Sprintf("Non-boolean expression in if statement, line %s", e.deco["lineno"]))
		}
		ss := append(e.ibody, e.ebody...)
		for i := range ss {
			ss[i] = processStat(ss[i], symtable)
		}
		return e
	default:
		panic(fmt.Sprintln("Unknown statement type", e))
	}
}

func processExpr(n Expression, symtable *SymbolTable) Expression {
	switch e := n.(type) {
	case ArithOp:
		if len(e.deco) == 0 {
			e.deco = make(map[string]any)
		}
		e.left = processExpr(e.left, symtable)
		e.right = processExpr(e.right, symtable)
		if e.left.getDeco()["type"].(Type) != INT || e.right.getDeco()["type"].(Type) != INT {
			panic(fmt.Sprintf("Arithmetic operation over non-integer type in line %s", e.deco["lineno"]))
		}
		return e
	case LogicOp:
		if len(e.deco) == 0 {
			e.deco = make(map[string]any)
		}
		e.left = processExpr(e.left, symtable)
		e.right = processExpr(e.right, symtable)
		if e.left.getDeco()["type"].(Type) != e.right.getDeco()["type"] {
			panic(fmt.Sprintf("Arithmetic operation over non-integer type in line %s", e.deco["lineno"]))
		}
		switch e.op {
		case "<=", "<", ">=", ">":
			if e.left.getDeco()["type"].(Type) != INT {
				panic(fmt.Sprintf("Arithmetic operation over non-integer type in line %#v", e.deco["lineno"]))
			}
		case "&&", "||":
			if e.left.getDeco()["type"].(Type) != BOOL {
				panic(fmt.Sprintf("Arithmetic operation over non-integer type in line %s", e.deco["lineno"]))
			}
		}
		return e
	case Var: // no type checking is necessary
		deco := symtable.findVar(e.name)
		if len(e.deco) == 0 {
			e.deco = make(map[string]any)
		}

		for k := range deco {
			e.deco[k] = deco[k]
		}
		return e
	case FunCall:
		for i := range e.args {
			e.args[i] = processExpr(e.args[i], symtable)
		}
		argtypes := []Type{}
		for _, arg := range e.args {
			argtypes = append(argtypes, arg.getDeco()["type"].(Type))
		}
		deco := symtable.findFun(e.name, argtypes)
		if len(e.deco) == 0 {
			e.deco = make(map[string]any)
		}

		for k := range deco {
			e.deco[k] = deco[k]
		}
		return e
	case String:
		if _, ok := (*symtable.retStack[1])["strings"]; !ok {
			(*symtable.retStack[1])["strings"] = make(map[string]string)
		}
		(*symtable.retStack[1])["strings"].(map[string]string)[e.deco["label"].(string)] = e.value
		return e
	case Integer: // no type checking is necessary
		return e
	case Boolean: // no type checking is necessary
		return e
	default:
		panic(fmt.Sprintln("Unknown expression type", e))
	}
}
