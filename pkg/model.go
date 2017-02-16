package pkg

import (
	"go/ast"
)

type Model struct {
	*ast.Object
	Name     string // model struct name
	Filename string // in which file
}
