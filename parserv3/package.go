package parserv3

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"strings"

	"github.com/teambition/swaggo/swaggerv3"
)

type pkg struct {
	*ast.Package
	localName  string // alias name of package include "."
	importPath string // the import package name
	absPath    string // whereis package in filesystem
	// vendor     string // project vendor for lookup package
	// filename -> import pkgs
	importPkgs map[string][]*pkg
	// model name -> model
	// cache of model struct
	models []*model
}

// newPackage
func newPackage(localName, importPath string, justGoPath bool) (p *pkg, err error) {
	absPath := ""
	ok := false
	if justGoPath {
		absPath, ok = absPathFromGoPath(importPath)
	} else {
		absPath, ok = absPathFromGoPath(importPath)
		if !ok {
			absPath, ok = absPathFromGoRoot(importPath)
		}
	}
	if !ok {
		err = fmt.Errorf("localName(%s), package(%s) does not existed", localName, importPath)
		return
	}
	pkgs, err := parser.ParseDir(token.NewFileSet(), absPath, func(info os.FileInfo) bool {
		name := info.Name()
		return !info.IsDir() &&
			!strings.HasPrefix(name, ".") &&
			// ignore the test file
			!strings.HasSuffix(name, "_test.go")
	}, parser.ParseComments)
	if err != nil {
		log.Printf("[Warning] %v\n", err)
	}

	for k, p := range pkgs {
		// ignore the main package
		if k == "main" {
			continue
		}
		return &pkg{
			Package:    p,
			localName:  localName,
			importPath: importPath,
			//vendor:     vendor,
			absPath:    absPath,
			importPkgs: map[string][]*pkg{},
			models:     []*model{},
		}, nil
	}
	return
}

// parseSchema Parse schema in this code file
func (p *pkg) parseSchema(s *swaggerv3.Swagger, filename, schema string, pathSchema *swaggerv3.PathSchema) (err error) {
	r, err := newModel(filename, ast.NewIdent(schema), p).parse(s)
	if err != nil {
		return err
	}
	r.parsePathSchema(pathSchema)
	return
}

// parseParam Parse param in this code file
func (p *pkg) parseParam(s *swaggerv3.Swagger, sp *swaggerv3.Parameter, filename, schema string) (err error) {
	r, err := newModel(filename, ast.NewIdent(schema), p).parse(s)
	if err != nil {
		return err
	}
	return r.parseParam(sp)
}

// parseParam Parse param in this code file
func (p *pkg) parseBodyParam(s *swaggerv3.Swagger, sp *swaggerv3.RequestBody, filename, schema string) (err error) {
	r, err := newModel(filename, ast.NewIdent(schema), p).parse(s)
	if err != nil {
		return err
	}
	return r.parseBodyParam(sp)
}

// parseImports parse packages from file
// when the qualified identifier has package name
// or cann't be find in self(imported with `.`)
func (p *pkg) parseImports(filename string) ([]*pkg, error) {
	f, ok := p.Files[filename]
	if !ok {
		return nil, fmt.Errorf("file(%s) doesn't existed in package(%s)", filename, p.importPath)
	}

	pkgs := []*pkg{}
	for _, im := range f.Imports {
		importPath := strings.Trim(im.Path.Value, "\"")
		// alias name
		localName := ""
		if im.Name != nil {
			localName = im.Name.Name
		}
		switch localName {
		case ".":
			// import . "lib/math"         Sin
			// all the package's exported identifiers declared in that package's package block
			// will be declared in the importing source file's file block
			// and must be accessed without a qualifier.
		case "_":
			// import _ "path/to/package"  cann't use
			// ignore the imported package
		case "":
			// import   "lib/math"         math.Sin
		default:
			// import m "lib/math"         m.Sin
		}
		importPkg, err := newPackage(localName, importPath, false)
		if err != nil {
			return nil, err
		}
		pkgs = append(pkgs, importPkg)
	}
	p.importPkgs[filename] = pkgs
	return pkgs, nil
}

// importPackages find the cached packages or parse from file
func (p *pkg) importPackages(filename string) ([]*pkg, error) {
	if pkgs, ok := p.importPkgs[filename]; ok {
		return pkgs, nil
	}
	return p.parseImports(filename)
}

var errModelNotFound = errors.New("model not found")

// findModel find typeSpec in self by object name
func (p *pkg) findModel(modelName string) (*model, error) {
	// check in cache
	for _, m := range p.models {
		if m.name == modelName {
			return m, nil
		}
	}

	// check in package
	for filename, f := range p.Files {
		for name, obj := range f.Scope.Objects {
			if name == modelName {
				ts, ok := obj.Decl.(*ast.TypeSpec)
				if ok {
					m := &model{
						name:     name,
						filename: filename,
						p:        p,
						Expr:     ts.Type,
					}
					p.models = append(p.models, m)
					return m, nil
				}
				return nil, fmt.Errorf("unsupport type(%#v) of model(%s)", obj.Decl, modelName)
			}
		}
	}
	return nil, errModelNotFound
}

func (p *pkg) findModelBySchema(filename, schema string) (model *model, err error) {
	expr := strings.Split(schema, ".")
	switch len(expr) {
	case 1:
		modelName := expr[0]
		if model, err = p.findModel(modelName); err != nil {
			if err != errModelNotFound {
				return
			}
			// perhaps in the package imported by `.`
			var pkgs []*pkg
			if pkgs, err = p.importPackages(filename); err != nil {
				return
			}
			for _, v := range pkgs {
				if v.localName == "." {
					if model, err = v.findModel(modelName); err != nil {
						if err == errModelNotFound {
							continue
						}
						return
					}
				}
			}
		}
	case 2:
		pkgName := expr[0]
		modelName := expr[1]
		var pkgs []*pkg
		if pkgs, err = p.importPackages(filename); err != nil {
			return
		}

		for _, v := range pkgs {
			switch v.localName {
			case "", "_":
				if v.Name == pkgName {
					return v.findModel(modelName)
				}
			default:
				if v.localName == pkgName {
					return v.findModel(modelName)
				}
			}
		}
	default:
		err = fmt.Errorf("unsupport schema format(%s) in file(%s)", schema, filename)
		return
	}
	if model == nil {
		err = fmt.Errorf("missing schema(%s) in file(%s)", schema, filename)
		return
	}
	return
}
