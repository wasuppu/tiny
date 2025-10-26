package main

import (
	"fmt"
	"strings"
	"text/template"
)

var Templates = map[string]string{
	"ascii": `{{.Label}}: .ascii "{{.String}}"
	{{.Label}}_len = . - {{.Label}}
`,
	"var": `	movl display+{{.Scope}}, %eax
	movl -{{.Variable}}(%eax), %eax
`,
	"print_linebreak": `	pushl $10           # "\\n"
	movl $4, %eax       # write system call
	movl $1, %ebx       # stdout
	leal 0(%esp), %ecx  # address of the character
	movl $1, %edx       # one byte
	int  $0x80          # make system call
	addl $4, %esp
`,
	"print_int": `{{.Expr}}
	pushl %eax
	call print_int32
	addl $4, %esp
{{.Newline}}
`,
	"print_string": `	movl $4, %eax
	movl $1, %ebx
	movl ${{.Label}}, %ecx
	movl ${{.Label}}_len, %edx
	int  $0x80
{{.Newline}}
`,
	"print_bool": `{{.Expr}}
	movl $truestr, %ecx
	movl $truestr_len, %edx
	test %eax, %eax
	jnz 0f
	movl $falsestr, %ecx
	movl $falsestr_len, %edx
0:	movl $4, %eax
	movl $1, %ebx
	int  $0x80
{{.Newline}}
`,
	"assign": `{{.Expression}}
	pushl %eax
	movl display+{{.Scope}}, %eax
	popl %ebx
	movl %ebx, -{{.Variable}}(%eax)
`,
	"ifthenelse": `{{.Condition}}
	test %eax, %eax
	jz {{.Label1}}
{{.Ibody}}
	jmp {{.Label2}}
{{.Label1}}:
{{.Ebody}}
{{.Label2}}:
`,
	"while": `{{.Label1}}:
{{.Condition}}
	test %eax, %eax
	jz {{.Label2}}
{{.Body}}
	jmp {{.Label1}}
{{.Label2}}:
`,
	"funcall": `	pushl display+{{.Scope}}
{{.Allocargs}}
	subl ${{.Varsize}}, %esp
	leal {{.Disphead}}(%esp), %eax
	movl %eax, display+{{.Scope}}
	call {{.Funlabel}}
	movl display+{{.Scope}}, %esp
	addl $4, %esp
	popl display+{{.Scope}}
`,
	"program": `.global _start
	.data
{{.Strings}}
truestr: .ascii "true"
	truestr_len = . - truestr
falsestr: .ascii "false"
	falsestr_len = . - falsestr
	.align 2
display: .skip {{.DisplaySize}}
	.text
_start:
	leal -4(%esp), %eax
	movl %eax, display+{{.Offset}}
	subl ${{.Varsize}}, %esp # allocate locals
	call {{.Main}}
	addl ${{.Varsize}}, %esp # deallocate locals
_end:               # do not care about clearing the stack
	movl $1, %eax   # _exit system call (check asm/unistd_32.h for the table)
	movl $0, %ebx   # error code 0
	int $0x80       # make system call
{{.Functions}}
print_int32:
	movl 4(%esp), %eax  # the number to print
	cdq
	xorl %edx, %eax
	subl %edx, %eax     # abs(%eax)
	pushl $10           # base 10
	movl %esp, %ecx     # buffer for the string to print
	subl $16, %esp      # max 10 digits for a 32-bit number (keep %esp dword-aligned)
0:	xorl %edx, %edx     #     %edx = 0
	divl 16(%esp)       #     %eax = %edx:%eax/10 ; %edx = %edx:%eax % 10
	decl %ecx           #     allocate one more digit
	addb $48, %dl       #     %edx += "0"       # 0,0,0,0,0,0,0,0,0,0,"1","2","3","4","5","6"
	movb %dl, (%ecx)    #     store the digit   # ^                   ^                    ^
	test %eax, %eax     #                       # %esp                %ecx (after)         %ecx (before)
	jnz 0b              # until %eax==0         #                     <----- %edx = 6 ----->
	cmp %eax, 24(%esp)  # if the number is negative
	jge 0f
	decl %ecx           # allocate one more character
	movb $45, 0(%ecx)   # "-"
0:	movl $4, %eax       # write system call
	movl $1, %ebx       # stdout
	leal 16(%esp), %edx # the buffer to print
	subl %ecx, %edx     # number of digits
	int $0x80           # make system call
	addl $20, %esp      # deallocate the buffer
	ret
`}

