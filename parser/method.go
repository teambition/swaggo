package parser

import "go/ast"

// method the method of controllor
type method struct {
	*ast.FuncDecl
	name     string // function name
	filename string // whereis
}
