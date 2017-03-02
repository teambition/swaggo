package parser

import (
	"fmt"
	"go/ast"
	"reflect"
	"regexp"
	"strings"

	"github.com/teambition/swaggo/swagger"
)

type model struct {
	*ast.TypeSpec
	name     string // model struct name
	filename string // in which file
	p        *pkg
}

// parse parse the model in go code
func (m *model) parse(s *swagger.Swagger, e ast.Expr) (r *result, err error) {
	switch t := e.(type) {
	case *ast.StarExpr:
		return m.parse(s, t.X)
	case *ast.Ident, *ast.SelectorExpr:
		schema := fmt.Sprint(t)
		r = &result{}
		// []SomeStruct
		if strings.HasPrefix(schema, "[]") {
			schema = schema[2:]
			r.kind = arrayType
			r.item, err = m.parse(s, ast.NewIdent(schema))
			return
		}
		// &{foo Bar} to foo.Bar
		reInternalRepresentation := regexp.MustCompile("&\\{(\\w*) (\\w*)\\}")
		schema = string(reInternalRepresentation.ReplaceAll([]byte(schema), []byte("$1.$2")))
		// check if is basic type
		if swaggerType, ok := basicTypes[schema]; ok {
			return &result{
				kind:    innerType,
				buildin: schema,
				swagger: swaggerType,
			}, nil
		}
		if nm, err := m.p.findModelBySchema(m.filename, schema); err != nil {
			return nil, err
		} else {
			return nm.parse(s, nm.Type)
		}
	case *ast.ArrayType:
		r = &result{kind: arrayType}
		r.item, err = m.parse(s, t.Elt)
	case *ast.MapType:
		r = &result{kind: mapType}
		r.item, err = m.parse(s, t.Value)
	case *ast.InterfaceType:
		return &result{kind: interfaceType}, nil
	case *ast.StructType:
		r = &result{kind: objectType}
		// definitions: #/definitions/Model
		key := m.name
		// check if existed
		if ips, ok := cachedModels[m.name]; ok {
			exsited := false
			for k, v := range ips {
				if m.p.importPath == v {
					exsited = true
					if k != 0 {
						key = fmt.Sprintf("%s_%d", m.name, k)
					}
					break
				}
			}
			if !exsited {
				cachedModels[m.name] = append(ips, m.p.importPath)
			}
		} else {
			cachedModels[m.name] = []string{m.p.importPath}
		}

		if s.Definitions == nil {
			s.Definitions = map[string]swagger.Schema{}
		}
		// schema missing
		if _, ok := s.Definitions[key]; !ok {
			ss := swagger.Schema{Title: m.name, Type: "object", Properties: map[string]swagger.Propertie{}}
			for _, f := range t.Fields.List {
				var (
					sp   swagger.Propertie
					tmpR *result
				)
				if tmpR, err = m.parse(s, f.Type); err != nil {
					return
				}
				if f.Names == nil {
					// result must be a struct
					if tmpR.ref == "" {
						err = fmt.Errorf("anonymous member must has a struct type")
						return
					}
					// anonymous member
					// type A struct {
					//     B
					//     C
					// }
					ss.AllOf = append(ss.AllOf, &swagger.Schema{Ref: tmpR.ref})
					continue
				}
				tmpR.parsePropertie(&sp)
				name := f.Names[0].Name
				// check if it has tags
				if f.Tag == nil {
					ss.Properties[name] = sp
					continue
				}
				// parse tag for name
				stag := reflect.StructTag(strings.Trim(f.Tag.Value, "`"))
				// check jsonTag == "-"
				jsonTag := strings.Split(stag.Get("json"), ",")
				if len(jsonTag) != 0 && jsonTag[0] == "-" {
					continue
				}
				name = jsonTag[0]
				// swaggo:"(desc),(required),(default)"
				swaggoTag := stag.Get("swaggo")
				tmp := strings.Split(swaggoTag, ",")
				for k, v := range tmp {
					switch k {
					case 0:
						if v != "" && v != "-" {
							ss.Required = append(ss.Required, name)
						}
					case 1:
						sp.Description = v
					case 2:
						if v != "" {
							if sp.Default, err = str2RealType(v, r.buildin); err != nil {
								return
							}
						}
					}
				}

				ss.Properties[name] = sp
			}
			s.Definitions[key] = ss
		}
		r.ref = "#/definitions/" + key
	}
	return
}

// cachedModels
// model name -> import paths
var cachedModels = map[string][]string{}

const (
	innerType     = "inner"
	arrayType     = "array"
	mapType       = "map"
	objectType    = "object"
	interfaceType = "interface"
)

type result struct {
	kind    string
	buildin string  // buildin type
	swagger string  // kind == inner
	ref     string  // kind == object
	item    *result // kind in (array, map)
}

func (r *result) parseParam(sp *swagger.Parameter) error {
	switch sp.In {
	case body:
		if sp.Schema == nil {
			sp.Schema = &swagger.Schema{}
		}
		r.parseSchema(sp.Schema)
	default:
		switch r.kind {
		case innerType:
			tmp := strings.Split(r.swagger, ":")
			sp.Type = tmp[0]
			sp.Format = tmp[1]
		case arrayType:
			sp.Type = arrayType
			sp.Items = &swagger.ParameterItems{}
			if err := r.item.parseParamItem(sp.Items); err != nil {
				return err
			}
		default:
			// TODO
			// not support object and array in any value other than "body"
			return fmt.Errorf("not support(%s) in(%s) any value other than `body`", r.kind, sp.In)
		}
	}
	return nil
}

func (r *result) parseParamItem(sp *swagger.ParameterItems) error {
	switch r.kind {
	case innerType:
		tmp := strings.Split(r.swagger, ":")
		sp.Type = tmp[0]
		sp.Format = tmp[1]
	case arrayType:
		sp.Type = arrayType
		sp.Items = &swagger.ParameterItems{}
		if err := r.item.parseParamItem(sp.Items); err != nil {
			return err
		}
	default:
		// TODO
		// param not support object, map, interface
		return fmt.Errorf("not support(%s) in any value other than `body`", r.kind)
	}
	return nil
}

func (r *result) parseSchema(ss *swagger.Schema) {
	switch r.kind {
	case innerType:
		tmp := strings.Split(r.swagger, ":")
		ss.Type = tmp[0]
		ss.Format = tmp[1]
	case objectType:
		ss.Type = objectType
		ss.Ref = r.ref
	case arrayType:
		ss.Type = arrayType
		ss.Items = &swagger.Schema{}
		r.item.parseSchema(ss.Items)
	case interfaceType:
		ss.Type = objectType
	}
}

func (r *result) parsePropertie(sp *swagger.Propertie) {
	switch r.kind {
	case innerType:
		tmp := strings.Split(r.swagger, ":")
		sp.Type = tmp[0]
		sp.Format = tmp[1]
	case arrayType:
		sp.Type = arrayType
		sp.Items = &swagger.Propertie{}
		r.item.parsePropertie(sp.Items)
	case mapType:
		sp.Type = objectType
		sp.AdditionalProperties = &swagger.Propertie{}
		r.item.parsePropertie(sp.AdditionalProperties)
	case objectType:
		sp.Type = objectType
		sp.Ref = r.ref
	case interfaceType:
		sp.Type = objectType
		// TODO
	}
}

// inner type and swagger type
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
	"time.Time":  "string:date-time",
	"file":       "file:",
}
