package parserv3

import (
	"fmt"
	"go/ast"
	"reflect"
	"regexp"
	"strings"

	"github.com/teambition/swaggo/swaggerv3"
)

var typeMap = map[string]string{
	"util.JSONString": "string",
	"util.Bool":       "bool",
	"util.Int":        "int",
	"util.Interface":  "interface {}",
	"util.Number":     "float64",
	"util.ObjectID":   "bson.ObjectId",
	"util.String":     "string",
	"util.Strings":    "[]string",
	"util.Time":       "time.Time",
}

// feature if the expression is an anonymous member or an anonymous struct
// it's useful for displays the model with swagger
type feature int

const (
	noneFeature       feature = iota
	anonMemberFeature         // as an anonymous member type
	anonStructFeature         // as an anonymous struct type
)

// model the type of golang
type model struct {
	ast.Expr        // golang ast
	name     string // the real name of model
	filename string // appear in which file
	p        *pkg   // appear in which package
	f        feature
	extend   schemaExtend
}

func getSchema(schema string) string {
	if v, ok := typeMap[schema]; ok {
		return v
	}

	return schema
}

// newModel create a model with file path, schema expression and package object
func newModel(filename string, e ast.Expr, p *pkg, extend schemaExtend) *model {
	return &model{
		Expr:     e,
		filename: filename,
		p:        p,
		extend:   extend,
	}
}

// member the member of struct type with same environment
func (m *model) member(e ast.Expr) *model {
	return newModel(m.filename, e, m.p, schemaExtend{})
}

// clone clone the model expect model's name
func (m *model) clone(e ast.Expr) *model {
	nm := *m
	nm.name = ""
	nm.Expr = e
	return &nm
}

// inhertFeature inhert the feature from other model
func (m *model) inhertFeature(other *model) *model {
	m.f = other.f
	m.extend = other.extend
	return m
}

func (m *model) anonymousMember() *model {
	m.f = anonMemberFeature
	return m
}

func (m *model) anonymousStruct() *model {
	m.f = anonStructFeature
	return m
}

// parse parse the model in go code
func (m *model) parse(s *swaggerv3.Swagger) (r *result, err error) {
	switch t := m.Expr.(type) {
	case *ast.StarExpr:
		return m.clone(t.X).parse(s)
	case *ast.Ident, *ast.SelectorExpr:
		schema := fmt.Sprint(t)
		schema = getSchema(schema)
		r = &result{}
		// []SomeStruct
		if strings.HasPrefix(schema, "[]") {
			schema = schema[2:]
			schema = getSchema(schema)
			r.kind = arrayKind
			r.item, err = m.clone(ast.NewIdent(schema)).parse(s)
			return
		}
		// map[string]SomeStruct
		if strings.HasPrefix(schema, "map[string]") {
			schema = schema[11:]
			schema = getSchema(schema)
			r.kind = mapKind
			r.item, err = m.clone(ast.NewIdent(schema)).parse(s)
			return
		}
		// &{foo Bar} to foo.Bar
		reInternalRepresentation := regexp.MustCompile("&\\{(\\w*) (\\w*)\\}")
		schema = string(reInternalRepresentation.ReplaceAll([]byte(schema), []byte("$1.$2")))
		schema = getSchema(schema)
		// check if is basic type
		if swaggerType, ok := basicTypes[schema]; ok {
			tmp := strings.Split(swaggerType, ":")
			typ := tmp[0]
			format := tmp[1]

			if strings.HasPrefix(typ, "[]") {
				schema = typ[2:]
				schema = getSchema(schema)
				r.kind = arrayKind
				r.item, err = m.clone(ast.NewIdent(schema)).parse(s)
			} else {
				r.kind = innerKind
				r.buildin = schema
				r.sType = typ
				r.sFormat = format
			}
			return
		}

		nm, err := m.p.findModelBySchema(m.filename, schema)
		if err != nil {
			fmt.Println(err)
			return nil, fmt.Errorf("findModelBySchema filename(%s) schema(%s) error(%v)", m.filename, schema, err)
		}
		return nm.inhertFeature(m).parse(s)
	case *ast.ArrayType:
		r = &result{kind: arrayKind}
		r.item, err = m.clone(t.Elt).parse(s)
	case *ast.MapType:
		r = &result{kind: mapKind}
		r.item, err = m.clone(t.Value).parse(s)
	case *ast.InterfaceType:
		return &result{kind: interfaceKind}, nil
	case *ast.StructType:
		r = &result{kind: objectKind, items: map[string]*result{}}
		if m.name == "" {
			m.anonymousStruct()
		} else {
			if m.extend.template != "" {
				r.ref = "#/components/schemas/" + m.name + ":" + m.extend.template
			} else {
				r.ref = "#/components/schemas/" + m.name
			}

			if s.Components.Schemas == nil {
				s.Components.Schemas = map[string]*swaggerv3.Schema{}
			}
		}

		for _, f := range t.Fields.List {
			var (
				childR *result
				nm     = m.member(f.Type)
				//name   string
				names []string
			)
			if len(f.Names) == 0 {
				nm.anonymousMember()
			} else {
				//name = f.Names[0].Name
				names = append(names, f.Names[0].Name)
			}

			if childR, err = nm.parse(s); err != nil {
				return
			}

			// 支持 Result<TaskFlow> 这种输出
			if m.extend.template != "" && childR.kind == interfaceKind {
				m.extend.template = getSchema(m.extend.template)
				childR, err = newModel(m.filename, ast.NewIdent(m.extend.template), m.p, schemaExtend{}).parse(s)
				if err != nil {
					fmt.Println(err)
					return
				}
			}

			var (
				oneOfType string
				oneOfDesc string
			)

			if f.Tag != nil {
				var (
					required   bool
					tmpName    string
					mjsonNames []string
					ignore     bool
				)
				// if tmpName, childR.desc, childR.def, required, ignore, _ = parseTag(f.Tag.Value, childR.buildin); ignore {
				// 	continue // hanppens when `josn:"-"`
				// }
				if tmpName, mjsonNames, childR.desc, childR.def, required, oneOfType, oneOfDesc, ignore, _ = parseTag(f.Tag.Value, childR.buildin); ignore {
					continue // hanppens when `josn:"-"`
				}
				if tmpName != "" {
					names[0] = tmpName
				}
				if len(mjsonNames) > 0 {
					names = append(names, mjsonNames...)
				}
				if required {
					r.required = append(r.required, names...)
				}
			}

			// 处理oneOf
			if oneOfType != "" {
				oneOfType = getSchema(oneOfType)
				tmpR, err := newModel(m.filename, ast.NewIdent(oneOfType), m.p, schemaExtend{}).parse(s)
				if err != nil {
					return nil, err
				}

				tmpR.desc = oneOfDesc
				childR.oneof = append(childR.oneof, tmpR)
			}

			// must as a anonymous struct
			if nm.f == anonMemberFeature {
				hasKey := false
				for k1, v1 := range childR.items {
					for k2 := range r.items {
						if k1 == k2 {
							hasKey = true
							break
						}
					}
					if !hasKey {
						r.items[k1] = v1
						for _, v := range childR.required {
							if v == k1 {
								r.required = append(r.required, v)
								break
							}
						}
					}
				}
			} else {
				//r.items[name] = childR
				rawdesc := childR.desc
				for k, v := range names {
					if k == 0 && len(names) > 1 {
						recommend := names[1]
						tmpres := *childR
						tmpres.desc = "等同于 " + recommend + ", 未来会废弃，推荐使用 " + recommend + "。"
						tmpres.deprecated = true
						r.items[v] = &tmpres
						continue
					}

					childR.desc = rawdesc
					r.items[v] = childR
				}
			}

		}

		// 支持 TaskFlow+Pagezation 这种输出
		if len(m.extend.allof) > 0 {
			for _, v := range m.extend.allof {
				v = getSchema(v)
				cr, _ := newModel(m.filename, ast.NewIdent(v), m.p, schemaExtend{}).parse(s)
				if cr != nil {
					for k, _ := range cr.items {
						r.items[k] = cr.items[k]
					}
				}
			}
		}

		if m.f != anonStructFeature {
			// cache the result and definitions for swagger's schema
			ss, err := r.convertToSchema()
			if err != nil {
				return nil, err
			}

			if m.extend.template != "" {
				s.Components.Schemas[m.name+":"+m.extend.template] = ss
			} else {
				s.Components.Schemas[m.name] = ss
			}
		}
	}
	return
}

