package parser

import (
	"fmt"
	"go/ast"
	"go/token"

	"github.com/teambition/swaggo/swagger"
)

// Resource api resource
type Resource struct {
	*pkg
	// maybe one resource has several controller
	controllers map[string]*controller // ctrl name -> ctrl
}

func NewResoucre(importPath string, filter func(string) bool) (*Resource, error) {
	p, err := newPackage("_", importPath, filter)
	if err != nil {
		return nil, err
	}

	r := &Resource{
		pkg:         p,
		controllers: map[string]*controller{},
	}
	for filename, f := range p.Files {
		for _, d := range f.Decls {
			switch specDecl := d.(type) {
			case *ast.FuncDecl:
				if specDecl.Recv != nil && len(specDecl.Recv.List) != 0 {
					if t, ok := specDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
						if isDocComments(specDecl.Doc) {
							ctrlName := fmt.Sprint(t.X)
							if ctrl, ok := r.controllers[ctrlName]; !ok {
								r.controllers[ctrlName] = &controller{
									r: r,
									methods: []*method{
										&method{
											FuncDecl: specDecl,
											filename: filename,
											name:     specDecl.Name.Name,
										},
									},
								}
							} else {
								ctrl.methods = append(ctrl.methods, &method{
									FuncDecl: specDecl,
									filename: filename,
									name:     specDecl.Name.Name,
								})
							}
						}
					}
				}
			case *ast.GenDecl:
				if specDecl.Tok == token.TYPE {
					for _, s := range specDecl.Specs {
						t := s.(*ast.TypeSpec)
						switch t.Type.(type) {
						case *ast.StructType:
							if isDocComments(specDecl.Doc) {
								ctrlName := t.Name.String()
								if ctrl, ok := r.controllers[ctrlName]; !ok {
									r.controllers[ctrlName] = &controller{
										TypeSpec: t,
										doc:      specDecl.Doc,
										r:        r,
										name:     t.Name.Name,
										methods:  []*method{},
									}
								} else {
									ctrl.TypeSpec = t
									ctrl.doc = specDecl.Doc
									ctrl.name = t.Name.Name
								}
							}
						}
					}
				}
			}
		}
	}
	return r, nil
}

func (r *Resource) Run(s *swagger.Swagger) error {
	// parse controllers
	for _, ctrl := range r.controllers {
		if err := ctrl.parse(s); err != nil {
			return err
		}
	}
	return nil

}