func renderTemplate(templateName string, data any) string {
	t := template.Must(template.New(templateName).Parse(Templates[templateName]))
	var w strings.Builder
	t.Execute(&w, data)
	return w.String()
}

func templateFuncFactory(templateName string) func(map[string]any) string {
	return func(params map[string]any) string {
		return renderTemplate(templateName, params)
	}
}

var TemplateFuns = map[string]func(map[string]any) string{
	"ascii":           templateFuncFactory("ascii"),
	"var":             templateFuncFactory("var"),
	"print_linebreak": templateFuncFactory("print_linebreak"),
	"print_int":       templateFuncFactory("print_int"),
	"print_string":    templateFuncFactory("print_string"),
	"print_bool":      templateFuncFactory("print_bool"),
	"assign":          templateFuncFactory("assign"),
	"ifthenelse":      templateFuncFactory("ifthenelse"),
	"while":           templateFuncFactory("while"),
	"funcall":         templateFuncFactory("funcall"),
	"program":         templateFuncFactory("program"),
}

func transasm(n Function) string {
	var strings string
	for label, strs := range n.deco["strings"].(map[string]string) {
		strings += TemplateFuns["ascii"](map[string]any{"Label": label, "String": strs})
	}
	displaySize := n.deco["scopeCnt"].(int) * 4
	offset := n.deco["scope"].(int) * 4
	main := n.deco["label"]
	varsize := len(n.vars)
	functions := funasm(n)
	return TemplateFuns["program"](
		map[string]any{
			"Strings":     strings,
			"DisplaySize": displaySize,
			"Offset":      offset,
			"Varsize":     varsize,
			"Main":        main,
			"Functions":   functions,
		})
}

func funasm(n Function) string {
	label := n.deco["label"]
	nested := ""
	for _, f := range n.fun {
		nested += funasm(f)
	}
	body := ""
	for _, s := range n.body {
		body += statasm(s)
	}
	return fmt.Sprintf("%s:\n%s\n\tret\n%s\n", label, body, nested)
}

func statasm(n Statement) string {
	switch e := n.(type) {
	case Print:
		var newline string
		if e.newline {
			newline = Templates["print_linebreak"]
		}
		switch e.expr.getDeco()["type"].(Type) {
		case INT:
			return TemplateFuns["print_int"](map[string]any{"Expr": exprasm(e.expr), "Newline": newline})
		case BOOL:
			return TemplateFuns["print_bool"](map[string]any{"Expr": exprasm(e.expr), "Newline": newline})
		case STRING:
			return TemplateFuns["print_string"](map[string]any{"Label": e.expr.getDeco()["label"].(string), "Newline": newline})
		default:
			panic(fmt.Sprintln("Unknown expression type", e.expr))
		}
	case Return:
		if e.expr != nil && e.expr.getDeco()["type"].(Type) != VOID {
			return exprasm(e.expr) + "\tret\n"
		} else {
			return "\tret\n"
		}
	case Assign:
		return TemplateFuns["assign"](map[string]any{"Expression": exprasm(e.expr), "Scope": e.deco["scope"].(int) * 4, "Variable": e.deco["offset"].(int) * 4})
	case FunCall:
		return exprasm(e)
	case While:
		var body string
		for _, s := range e.body {
			body += statasm(s)
		}
		return TemplateFuns["while"](map[string]any{"Condition": exprasm(e.expr), "Label1": newLabel(), "Label2": newLabel(), "Body": body})
	case IfThenElse:
		var ibody string
		for _, s := range e.ibody {
			ibody += statasm(s)
		}
		var ebody string
		for _, s := range e.ebody {
			ebody += statasm(s)
		}
		return TemplateFuns["ifthenelse"](map[string]any{"Condition": exprasm(e.expr), "Label1": newLabel() + "if1", "Label2": newLabel() + "if2", "Ibody": ibody, "Ebody": ebody})
	default:
		panic(fmt.Sprint("Unknown statement type", e))
	}
}

