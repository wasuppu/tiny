package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: ./compiler path/source.wend")
		return
	}

	source, err := os.ReadFile(os.Args[1])
	if err != nil {
		fmt.Printf("failed to read %s: %v\n", os.Args[1], err)
		return
	}

	tokens := tokenize(string(source))
	ast := (&WendParser{}).Parse(tokens)
	buildSymtable(ast)
	fmt.Println(transasm(ast.(Function)))
}
