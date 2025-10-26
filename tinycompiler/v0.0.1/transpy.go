package main

import (
	"fmt"
	"strings"
)

func transpy(n Function) string {
	funname := n.name
	var funargs string
	if len(n.args) > 0 {
		funargs = n.args[0].name
		for _, arg := range n.args[1:] {
			funargs += ", " + arg.name
		}
	}
	var allocvars string
	for _, va := range n.vars {
		allocvars += fmt.Sprintf("%s = None\n", va.name)
	}
	var nestedfun string
	for _, nf := range n.fun {
		nestedfun += transpy(nf)
	}
	var funbody string
	for _, fb := range n.body {
		funbody += stat(fb)
	}

	str := fmt.Sprintf("def %s(%s):\n", funname, funargs) + indent([]string{allocvars + "\n", nestedfun + "\n", funbody + "\n"})
	if n.name == "main" {
		str += fmt.Sprintf("%s()\n", funname)
	}
	return str
}

func stat(n Statement) string {
	sstostr := func(ss []Statement) string {
		body := []string{}
		for _, s := range ss {
			body = append(body, stat(s))
		}
		if len(body) == 0 {
			body = append(body, "pass")
		}
		return indent(body)
	}

	switch e := n.(type) {
	case Print:
		var sep string
		if e.newline {
			sep = "'\\n'"
		} else {
			sep = "''"
		}
		return fmt.Sprintf("print(%s, end=%s)\n", expr(e.expr), sep)
	case Return:
		if e.expr != nil {
			return fmt.Sprintf("return %s\n", expr(e.expr))
		} else {
			return "return\n"
		}
	case Assign:
		return fmt.Sprintf("%s = %s\n", e.name, expr(e.expr))
	case FunCall:
		return expr(e) + "\n"
	case While:
		return fmt.Sprintf("while %s:\n", expr(e.expr)) + sstostr(e.body)
	case IfThenElse:
		return fmt.Sprintf("if %s:\n%selse:\n%s\n", expr(e.expr), sstostr(e.ibody), sstostr(e.ebody))
	default:
		panic(fmt.Sprint("Unknown statement type", e))
	}
}

func expr(n Expression) string {
	getop := func(eop string) string {
		pyeq := map[string]string{"/": "//", "||": "or", "&&": "and"}
		op, ok := pyeq[eop]
		if ok {
			return op
		} else {
			return eop
		}
	}
	switch e := n.(type) {
	case ArithOp:
		pyop := getop(e.op)
		return fmt.Sprintf("(%s) %s (%s)", expr(e.left), pyop, expr(e.right))
	case LogicOp:
		pyop := getop(e.op)
		return fmt.Sprintf("(%s) %s (%s)", expr(e.left), pyop, expr(e.right))
	case Integer:
		return fmt.Sprintf("%d", e.value)
	case Boolean:
		if e.value {
			return "True"
		} else {
			return "False"
		}
	case String:
		return "'" + e.value + "'"
	case Var:
		return e.name
	case FunCall:
		var args string
		if len(e.args) > 0 {
			args = expr(e.args[0])
			for _, arg := range e.args[1:] {
				args += ", " + expr(arg)
			}
		}
		return fmt.Sprintf("%s(%s)", e.name, args)
	default:
		panic(fmt.Sprint("Unknown expression type", e))
	}
}

func indent(array []string) string {
	var multiline string
	for _, e := range array {
		multiline += e
	}

	lines := strings.Split(multiline, "\n")
	if lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	var res string
	for _, line := range lines {
		res += "\t" + line + "\n"
	}
	return res
}
