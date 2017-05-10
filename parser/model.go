package parser

import (
	"errors"
	"fmt"
	"go/ast"
	"reflect"
	"regexp"
	"strings"

	"github.com/teambition/swaggo/swagger"
)

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
	isMember bool // is a member of struct
}

func newModel(filename, schema string, p *pkg) *model {
	return &model{
		Expr:     ast.NewIdent(schema),
		filename: filename,
		p:        p,
	}
}

// raw the raw type of model
func (m *model) newModel(e ast.Expr) *model {
	return &model{
		Expr:     e,
		filename: m.filename,
		p:        m.p,
	}
}

// inhert the feature from other model
func (m *model) inhert(other *model) *model {
	m.f = other.f
	m.isMember = other.isMember
	return m
}

func (m *model) member() *model {
	m.isMember = true
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
func (m *model) parse(s *swagger.Swagger) (r *result, err error) {
	switch t := m.Expr.(type) {
	case *ast.StarExpr:
		return m.newModel(t.X).inhert(m).parse(s)
	case *ast.Ident, *ast.SelectorExpr:
		schema := fmt.Sprint(t)
		r = &result{}
		// []SomeStruct
		if strings.HasPrefix(schema, "[]") {
			schema = schema[2:]
			r.kind = arrayKind
			r.item, err = m.newModel(ast.NewIdent(schema)).parse(s)
			return
		}
		// map[string]SomeStruct
		if strings.HasPrefix(schema, "map[string]") {
			fmt.Println(schema)
			schema = schema[11:]
			r.kind = mapKind
			r.item, err = m.newModel(ast.NewIdent(schema)).parse(s)
			return
		}
		// &{foo Bar} to foo.Bar
		reInternalRepresentation := regexp.MustCompile("&\\{(\\w*) (\\w*)\\}")
		schema = string(reInternalRepresentation.ReplaceAll([]byte(schema), []byte("$1.$2")))
		// check if is basic type
		if swaggerType, ok := basicTypes[schema]; ok {
			r.kind = innerKind
			r.buildin = schema
			tmp := strings.Split(swaggerType, ":")
			r.sType = tmp[0]
			r.sFormat = tmp[1]
			return
		}
		if nm, err := m.p.findModelBySchema(m.filename, schema); err != nil {
			return nil, err
		} else {
			return nm.inhert(m).parse(s)
		}
	case *ast.ArrayType:
		r = &result{kind: arrayKind}
		r.item, err = m.newModel(t.Elt).inhert(m).parse(s)
	case *ast.MapType:
		r = &result{kind: mapKind}
		r.item, err = m.newModel(t.Value).inhert(m).parse(s)
	case *ast.InterfaceType:
		return &result{kind: interfaceKind}, nil
	case *ast.StructType:
		r = &result{kind: objectKind, items: map[string]*result{}}
		// anonymous struct
		// type A struct {
		//     B struct {}
		// }
		if m.isMember && m.f != anonMemberFeature {
			m.anonymousStruct()
		}
		var key string
		switch m.f {
		case anonStructFeature:
		default:
			// find schema cache
			// check if existed
			if s.Definitions == nil {
				s.Definitions = map[string]*swagger.Schema{}
			}

			key = m.name
			r.title = m.name
			if ips, ok := cachedModels[m.name]; ok {
				for k, v := range ips {
					if m.p.importPath == v.path {
						if k != 0 {
							key = fmt.Sprintf("%s_%d", m.name, k)
						}
						if m.f == anonMemberFeature {
							r = v.r
							return
						}
						break
					}
				}
			} else {
				cachedModels[m.name] = []*kv{}
			}
			if _, ok := s.Definitions[key]; ok {
				r.ref = "#/definitions/" + key
				return
			}
		}

		for _, f := range t.Fields.List {
			var (
				childR *result
				nm     = m.newModel(f.Type).member()
				name   string
			)

			if f.Names == nil {
				// anonymous member
				// type A struct {
				//     B
				//     C
				// }
				nm = nm.anonymousMember()
			}
			if childR, err = nm.parse(s); err != nil {
				return
			}

			if len(f.Names) != 0 {
				name = f.Names[0].Name
			}

			if f.Tag != nil {
				tmpName, desc, def, required, _ := parseTag(f.Tag.Value, childR.buildin)
				if tmpName != "" {
					name = tmpName
				}
				if required {
					r.required = append(r.required, name)
				}
				childR.desc = desc
				childR.def = def
			}

			// must as a anonymous struct
			if nm.f == anonMemberFeature {
				hasKey := false
				for k1, v1 := range childR.items {
					for k2, _ := range r.items {
						if k1 == k2 {
							hasKey = true
							break
						}
					}
					if !hasKey {
						r.items[k1] = v1
					}
				}
			} else {
				r.items[name] = childR
			}
		}

		if m.f != anonStructFeature {
			ss, err := r.convertToSchema()
			if err != nil {
				return nil, err
			}
			s.Definitions[key] = ss
			cachedModels[m.name] = append(cachedModels[m.name], &kv{m.p.importPath, r})
			if m.f != anonMemberFeature {
				r.ref = "#/definitions/" + key
			}
		}
	}
	return
}

// cachedModels
// model name -> import paths
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

type result struct {
	kind     kind
	title    string
	buildin  string      // golang type
	sType    string      // swagger type
	sFormat  string      // swagger format
	def      interface{} // default value
	desc     string
	ref      string
	item     *result
	required []string
	items    map[string]*result
}

func parseTag(tagStr, buildin string) (name, desc string, def interface{}, required bool, err error) {
	// parse tag for name
	stag := reflect.StructTag(strings.Trim(tagStr, "`"))
	// check jsonTag == "-"
	jsonTag := strings.Split(stag.Get("json"), ",")
	if len(jsonTag) != 0 && jsonTag[0] != "-" {
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
	return
}

func (r *result) convertToSchema() (*swagger.Schema, error) {
	ss := &swagger.Schema{}
	switch r.kind {
	case objectKind:
		r.parseSchema(ss)
	default:
		return nil, errors.New("result need object kind")
	}
	return ss, nil
}

func (r *result) parseSchema(ss *swagger.Schema) {
	ss.Title = r.title
	switch r.kind {
	case innerKind:
		ss.Type = r.sType
		ss.Format = r.sFormat
	case objectKind:
		ss.Type = "object"
		if r.ref != "" {
			ss.Ref = r.ref
			return
		}
		ss.Required = r.required
		if ss.Properties == nil {
			ss.Properties = make(map[string]*swagger.Propertie)
		}
		for k, v := range r.items {
			sp := &swagger.Propertie{}
			v.parsePropertie(sp)
			ss.Properties[k] = sp
		}
	case arrayKind:
		ss.Type = "array"
		ss.Items = &swagger.Schema{}
		r.item.parseSchema(ss.Items)
	case mapKind:
		ss.Type = "object"
		ss.AdditionalProperties = &swagger.Propertie{}
		r.item.parsePropertie(ss.AdditionalProperties)
	case interfaceKind:
		ss.Type = "object"
	}
}

func (r *result) parsePropertie(sp *swagger.Propertie) {
	switch r.kind {
	case innerKind:
		sp.Type = r.sType
		sp.Format = r.sFormat
	case arrayKind:
		sp.Type = "array"
		sp.Items = &swagger.Propertie{}
		r.item.parsePropertie(sp.Items)
	case mapKind:
		sp.Type = "object"
		sp.AdditionalProperties = &swagger.Propertie{}
		r.item.parsePropertie(sp.AdditionalProperties)
	case objectKind:
		sp.Type = "object"
		if r.ref != "" {
			sp.Ref = r.ref
			return
		}
		sp.Required = r.required
		if sp.Properties == nil {
			sp.Properties = make(map[string]*swagger.Propertie)
		}
		for k, v := range r.items {
			tmpSp := &swagger.Propertie{}
			v.parsePropertie(tmpSp)
			sp.Properties[k] = tmpSp
		}
	case interfaceKind:
		sp.Type = "object"
		// TODO
	}
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
		case innerKind:
			sp.Type = r.sType
			sp.Format = r.sFormat
		case arrayKind:
			sp.Type = "array"
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
	case innerKind:
		sp.Type = r.sType
		sp.Format = r.sFormat
	case arrayKind:
		sp.Type = "array"
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
