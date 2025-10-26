package main

import (
	"fmt"
	"testing"
)

func treeSignature(n any) string {
	switch e := n.(type) {
	case Function:
		var fun, body string
		for _, arg := range e.fun {
			fun += treeSignature(arg)
		}
		for _, arg := range e.body {
			body += treeSignature(arg)
		}
		return fmt.Sprintf("Function{%s,%s}", fun, body)
	case Print:
		return fmt.Sprintf("Print{%s}", treeSignature(e.expr))
	case Return:
		return fmt.Sprintf("Return{%s}", treeSignature(e.expr))
	case Assign:
		return fmt.Sprintf("Assign{%s}", treeSignature(e.expr))
	case FunCall:
		var args string
		for _, v := range e.args {
			args += treeSignature(v)
		}
		return fmt.Sprintf("FunCall{%s}", args)
	case While:
		var body string
		for _, v := range e.body {
			body += treeSignature(v)
		}
		return fmt.Sprintf("While{%s,%s}", treeSignature(e.expr), body)

	case IfThenElse:
		var ibody, ebody string
		for _, v := range e.ibody {
			ibody += treeSignature(v)
		}
		for _, v := range e.ebody {
			ebody += treeSignature(v)
		}
		return fmt.Sprintf("IfThenElse{%s,%s,%s}", treeSignature(e.expr), ibody, ebody)
	case ArithOp:
		return fmt.Sprintf("ArithOp{%s,%s}", treeSignature(e.left), treeSignature(e.right))
	case LogicOp:
		return fmt.Sprintf("LogicOp{%s,%s}", treeSignature(e.left), treeSignature(e.right))
	case Integer:
		return "Integer"
	case Boolean:
		return "Boolean"
	case Var:
		return "Var"
	case String:
		return "String"
	default:
		return ""
	}
}

func TestParser(t *testing.T) {
	tests := []struct {
		program           string
		expectedOutput    string
		expectedSignature string
	}{
		{"main() {print +3 + 5 * -2;}", "-7", "Function{,Print{ArithOp{Integer,ArithOp{Integer,ArithOp{Integer,Integer}}}}}"},
		{"main() {print 3 - 4 * 5;}", "-17", "Function{,Print{ArithOp{Integer,ArithOp{Integer,Integer}}}}"},
		{"main() {print 3 - 10 / 5;}", "1", "Function{,Print{ArithOp{Integer,ArithOp{Integer,Integer}}}}"},
		{"main() {print (-2+3*4)+5/(7-6)%8;}", "15", "Function{,Print{ArithOp{ArithOp{ArithOp{Integer,Integer},ArithOp{Integer,Integer}},ArithOp{ArithOp{Integer,ArithOp{Integer,Integer}},Integer}}}}"},
		{"main() {print 5<3;}", "False", "Function{,Print{LogicOp{Integer,Integer}}}"},
		{"main() {print 3==3;}", "True", "Function{,Print{LogicOp{Integer,Integer}}}"},
		{"main() {print 3 * (4 + 5) / 7 == 3;}", "True", "Function{,Print{LogicOp{ArithOp{ArithOp{Integer,ArithOp{Integer,Integer}},Integer},Integer}}}"},
		{"main() {print true && false;}", "False", "Function{,Print{LogicOp{Boolean,Boolean}}}"},
		{"main() {print true && false || true;}", "True", "Function{,Print{LogicOp{LogicOp{Boolean,Boolean},Boolean}}}"},
		{"main() {print !true;}", "False", "Function{,Print{LogicOp{Boolean,Boolean}}}"},
		{"main() {print 3<=3 && 3>=3;}", "True", "Function{,Print{LogicOp{LogicOp{Integer,Integer},LogicOp{Integer,Integer}}}}"},
	}

	for _, tt := range tests {
		signature := treeSignature((&WendParser{}).Parse(tokenize(tt.program)))
		if signature != tt.expectedSignature {
			t.Fatalf("exprected: %s, got: %s", tt.expectedSignature, signature)
		}
	}
}
