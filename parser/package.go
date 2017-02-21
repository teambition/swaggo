package parser

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/teambition/swaggo/swagger"
)

type pkg struct {
	*ast.Package
	localName  string // alias name of package include "."
	importPath string // the import package name
	absPath    string // whereis package in filesystem
	// filename -> import pkgs
	importPkgs map[string][]*pkg
	// model name -> model
	// cache of model struct
	models []*model
}

// newPackage
// filter used for imported packages
func newPackage(localName, importPath string) (p *pkg, err error) {
	goPaths := os.Getenv("GOPATH")
	if goPaths == "" {
		err = fmt.Errorf("GOPATH environment variable is not set or empty")
		return
	}
	// find absolute path
	absPath := ""
	for _, goPath := range filepath.SplitList(goPaths) {
		wg, _ := filepath.EvalSymlinks(filepath.Join(goPath, "src", importPath))
		if fileExists(wg) {
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
		return &pkg{
			Package:    p,
			localName:  localName,
			importPath: importPath,
			absPath:    absPath,
			importPkgs: map[string][]*pkg{},
			models:     []*model{},
		}, nil
	}
	return
}

// parseSchema parse schema in the file
func (p *pkg) parseSchema(s *swagger.Swagger, m *swagger.Schema, filename, schema string) (err error) {
	if strings.HasPrefix(schema, "[]") {
		m.Type = "array"
		m.Items = &swagger.Schema{}
		return p.parseSchema(s, m.Items, filename, schema[2:])
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
		if systemPackageFilter(importPath) {
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
				continue
			case "":
				// import   "lib/math"         math.Sin
			default:
				// import m "lib/math"         m.Sin
			}
			if importPkg, err := newPackage(localName, importPath); err != nil {
				return nil, err
			} else {
				pkgs = append(pkgs, importPkg)
			}
		}
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
				if ts, ok := obj.Decl.(*ast.TypeSpec); ok {
					m := &model{
						TypeSpec: ts,
						name:     name,
						filename: filename,
						p:        p,
					}
					p.models = append(p.models, m)
					return m, nil
				} else {
					return nil, fmt.Errorf("unsupport type(%#v) of model(%s)", obj.Decl, modelName)
				}
			}
		}
	}
	return nil, errModelNotFound
}

// cachedModels
// model name -> import paths
var cachedModels = map[string][]string{}

// parseModel
func (p *pkg) parseModel(s *swagger.Swagger, filename, schema string) (refUrl string, err error) {
	var (
		where     *pkg   // the pakcage that the model in it
		model     *model // model object
		modelName string
	)
	expr := strings.Split(schema, ".")
	switch len(expr) {
	case 1:
		modelName = expr[0]
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
						if err != errModelNotFound {
							return
						}
						continue
					}
					where = v
					break
				}
			}
		} else {
			// in itself
			where = p
		}
	case 2:
		pkgName := expr[0]
		modelName = expr[1]
		var pkgs []*pkg
		if pkgs, err = p.importPackages(filename); err != nil {
			return
		}
		for _, v := range pkgs {
			if v.localName == pkgName || (v.localName == "" && v.Name == pkgName) {
				if model, err = v.findModel(modelName); err != nil {
					return
				}
				where = v
				break
			}
		}
	default:
		err = fmt.Errorf("unsupport schema format(%s) in file(%s)", schema, filename)
		return
	}

	if where == nil || model == nil {
		err = fmt.Errorf("invalid schema(%s) in file(%s)", schema, filename)
		return
	}

	var iterField func(ast.Expr, *swagger.Propertie) (string, string, error)
	iterField = func(e ast.Expr, mp *swagger.Propertie) (innerType, ref string, err error) {
		switch t := e.(type) {
		case *ast.StarExpr:
			return iterField(t.X, mp)
		case *ast.Ident:
			swaggerType, ok := basicTypes[t.Name]
			if ok {
				innerType = t.Name
				typeFormat := strings.Split(swaggerType, ":")
				mp.Type = typeFormat[0]
				mp.Format = typeFormat[1]
				return
			}
			mp.Type = "object"
			ref, err = where.parseModel(s, model.filename, t.Name)
		case *ast.SelectorExpr:
			schema := fmt.Sprint(t)
			schema = strings.Replace(schema, " ", ".", -1)
			schema = strings.Replace(schema, "&", "", -1)
			schema = strings.Replace(schema, "{", "", -1)
			schema = strings.Replace(schema, "}", "", -1)
			mp.Type = "object"
			ref, err = where.parseModel(s, model.filename, schema)
		case *ast.ArrayType:
			mmp := &swagger.Propertie{}
			if innerType, ref, err = iterField(t.Elt, mmp); err != nil {
				return
			}
			mp.Type = "array"
			mp.Items = mmp
		case *ast.MapType:
			mmp := &swagger.Propertie{}
			if innerType, ref, err = iterField(t.Value, mmp); err != nil {
				return
			}
			mp.Type = "object"
			mp.AdditionalProperties = mmp
		}
		return
	}

	st, ok := model.Type.(*ast.StructType)
	if !ok {
		err = fmt.Errorf("unsupport type(%s) only support struct", schema)
		return
	}
	m := swagger.Schema{Title: modelName, Type: "object", Properties: map[string]swagger.Propertie{}}
	for _, f := range st.Fields.List {
		mp := swagger.Propertie{}

		innerType, ref := "", ""
		if innerType, ref, err = iterField(f.Type, &mp); err != nil {
			return
		}
		if innerType == "" && ref == "" {
			err = errors.New("must have a inner type or a decl type")
			return
		}
		if ref != "" {
			if f.Names == nil {
				// anonymous member
				// type A struct {
				//     B
				//     C
				// }
				m.AllOf = append(m.AllOf, &swagger.Schema{Ref: ref})
				continue
			}
			mp.Ref = ref
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
			if mp.Default, err = str2RealType(defautlValue, innerType); err != nil {
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

	// definitions: #/definitions/Model
	key := modelName
	if ips, ok := cachedModels[modelName]; ok {
		exsited := false
		for k, v := range ips {
			if where.importPath == v {
				exsited = true
				if k != 0 {
					key = fmt.Sprintf("%s_%d", modelName, k)
				}
				break
			}
		}
		if !exsited {
			cachedModels[modelName] = append(ips, where.importPath)
		}
	} else {
		cachedModels[modelName] = []string{where.importPath}
	}
	refUrl = "#/definitions/" + key
	if s.Definitions == nil {
		s.Definitions = map[string]swagger.Schema{}
	} else {
		if _, ok := s.Definitions[key]; ok {
			return
		}
	}
	s.Definitions[key] = m
	return
}
