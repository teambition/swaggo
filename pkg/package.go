package pkg

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/teambition/swaggo/swagger"
	"github.com/teambition/swaggo/utils"
)

// refer to builtin.go
var basicTypes = map[string]string{
	"bool":       "boolean:",
	"uint":       "integer:int32",
	"uint8":      "integer:int32",
	"uint16":     "integer:int32",
	"uint32":     "integer:int32",
	"uint64":     "integer:int64",
	"int":        "integer:int32",
	"int8":       "integer:int32",
	"int16":      "integer:int32",
	"int32":      "integer:int32",
	"int64":      "integer:int64",
	"uintptr":    "integer:int64",
	"float32":    "number:float",
	"float64":    "number:double",
	"string":     "string:",
	"complex64":  "number:float",
	"complex128": "number:double",
	"byte":       "string:byte",
	"rune":       "string:byte",
}

type Package struct {
	*ast.Package
	LocalName  string // alias name of package include "."
	ImportPath string // the import package name
	AbsPath    string // whereis package in filesystem
	// model name -> model
	// filename -> import pkgs
	importPkgs map[string][]*Package
	// cache of model struct
	models []*Model
	// filter the import path
	filter func(string) bool
}

func NewPackage(localName, importPath string, filter func(string) bool) (p *Package, err error) {
	goPaths := os.Getenv("GOPATH")
	if goPaths == "" {
		err = fmt.Errorf("GOPATH environment variable is not set or empty")
		return
	}
	// find absolute path
	absPath := ""
	for _, goPath := range filepath.SplitList(goPaths) {
		wg, _ := filepath.EvalSymlinks(filepath.Join(goPath, "src", importPath))
		if utils.FileExists(wg) {
			absPath = wg
		}
	}
	if absPath == "" {
		err = fmt.Errorf("package(%s) does not exist in the GOPATH", importPath)
		return
	}

	pkgs, err := parser.ParseDir(token.NewFileSet(), absPath, func(info os.FileInfo) bool {
		name := info.Name()
		return !info.IsDir() && !strings.HasPrefix(name, ".")
	}, parser.ParseComments)
	if err != nil {
		return
	}

	for _, p := range pkgs {
		return &Package{
			Package:    p,
			LocalName:  localName,
			ImportPath: importPath,
			AbsPath:    absPath,
			importPkgs: map[string][]*Package{},
			models:     []*Model{},
			filter:     filter,
		}, nil
	}
	return
}

func (p *Package) parseImports(filename string) ([]*Package, error) {
	f, ok := p.Files[filename]
	if !ok {
		return nil, fmt.Errorf("file(%s) doesn't existed in package(%s)", filename, p.AbsPath)
	}

	pkgs := []*Package{}
	for _, im := range f.Imports {
		importPath := strings.Trim(im.Path.Value, "\"")
		if p.filter == nil || p.filter(importPath) {
			// alias name
			localName := ""
			if im.Name != nil {
				localName = im.Name.Name
			}
			if importPkg, err := NewPackage(localName, importPath, p.filter); err != nil {
				return nil, err
			} else {
				pkgs = append(pkgs, importPkg)
			}
		}
	}
	p.importPkgs[filename] = pkgs
	return pkgs, nil
}

func (p *Package) importPackages(filename string) ([]*Package, error) {
	if pkgs, ok := p.importPkgs[filename]; ok {
		return pkgs, nil
	}
	return p.parseImports(filename)
}

func (p *Package) findModel(modelName string) (*Model, error) {
	// check in cache
	for _, m := range p.models {
		if m.Name == modelName {
			return m, nil
		}
	}
	// check in package
	for filename, f := range p.Files {
		for name, obj := range f.Scope.Objects {
			if name == modelName {
				m := &Model{
					Object:   obj,
					Name:     name,
					Filename: filename,
				}
				p.models = append(p.models, m)
				return m, nil
			}
		}
	}
	return nil, fmt.Errorf("model(%s) cann't found in package(%s)", modelName, p.AbsPath)
}

func (p *Package) Parse(s *swagger.Swagger, m *swagger.Schema, filename, schema string) (err error) {
	if strings.HasPrefix(schema, "[]") {
		m.Type = "array"
		m.Items = &swagger.Schema{}
		return p.Parse(s, m.Items, filename, schema[2:])
	}
	// file body
	if schema == "file" {
		m.Type = "file"
		return
	}
	if swaggerType, ok := basicTypes[schema]; ok {
		typeFormat := strings.Split(swaggerType, ":")
		m.Type = typeFormat[0]
		m.Format = typeFormat[1]
		return
	}
	m.Type = "object"
	ref, err := p.parseModel(s, filename, schema)
	if err != nil {
		return err
	}
	m.Ref = ref
	return
}

