package parser

import (
	"fmt"
	"go/ast"
	"reflect"
	"strings"

	"github.com/teambition/swaggo/swagger"
)

type model struct {
	*ast.TypeSpec
	name     string // model struct name
	filename string // in which file
	p        *pkg
}

// parseModel
func (m *model) parse(s *swagger.Swagger, e ast.Expr) (r *result, err error) {
	switch t := e.(type) {
	case *ast.StarExpr:
		return m.parse(s, t.X)
	case *ast.Ident:
		if swaggerType, ok := basicTypes[t.Name]; ok {
			return &result{
				kind:    innerType,
				swagger: swaggerType,
			}, nil
		}
		if nm, err := m.p.findModelBySchema(m.filename, t.Name); err != nil {
			return nil, err
		} else {
			return nm.parse(s, nm.Type)
		}
	case *ast.SelectorExpr:
		schema := fmt.Sprint(t)
		schema = strings.Replace(schema, " ", ".", -1)
		schema = strings.Replace(schema, "&", "", -1)
		schema = strings.Replace(schema, "{", "", -1)
		schema = strings.Replace(schema, "}", "", -1)
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
				// doc tag include
				// if the filed is a integer we can set: default:`123`
				defautlValue := stag.Get("default")
				if defautlValue != "" {
					if sp.Default, err = str2RealType(defautlValue, r.buildin); err != nil {
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
						ss.Required = append(ss.Required, name)
					}
					if desc := stag.Get("desc"); desc != "" {
						sp.Description = desc
					}
					ss.Properties[name] = sp
				}
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
	innerType  = "inner"
	arrayType  = "array"
	mapType    = "map"
	objectType = "object"
)

type result struct {
	kind    string
	buildin string  // buildin type
	swagger string  // kind == inner
	ref     string  // kind == object
	item    *result // kind in (array, map)
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
	}
}