func exprasm(n Expression) string {
	pyeq1 := map[string]string{"+": "addl", "-": "subl", "*": "imull", "||": "orl", "&&": "andl"}
	pyeq2 := map[string]string{"<=": "jle", "<": "jl", ">=": "jge", ">": "jg", "==": "je", "!=": "jne"}
	isContain := func(eop string, pyeq map[string]string) bool {
		_, ok := pyeq[eop]
		return ok
	}
	switch e := n.(type) {
	case ArithOp:
		args := exprasm(e.left) + "\tpushl %eax\n" + exprasm(e.right) + "\tmovl %eax, %ebx\n\tpopl %eax\n"
		if isContain(e.op, pyeq1) {
			return args + fmt.Sprintf("\t%s %%ebx, %%eax\n", pyeq1[e.op])
		} else if isContain(e.op, pyeq2) {
			return args + fmt.Sprintf("\tcmp %%ebx, %%eax\n\tmovl $1, %%eax\n\t%s 1f\n\txorl %%eax, %%eax\n1:\n", pyeq2[e.op])
		} else if e.op == "/" {
			return args + "\tcdq\n\tidivl %ebx, %eax\n"
		} else if e.op == "%" {
			return args + "\tcdq\n\tidivl %ebx, %eax\n\tmovl %edx, %eax\n"
		}
		panic("Unknown binary operation")
	case LogicOp:
		args := exprasm(e.left) + "\tpushl %eax\n" + exprasm(e.right) + "\tmovl %eax, %ebx\n\tpopl %eax\n"
		if isContain(e.op, pyeq1) {
			return args + fmt.Sprintf("\t%s %%ebx, %%eax\n", pyeq1[e.op])
		} else if isContain(e.op, pyeq2) {
			return args + fmt.Sprintf("\tcmp %%ebx, %%eax\n\tmovl $1, %%eax\n\t%s 1f\n\txorl %%eax, %%eax\n1:\n", pyeq2[e.op])
		} else if e.op == "/" {
			return args + "\tcdq\n\tidivl %ebx, %eax\n"
		} else if e.op == "%" {
			return args + "\tcdq\n\tidivl %ebx, %eax\n\tmovl %edx, %eax\n"
		}
		panic("Unknown binary operation")
	case Integer:
		return fmt.Sprintf("\tmovl $%d, %%eax\n", e.value)
	case Boolean:
		var value int
		if e.value {
			value = 1
		} else {
			value = 0
		}
		return fmt.Sprintf("\tmovl $%d, %%eax\n", value)
	case Var:
		return TemplateFuns["var"](map[string]any{"Scope": e.deco["scope"].(int) * 4, "Variable": e.deco["offset"].(int) * 4})
	case FunCall:
		var allocargs string
		for _, arg := range e.args {
			allocargs += fmt.Sprintf("%s\tpushl %%eax\n", exprasm(arg))
		}
		varsize := len(e.deco["fundeco"].(map[string]any)["local"].([]string)) * 4
		disphead := varsize + len(e.args)*4 - 4
		scope := e.deco["fundeco"].(map[string]any)["scope"].(int) * 4
		funlabel := e.deco["fundeco"].(map[string]any)["label"].(string)
		return TemplateFuns["funcall"](map[string]any{"Scope": scope, "Allocargs": allocargs, "Varsize": varsize, "Disphead": disphead, "Funlabel": funlabel})
	default:
		panic(fmt.Sprint("Unknown expression type", e))
	}
}
