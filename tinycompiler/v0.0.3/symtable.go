package main

import "fmt"

type SymbolTable struct {
	variables []map[string]map[string]any
	functions []map[string]map[string]any
	retStack  []map[string]any
}

type Signature struct {
	name     string
	argtypes []Type
}

func (s Signature) String() string {
	var argtypes string
	if len(s.argtypes) > 0 {
		argtypes = TypeNames[s.argtypes[0]]
		for _, argtype := range s.argtypes[1:] {
			argtypes += "," + TypeNames[argtype]
		}
	}
	return fmt.Sprintf("Signature{name:%s,argtypes:%s}", s.name, argtypes)
}

func (s *SymbolTable) addFun(name string, argtypes []Type, deco map[string]any) {
	signature := Signature{name, argtypes}
	if len(s.functions) > 0 {
		if _, ok := s.functions[len(s.functions)-1][signature.String()]; ok {
			panic(fmt.Sprintf("Double declaration of the variable %s", name))
		}
	} else {
		s.functions = append(s.functions, make(map[string]map[string]any))
	}

	s.functions[len(s.functions)-1][signature.String()] = deco
}

func (s *SymbolTable) addVar(name string, deco map[string]any) {
	if len(s.variables) > 0 {
		if _, ok := s.variables[len(s.variables)-1][name]; ok {
			panic(fmt.Sprintf("Double declaration of the variable %s", name))
		}
	} else {
		s.variables = append(s.variables, make(map[string]map[string]any))
	}

	s.variables[len(s.variables)-1][name] = deco
}

func (s *SymbolTable) pushScope(deco map[string]any) {
	s.variables = append(s.variables, make(map[string]map[string]any))
	s.functions = append(s.functions, make(map[string]map[string]any))
	s.retStack = append(s.retStack, deco)
}

func (s *SymbolTable) popScope() {
	s.variables = s.variables[:len(s.variables)-1]
	s.functions = s.functions[:len(s.functions)-1]
	s.retStack = s.retStack[:len(s.retStack)-1]
}

func (s *SymbolTable) findVar(name string) map[string]any {
	for i := len(s.variables) - 1; i >= 0; i-- {
		if v, ok := s.variables[i][name]; ok {
			return v
		}
	}
	panic(fmt.Sprintf("No declaration for the variable %s", name))
}

func (s *SymbolTable) findFun(name string, argtypes []Type) map[string]any {
	signature := Signature{name, argtypes}
	for i := len(s.functions) - 1; i >= 0; i-- {
		if v, ok := s.functions[i][signature.String()]; ok {
			return v
		}
	}
	panic(fmt.Sprintf("No declaration for the variable %s", signature))
}
