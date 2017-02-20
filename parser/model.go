package parser

import "go/ast"

type model struct {
	*ast.TypeSpec
	name     string // model struct name
	filename string // in which file
	p        *pkg
}
