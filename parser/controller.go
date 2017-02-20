package parser

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"
	"unicode"

	"github.com/teambition/swaggo/swagger"
	"github.com/teambition/swaggo/utils"
)

type controller struct {
	*ast.TypeSpec
	doc     *ast.CommentGroup
	r       *Resource
	name    string
	methods []*method
}

func (ctrl *controller) parse(s *swagger.Swagger) (err error) {
	cName := ""
	tag := swagger.Tag{}
	for _, c := range strings.Split(ctrl.doc.Text(), "\n") {
		switch {
		case strings.HasPrefix(c, ctrlName):
			cName = strings.TrimSpace(c[len(ctrlName):])
		case strings.HasPrefix(c, ctrlDesc):
			tag.Description = strings.TrimSpace(c[len(ctrlDesc):])
		}
	}
	if cName == "" {
		cName = ctrl.name
	}
	tag.Name = cName
	s.Tags = append(s.Tags, tag)

	for _, method := range ctrl.methods {
		var routerPath, HTTPMethod string
		opt := swagger.Operation{
			Responses: make(map[string]swagger.Response),
		}
		for _, c := range strings.Split(method.Doc.Text(), "\n") {
			switch {
			case strings.HasPrefix(c, methodTitle):
				opt.OperationID = cName + "." + strings.TrimSpace(c[len(methodTitle):])
			case strings.HasPrefix(c, methodDesc):
				opt.Description = strings.TrimSpace(c[len(methodDesc):])
			case strings.HasPrefix(c, methodSummary):
				opt.Summary = strings.TrimSpace(c[len(methodSummary):])
			case strings.HasPrefix(c, methodSuccess):
				ss := strings.TrimSpace(c[len(methodSuccess):])
				rs := swagger.Response{}
				rs.Schema = &swagger.Schema{}
				respCode, pos := peekNextSplitString(ss)
				ss = strings.TrimSpace(ss[pos:])
				schemaName, pos := peekNextSplitString(ss)
				rs.Description = strings.TrimSpace(ss[pos:])
				if err = ctrl.r.parse(s, rs.Schema, method.filename, schemaName); err != nil {
					return
				}
				opt.Responses[respCode] = rs
			case strings.HasPrefix(c, methodParam):
				para := swagger.Parameter{}
				p := getparams(strings.TrimSpace(c[len(methodParam):]))
				if len(p) < 4 {
					err = fmt.Errorf("(%s.%s) comments %s shuold have 4 params at least", cName, method.name)
					return
				}
				para.Name = p[0]
				switch p[1] {
				case query:
					fallthrough
				case header:
					fallthrough
				case path:
					fallthrough
				case formData:
					fallthrough
				case body:
					break
				default:
					err = fmt.Errorf("(%s.%s) unknown param(%s). Maybe in(query, header, path, formData, body)", cName, method.name, para.Name)
					return
				}
				para.In = p[1]
				m := &swagger.Schema{}
				if err = ctrl.r.parse(s, m, method.filename, p[2]); err != nil {
					return
				}
				para.Schema = m
				switch len(p) {
				case 5:
					para.Required, _ = strconv.ParseBool(p[3])
					para.Description = strings.Trim(p[4], `" `)
				case 6:
					para.Default, _ = utils.Str2RealType(p[3], para.Type)
					para.Required, _ = strconv.ParseBool(p[4])
					para.Description = strings.Trim(p[5], `" `)
				default:
					para.Description = strings.Trim(p[3], `" `)
				}
				opt.Parameters = append(opt.Parameters, para)
			case strings.HasPrefix(c, methodFailure):
				rs := swagger.Response{}
				st := strings.TrimSpace(c[len(methodFailure):])
				var cd []rune
				var start bool
				for i, s := range st {
					if unicode.IsSpace(s) {
						if start {
							rs.Description = strings.TrimSpace(st[i+1:])
							break
						} else {
							continue
						}
					}
					start = true
					cd = append(cd, s)
				}
				opt.Responses[string(cd)] = rs
			case strings.HasPrefix(c, methodDeprecated):
				opt.Deprecated, _ = strconv.ParseBool(strings.TrimSpace(c[len(methodDeprecated):]))
			case strings.HasPrefix(c, methodAccept):
				accepts := strings.Split(strings.TrimSpace(strings.TrimSpace(c[len(methodAccept):])), ",")
				for _, a := range accepts {
					switch a {
					case json:
						opt.Consumes = append(opt.Consumes, ajson)
					case xml:
						opt.Consumes = append(opt.Consumes, axml)
					case plain:
						opt.Consumes = append(opt.Consumes, tplain)
					case html:
						opt.Consumes = append(opt.Consumes, thtml)
					}
				}
			case strings.HasPrefix(c, methodProduce):
				produces := strings.Split(strings.TrimSpace(strings.TrimSpace(c[len(methodProduce):])), ",")
				for _, p := range produces {
					switch p {
					case json:
						opt.Produces = append(opt.Produces, ajson)
					case xml:
						opt.Produces = append(opt.Produces, axml)
					case plain:
						opt.Produces = append(opt.Produces, tplain)
					case html:
						opt.Produces = append(opt.Produces, thtml)
					}
				}
			case strings.HasPrefix(c, methodRouter):
				// @Router / [post]
				elements := strings.Split(strings.TrimSpace(c[len(methodRouter):]), " ")
				if len(elements) == 0 {
					return fmt.Errorf("method(%s) should has Router information", method.name)
				}
				if len(elements) == 1 {
					HTTPMethod = "GET"
					routerPath = elements[0]
				} else {
					HTTPMethod = strings.ToUpper(elements[0])
					routerPath = elements[1]
				}
			}
		}
		if routerPath != "" {
			if s.Paths == nil {
				s.Paths = map[string]*swagger.Item{}
			}
			item, ok := s.Paths[routerPath]
			if !ok {
				item = &swagger.Item{}
			}
			switch HTTPMethod {
			case "GET":
				item.Get = &opt
			case "POST":
				item.Post = &opt
			case "PUT":
				item.Put = &opt
			case "PATCH":
				item.Patch = &opt
			case "DELETE":
				item.Delete = &opt
			case "HEAD":
				item.Head = &opt
			case "OPTIONS":
				item.Options = &opt
			}
			if s.Paths == nil {
			}
			s.Paths[routerPath] = item
		}
	}
	return
}

func isDocComments(comments *ast.CommentGroup) bool {
	for _, c := range strings.Split(comments.Text(), "\n") {
		if strings.HasPrefix(c, docPrefix) {
			return true
		}
	}
	return false
}

func peekNextSplitString(ss string) (s string, spacePos int) {
	spacePos = strings.IndexFunc(ss, unicode.IsSpace)
	if spacePos < 0 {
		s = ss
		spacePos = len(ss)
	} else {
		s = strings.TrimSpace(ss[:spacePos])
	}
	return
}

// analisys params return []string
// @Param	query		form	 string	true		"The email for login"
// [query form string true "The email for login"]
func getparams(str string) []string {
	var s []rune
	var j int
	var start bool
	var r []string
	var quoted int8
	for _, c := range []rune(str) {
		if unicode.IsSpace(c) && quoted == 0 {
			if !start {
				continue
			} else {
				start = false
				j++
				r = append(r, string(s))
				s = make([]rune, 0)
				continue
			}
		}

		start = true
		if c == '"' {
			quoted ^= 1
			continue
		}
		s = append(s, c)
	}
	if len(s) > 0 {
		r = append(r, string(s))
	}
	return r
}
