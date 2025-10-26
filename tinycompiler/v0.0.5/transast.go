package main

import (
	"fmt"
)

func varstostr(es []Var) string {
	var args string
	if len(es) > 0 {
		args = exprast(es[0])
		for _, arg := range es[1:] {
			args += ",\n" + exprast(arg)
		}
	}
	return args
}

func estostr(es []Expression) string {
	var args string
	if len(es) > 0 {
		args = exprast(es[0])
		for _, arg := range es[1:] {
			args += ",\n" + exprast(arg)
		}
	}
	return args
}

func sstostr(ss []Statement) string {
	body := ""
	if len(ss) > 0 {
		body = statast(ss[0])
		for _, s := range ss[1:] {
			body += ",\n" + statast(s)
		}
	}
	return body
}

func funstostr(ss []Function) string {
	body := ""
	if len(ss) > 0 {
		body = transast(ss[0])
		for _, s := range ss[1:] {
			body += ",\n" + transast(s)
		}
	}
	return body
}

func transast(n Function) string {
	str := fmt.Sprintf(`Function{
        name: %s,
        args: []Var{%s},
        vars: []Var{%s}}
        fun: []Function{%s},
        body: []Statement{
        %s},
        deco: %s`, n.name, varstostr(n.args), varstostr(n.vars), funstostr(n.fun), sstostr(n.body), decoTostr(n.deco))

	return str
}

func statast(n Statement) string {
	switch e := n.(type) {
	case Print:
		return fmt.Sprintf("Print{expr: %s, newline: %t, deco: %s}", exprast(e.expr), e.newline, decoTostr(e.deco))
	case Return:
		return fmt.Sprintf("Return{expr: %s, deco: %s}", exprast(e.expr), decoTostr(e.deco))
	case Assign:
		return fmt.Sprintf("Assign{name: %s, \nexpr: %s, \ndeco: %s}", e.name, exprast(e.expr), decoTostr(e.deco))
	case FunCall:
		var args string
		if len(e.args) > 0 {
			args = exprast(e.args[0])
			for _, arg := range e.args[1:] {
				args += ",\n" + exprast(arg)
			}
		}
		return fmt.Sprintf("FunCall{name: %s, \nargs: %s, \ndeco: %s}", e.name, args, decoTostr(e.deco))
	case While:
		return fmt.Sprintf("While{expr: %s, \nbody: %s, \ndeco: %s}", exprast(e.expr), sstostr(e.body), decoTostr(e.deco))
	case IfThenElse:
		return fmt.Sprintf("IfThenElse{expr: %s, \nibody: %s, \nebody: %s, \ndeco: %s}", exprast(e.expr), sstostr(e.ibody), sstostr(e.ebody), decoTostr(e.deco))
	default:
		panic(fmt.Sprint("Unknown statastement type", e))
	}
}

func exprast(n Expression) string {
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
		return fmt.Sprintf(`ArithOp{op: %s,left: %s,right: %s,deco: %s}`, pyop, exprast(e.left), exprast(e.right), decoTostr(e.deco))
	case LogicOp:
		pyop := getop(e.op)
		return fmt.Sprintf(`LogicOp{op: %s,left: %s,right: %s,deco: %s}`, pyop, exprast(e.left), exprast(e.right), decoTostr(e.deco))
	case Integer:
		return fmt.Sprintf(`Integer{value: %d deco: %s}`, e.value, decoTostr(e.deco))
	case Boolean:
		return fmt.Sprintf(`Boolean{value: %t deco: %s}`, e.value, decoTostr(e.deco))
	case String:
		return fmt.Sprintf(`String{value: %s deco: %s}`, e.value, decoTostr(e.deco))
	case Var:
		return fmt.Sprintf(`Var{name: %s deco: %s}`, e.name, decoTostr(e.deco))
	case FunCall:
		return fmt.Sprintf(`FunCall{name: %s, args: []Expression{\n%s\n}\ndeco: %s}`, e.name, estostr(e.args), decoTostr(e.deco))
	default:
		panic(fmt.Sprint("Unknown exprastession type", e))
	}
}

func decoTostr(deco map[string]any) string {
	str := "{"
	for k, v := range deco {
		sub := ""
		if e, ok := v.(map[string]any); ok {
			sub += decoTostr(e)
		} else {
			sub += fmt.Sprintf("%v", v)
		}

		str += fmt.Sprintf("{%q: %s}, ", k, sub)
	}
	str += "}"

	return str
}
