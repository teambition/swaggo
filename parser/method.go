package parser

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"

	"github.com/teambition/swaggo/swagger"
)

// method the method of controllor
type method struct {
	doc      *ast.CommentGroup
	name     string // function name
	filename string // where it is
	ctrl     *controller
}

func (m *method) prettyErr(format string, e ...interface{}) error {
	f := ""
	if m.ctrl.name != "" {
		f = fmt.Sprintf("(%s:%s.%s) %s", m.filename, m.ctrl.name, m.name, format)
	} else {
		f = fmt.Sprintf("(%s:%s) %s", m.filename, m.name, format)
	}
	return fmt.Errorf(f, e...)
}

func (m *method) parse(s *swagger.Swagger) (err error) {
	var routerPath, HTTPMethod string
	tagName := m.ctrl.tagName
	opt := swagger.Operation{
		Responses: make(map[string]swagger.Response),
		Tags:      []string{tagName},
	}
	private := false
	for _, c := range strings.Split(m.doc.Text(), "\n") {
		switch {
		case tagTrimPrefixAndSpace(&c, methodPrivate):
			private = true
			break
		case tagTrimPrefixAndSpace(&c, methodTitle):
			opt.OperationID = tagName + "." + c
		case tagTrimPrefixAndSpace(&c, methodDesc):
			if opt.Description != "" {
				opt.Description += "\n" + c
			} else {
				opt.Description = c
			}
		case tagTrimPrefixAndSpace(&c, methodSummary):
			if opt.Summary != "" {
				opt.Summary += "\n" + c
			} else {
				opt.Summary = c
			}
		case tagTrimPrefixAndSpace(&c, methodParam):
			para := swagger.Parameter{}
			p := getparams(c)
			if len(p) < 4 {
				err = m.prettyErr("comments %s shuold have 4 params at least", c)
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
			case form:
				fallthrough
			case body:
				break
			default:
				err = m.prettyErr("unknown param(%s) type(%s), type must in(query, header, path, form, body)", p[0], p[1])
				return
			}
			para.In = p[1]
			if err = m.ctrl.r.parseParam(s, &para, m.filename, p[2]); err != nil {
				return
			}
			for idx, v := range p {
				switch idx {
				case 3:
					// required
					if v != "-" {
						para.Required, _ = strconv.ParseBool(v)
					}
				case 4:
					// description
					para.Description = strings.Trim(v, `" `)
				case 5:
					// default value
					if v != "-" {
						if para.Default, err = str2RealType(strings.Trim(v, `" `), p[2]); err != nil {
							err = m.prettyErr("parse default value of param(%s) type(%s) error(%v)", p[0], p[2], err)
							return
						}
					}
				}
			}
			opt.Parameters = append(opt.Parameters, para)
		case tagTrimPrefixAndSpace(&c, methodSuccess), tagTrimPrefixAndSpace(&c, methodFailure):
			sr := swagger.Response{Schema: &swagger.Schema{}}
			p := getparams(c)
			if len(p) != 3 {
				err = m.prettyErr("response (%s) format error, need(code, type, description)", c)
				return
			}
			if p[1] != "-" {
				if err = m.ctrl.r.parseSchema(s, sr.Schema, m.filename, p[1]); err != nil {
					return
				}
			}
			sr.Description = p[2]
			opt.Responses[p[0]] = sr
		case tagTrimPrefixAndSpace(&c, methodDeprecated):
			opt.Deprecated, _ = strconv.ParseBool(c)
		case tagTrimPrefixAndSpace(&c, methodConsumes):
			opt.Consumes = contentTypeByDoc(c)
		case tagTrimPrefixAndSpace(&c, methodProduces):
			opt.Produces = contentTypeByDoc(c)
		case tagTrimPrefixAndSpace(&c, methodRouter):
			// @Router / [post]
			elements := strings.Split(c, " ")
			if len(elements) == 0 {
				return m.prettyErr("should has Router information")
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

	if routerPath != "" && !private {
		// check body count
		hasBody := false
		for _, v := range opt.Parameters {
			if v.In == body {
				if v.Name != body {
					fmt.Println("[Warnning] ", m.prettyErr("body-type parameter(%s)'s name shuold be `body`", v.Name))
				}
				if hasBody {
					fmt.Println("[Warnning] ", m.prettyErr("has more than one body-type parameter, not all body works"))
					break
				} else {
					hasBody = true
				}
			}
		}
		if opt.OperationID == "" {
			opt.OperationID = tagName + "." + m.name
		}
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
	return
}
