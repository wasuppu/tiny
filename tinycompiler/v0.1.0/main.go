package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
	asmProgram := transasm(ast.(Function))

	basename := strings.TrimSuffix(filepath.Base(os.Args[1]), filepath.Ext(os.Args[1]))
	oname := basename + ".o"
	asmname := basename + ".asm"
	err = os.MkdirAll("out", 0755)
	if err != nil {
		fmt.Printf("failed to mkdir %s: %v\n", "out", err)
		return
	}
	os.WriteFile(filepath.Join("out", asmname), []byte(asmProgram), 0644)

	cmd1 := exec.Command("as", "--march=i386", "--32", "-o", filepath.Join("out", oname), filepath.Join("out", asmname))
	if err := cmd1.Run(); err != nil {
		fmt.Printf("error running command: as %s\n", err)
		return
	}
	cmd2 := exec.Command("ld", "-m", "elf_i386", "-o", filepath.Join("out", basename), filepath.Join("out", oname))
	if err := cmd2.Run(); err != nil {
		fmt.Printf("Error running command ld %s\n", err)
		return
	}
}
