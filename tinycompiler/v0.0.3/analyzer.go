package main

import (
	"fmt"
)

var counter = 0

func newLabel() string {
	counter++
	return fmt.Sprintf("uniqstr%d", counter)
}

func buildSymtable(ast any) {
	fun, ok := ast.(Function)
	if !ok || fun.name != "main" || fun.deco["type"] != VOID || len(fun.args) > 0 {
		panic("Cannot find a valid entry point")
	}
	symtable := &SymbolTable{}
	symtable.addFun(fun.name, []Type{}, fun.deco)
	fun.deco["label"] = fun.name + "_" + newLabel() // unique label
	processScope(&fun, symtable)
}

func processScope(fun *Function, symtable *SymbolTable) {
	fun.deco["nonlocal"] = make(map[string]bool) // set of nonlocal variable names in the function body, used in "readable" python transpilation only
	symtable.pushScope(fun.deco)

	for _, v := range fun.args { // process function arguments
		symtable.addVar(v.name, v.deco)
	}

	for _, v := range fun.vars { // process local variables
		symtable.addVar(v.name, v.deco)
	}

	for _, f := range fun.fun { // process nested functions: first add function symbols to the table
		argtypes := []Type{}
		for _, arg := range f.args {
			argtypes = append(argtypes, arg.deco["type"].(Type))
		}
		symtable.addFun(f.name, argtypes, f.deco)
		f.deco["label"] = f.name + "_" + newLabel() // still need unique labels
	}

	for i := range fun.fun { // then process nested function bodies
		processScope(&fun.fun[i], symtable)
	}

	for i := range fun.body { // process the list of statements
		fun.body[i] = processStat(fun.body[i], symtable)
	}

	symtable.popScope()
}

// process "statement" syntax tree nodes
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
		if symtable.retStack[len(symtable.retStack)-1]["type"] != e.expr.getDeco()["type"] {
			panic(fmt.Sprintf("Incompatible types in return statement, line %s", e.deco["lineno"]))
		}
		return e
	case Assign:
		e.expr = processExpr(e.expr, symtable)
		deco := symtable.findVar(e.name) // used in "readable" python transpilation only
		if len(e.deco) == 0 {
			e.deco = make(map[string]any)
		}
		e.deco["type"] = deco["type"]
		if e.deco["type"] != e.expr.getDeco()["type"] {
			panic(fmt.Sprintf("Incompatible types in assignment statement, line %s", e.deco["lineno"]))
		}
		updateNonlocals(e.name, symtable) // used in "readable" python transpilation only
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
		e.deco["type"] = INT
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
		e.deco["type"] = BOOL
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
	case Integer: // no type checking is necessary
		if len(e.deco) == 0 {
			e.deco = make(map[string]any)
		}
		e.deco["type"] = INT
		return e
	case Boolean: // no type checking is necessary
		if len(e.deco) == 0 {
			e.deco = make(map[string]any)
		}
		e.deco["type"] = BOOL
		return e
	case Var: // no type checking is necessary
		deco := symtable.findVar(e.name)
		if len(e.deco) == 0 {
			e.deco = make(map[string]any)
		}
		e.deco["type"] = deco["type"]
		updateNonlocals(e.name, symtable) // used in "readable" python transpilation only
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
		e.deco["fundeco"] = deco // save the function symbol, useful for overloading and for stack preparation
		e.deco["type"] = deco["type"]
		return e
	case String:
		if len(e.deco) == 0 {
			e.deco = make(map[string]any)
		}
		e.deco["type"] = STRING
		return e
	default:
		panic(fmt.Sprintln("Unknown expression type", e))
	}
}

// add the variable name to the set of nonlocals
func updateNonlocals(name string, symtable *SymbolTable) {
	// for all the enclosing scopes until we find the instance
	for i := len(symtable.variables) - 1; i >= 0; i-- {
		// used in "readable" python transpilation only
		if _, ok := symtable.variables[i][name]; ok {
			break
		}
		if retStack, ok := symtable.retStack[i]["nonlocal"].(map[string]bool); ok {
			retStack[name] = true
		}
	}
}