// parseModel
func (p *Package) parseModel(s *swagger.Swagger, filename, schema string) (key string, err error) {
	var pkg *Package
	expr := strings.Split(schema, ".")
	modelName := ""
	switch len(expr) {
	case 1:
		pkg = p
		modelName = expr[0]
	case 2:
		pkgName := expr[0]
		modelName = expr[1]
		var pkgs []*Package
		if pkgs, err = p.importPackages(filename); err != nil {
			return
		}
		for _, p := range pkgs {
			if p.LocalName == pkgName || p.Name == pkgName {
				pkg = p
				break
			}
		}
	default:
		err = fmt.Errorf("unsupport schema(%s) in file(%s)", schema, filename)
		return
	}

	if pkg == nil {
		err = fmt.Errorf("invalid schema(%s) in file(%s)", schema, filename)
		return
	}

	key = p.ImportPath + "#" + modelName
	if s.Definitions == nil {
		s.Definitions = map[string]swagger.Schema{}
	} else {
		if _, ok := s.Definitions[key]; ok {
			return
		}
	}

	model, err := pkg.findModel(modelName)
	if err != nil {
		return
	}

	var iterFiled func(ast.Expr, *swagger.Propertie) (string, error)
	iterFiled = func(f ast.Expr, mp *swagger.Propertie) (Type string, err error) {
		switch t := f.(type) {
		case *ast.StarExpr:
			Type, err = iterFiled(t.X, mp)
		case *ast.Ident:
			swaggerType, ok := basicTypes[t.Name]
			if ok {
				Type = t.Name
				typeFormat := strings.Split(swaggerType, ":")
				mp.Type = typeFormat[0]
				mp.Format = typeFormat[1]
				return
			}
			mp.Type = "object"
			ref := ""
			if ref, err = pkg.parseModel(s, model.Filename, t.Name); err != nil {
				return
			}
			mp.Ref = "#/definitions/" + ref
		case *ast.SelectorExpr:
			schema := fmt.Sprint(t)
			schema = strings.Replace(schema, " ", ".", -1)
			schema = strings.Replace(schema, "&", "", -1)
			schema = strings.Replace(schema, "{", "", -1)
			schema = strings.Replace(schema, "}", "", -1)
			mp.Type = "object"
			ref := ""
			if ref, err = pkg.parseModel(s, model.Filename, schema); err != nil {
				return
			}
			mp.Ref = "#/definitions/" + ref
		case *ast.ArrayType:
			mp.Type = "array"
			mmp := &swagger.Propertie{}
			if _, err = iterFiled(t.Elt, mmp); err != nil {
				return
			}
			mp.Items = mmp
		case *ast.MapType:
			mp.Type = "object"
			mmp := &swagger.Propertie{}
			if _, err = iterFiled(t.Value, mmp); err != nil {
				return
			}
			mp.AdditionalProperties = mmp
		}
		return
	}

	var iterStruct func(*ast.StructType, *swagger.Schema) error
	iterStruct = func(st *ast.StructType, m *swagger.Schema) (err error) {
		for _, f := range st.Fields.List {
			mp := swagger.Propertie{}

			realType := ""
			if realType, err = iterFiled(f.Type, &mp); err != nil {
				return
			}
			if f.Names == nil {
				// anonymous struct
				// type A struct {
				//     B
				// }
				continue
			}
			name := f.Names[0].Name
			// check if it has tags
			if f.Tag == nil {
				m.Properties[name] = mp
				continue
			}
			// parse tag for name
			stag := reflect.StructTag(strings.Trim(f.Tag.Value, "`"))
			// doc tag include
			// if the filed is a integer we can set: default:`123`
			defautlValue := stag.Get("default")
			if defautlValue != "" {
				if mp.Default, err = utils.Str2RealType(defautlValue, realType); err != nil {
					return
				}
			}

			tagValues := strings.Split(stag.Get("json"), ",")
			// dont add property if json tag first value is "-"
			if len(tagValues) == 0 || tagValues[0] != "-" {
				// set property name to the left most json tag value only if is not empty
				if len(tagValues) > 0 && tagValues[0] != "" {
					name = tagValues[0]
				}

				if required := stag.Get("required"); required != "" {
					m.Required = append(m.Required, name)
				}
				if desc := stag.Get("desc"); desc != "" {
					mp.Description = desc
				}
				m.Properties[name] = mp
			}
		}
		return
	}

	ts, ok := model.Decl.(*ast.TypeSpec)
	if !ok {
		err = fmt.Errorf("unknown type without TypeSpec(%v)", model)
		return
	}
	st, ok := ts.Type.(*ast.StructType)
	if !ok {
		err = fmt.Errorf("unsupport type(%s) only support struct", schema)
		return
	}

	m := swagger.Schema{Title: modelName, Type: "object", Properties: map[string]swagger.Propertie{}}
	if err = iterStruct(st, &m); err != nil {
		return
	}

	s.Definitions[key] = m
	return
}