// cachedModels the cache of models
// Format:
//   model name -> import path and result
var cachedModels = map[string][]*kv{}

type kv struct {
	path string
	r    *result
}

type kind int

const (
	noneKind kind = iota
	innerKind
	arrayKind
	mapKind
	objectKind
	interfaceKind
)

func parseTag(tagStr, buildin string) (name string, mjsonNames []string, desc string, def interface{}, required bool, oneOfType string, oneOfDesc string, ignore bool, err error) {
	// parse tag for name
	stag := reflect.StructTag(strings.Trim(tagStr, "`"))
	// check jsonTag == "-"
	jsonTag := strings.Split(stag.Get("json"), ",")
	if len(jsonTag) != 0 {
		if jsonTag[0] == "-" {
			ignore = true
			return
		}
		name = jsonTag[0]
	}
	// swaggo:"(required),(desc),(default)"
	swaggoTag := stag.Get("swaggo")
	tmp := strings.Split(swaggoTag, ",")
	for k, v := range tmp {
		switch k {
		case 0:
			if v == "true" {
				required = true
			}
		case 1:
			desc = v
		case 2:
			if v != "" {
				def, err = str2RealType(v, buildin)
			}
		}
	}

	// mjson: "-,name1,name2,...,omitempty"
	mjsonTag := stag.Get("mjson")
	tmp = strings.Split(mjsonTag, ",")
	for _, v := range tmp {
		v = strings.TrimSpace(v)
		if v == "-" {
			ignore = true
			return
		}

		if v == "omitempty" || v == "" {
			continue
		}

		if v != name {
			mjsonNames = append(mjsonNames, v)
		}
	}

	// oneof:"(type),(desc)"
	oneOf := stag.Get("oneof")
	tmp = strings.Split(oneOf, ",")
	for k, v := range tmp {
		switch k {
		case 0:
			oneOfType = v
		case 1:
			oneOfDesc = v
		}
	}
	return
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
	// option.XXX from code.teambition.com/soa/go-lib/pkg/option
	"option.Interface": "object:",
	"option.ObjectID":  "string:",
	"option.ObjectIDs": "[]string:",
	"option.String":    "string:",
	"option.Strings":   "[]string:",
	"option.Time":      "string:date-time",
	"option.Number":    "number:int32",
	"option.Numbers":   "[]number:int32",
	"option.Bool":      "boolean:",
	"option.Bools":     "[]boolean:",
}
