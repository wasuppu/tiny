package main

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"testing"
)

func TestSyntree(t *testing.T) {
	absfun := Function{
		"abs", // function name
		[]Var{{"x", map[string]any{"type": INT}}}, // one integer argument
		[]Var{},      // no local variables
		[]Function{}, // no nested functions
		[]Statement{
			IfThenElse{
				expr:  LogicOp{op: "<", left: Var{name: "x"}, right: Integer{value: 0}},
				ibody: []Statement{Return{expr: ArithOp{op: "-", left: Integer{value: 0}, right: Var{name: "x"}}}},
				ebody: []Statement{Return{expr: Var{name: "x"}}}}},
		map[string]any{"type": INT},
	}

	sqrtfun := Function{
		"sqrt", // function name
		[]Var{{"n", map[string]any{"type": INT}}, {"shift", map[string]any{"type": INT}}}, // input fixed-point number (two integer variables)
		[]Var{{"x", map[string]any{"type": INT}}, // three local integer variables
			{"x_old", map[string]any{"type": INT}},
			{"n_one", map[string]any{"type": INT}}},
		[]Function{}, // no nested functions
		[]Statement{ // function body
			IfThenElse{
				expr: LogicOp{op: ">", left: Var{name: "n"}, right: Integer{value: 65535}}, // if n > 65535
				ibody: []Statement{
					Return{
						expr: ArithOp{
							op: "*", left: Integer{value: 2}, //    return 2*sqrt{n/4}
							right: FunCall{name: "sqrt", args: []Expression{ArithOp{op: "/", left: Var{name: "n"}, right: Integer{value: 4}}, Var{name: "shift"}}}}}},
				ebody: []Statement{}}, // no else statements
			Assign{name: "x", expr: Var{name: "shift"}},                                                    // x = shift
			Assign{name: "n_one", expr: ArithOp{op: "*", left: Var{name: "n"}, right: Var{name: "shift"}}}, // n_one = n*shift
			While{
				expr: Boolean{value: true},
				body: []Statement{ // while true
					Assign{name: "x_old", expr: Var{name: "x"}}, //     x_old = x
					Assign{name: "x", //     x = {x + n_one / x} / 2
						expr: ArithOp{op: "/",
							left: ArithOp{op: "+",
								left:  Var{name: "x"},
								right: ArithOp{op: "/", left: Var{name: "n_one"}, right: Var{name: "x"}}},
							right: Integer{value: 2}}},
					IfThenElse{ //     if abs{x-x_old} <= 1
						expr: LogicOp{op: "<=",
							left:  FunCall{name: "abs", args: []Expression{ArithOp{op: "-", left: Var{name: "x"}, right: Var{name: "x_old"}}}},
							right: Integer{value: 1}},
						ibody: []Statement{Return{expr: Var{name: "x"}}}, //        return x
						ebody: []Statement{}},                            //     no else statements
				}}},
		map[string]any{"type": INT},
	} // return type

	mainfun := Function{"main", // function name
		[]Var{},                     // no arguments
		[]Var{},                     // no local variables
		[]Function{sqrtfun, absfun}, // two nested functions
		[]Statement{Print{ // println sqrt(25735, 8192);
			expr:    FunCall{name: "sqrt", args: []Expression{Integer{value: 25735}, Integer{value: 8192}}},
			newline: true}},
		map[string]any{"type": VOID},
	}

	program := transpy(mainfun)
	cmd := exec.Command("python")
	cmd.Stdin = bytes.NewBuffer([]byte(program))
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()
	if err != nil {
		fmt.Println(err)
		return
	}
	result := strings.TrimSpace(out.String())
	if result != "14519" {
		t.Errorf("expected: %s, got: %s", "14519", result)
	}
}
