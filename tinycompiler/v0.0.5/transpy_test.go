package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var (
	rootpath string
)

func init() {
	_, exepath, _, _ := runtime.Caller(0)
	rootpath = filepath.Dir(filepath.Dir(exepath))
}

func TestTranspy(t *testing.T) {
	testfiles := []string{"helloworld", "sqrt", "fixed-point", "mutual-recursion", "scope", "overload"}

	for _, base := range testfiles {
		sourceFile := base + ".wend"
		expectedOutput := base + ".expected"

		wendsource, err := os.ReadFile(filepath.Join(rootpath, "test-data", sourceFile))
		if err != nil {
			fmt.Printf("Error read wend file: %s\n", err)
			return
		}

		expected, err := os.ReadFile(filepath.Join(rootpath, "test-data", expectedOutput))
		if err != nil {
			fmt.Printf("Error read expected file: %s\n", err)
			return
		}

		tokens := tokenize(string(wendsource))
		ast := (&WendParser{}).Parse(tokens)
		buildSymtable(ast)
		program := transnovars(ast.(Function))

		cmd := exec.Command("python")
		cmd.Stdin = bytes.NewBuffer([]byte(program))
		var out bytes.Buffer
		cmd.Stdout = &out
		err = cmd.Run()
		if err != nil {
			t.Errorf("fail to exec program %s: %s\n", sourceFile, err)
			return
		}
		result := strings.TrimSpace(out.String())
		expectedResult := strings.TrimSpace(string(expected))
		if result != expectedResult {
			t.Errorf("expected: %s, got: %s\n", expectedResult, result)
		}
	}
}
