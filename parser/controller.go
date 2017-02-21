package parser

import (
	"fmt"
	"go/ast"
	"strconv"
	"strings"
	"unicode"

	"github.com/teambition/swaggo/swagger"
)

type controller struct {
	*ast.TypeSpec
	doc     *ast.CommentGroup
	r       *resource
	name    string
	methods []*method
}

func (ctrl *controller) parse(s *swagger.Swagger) (err error) {
	cName := ""
	tag := swagger.Tag{}
	for _, c := range strings.Split(ctrl.doc.Text(), "\n") {
		switch {
		case tagTrimPrefixAndSpace(&c, ctrlName):
			cName = c
		case tagTrimPrefixAndSpace(&c, ctrlDesc):
			tag.Description = c
		case tagTrimPrefixAndSpace(&c, ctrlPrivate):
			// private controller
			return
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
		private := false
		for _, c := range strings.Split(method.Doc.Text(), "\n") {
			switch {
			case tagTrimPrefixAndSpace(&c, methodPrivate):
				private = true
			case tagTrimPrefixAndSpace(&c, methodTitle):
				opt.OperationID = cName + "." + c
			case tagTrimPrefixAndSpace(&c, methodDesc):
				opt.Description = c
			case tagTrimPrefixAndSpace(&c, methodSummary):
				if opt.Summary != "" {
					opt.Summary = opt.Summary + "\n" + c
				} else {
					opt.Summary = c
				}
			case tagTrimPrefixAndSpace(&c, methodSuccess):
				rs := swagger.Response{}
				rs.Schema = &swagger.Schema{}
				respCode, pos := peekNextSplitString(c)
				c = strings.TrimSpace(c[pos:])
				schemaName, pos := peekNextSplitString(c)
				rs.Description = strings.TrimSpace(c[pos:])
				if err = ctrl.r.parseSchema(s, rs.Schema, method.filename, schemaName); err != nil {
					return
				}
				opt.Responses[respCode] = rs
			case tagTrimPrefixAndSpace(&c, methodParam):
				para := swagger.Parameter{}
				p := getparams(c)
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
				if err = ctrl.r.parseSchema(s, m, method.filename, p[2]); err != nil {
					return
				}
				para.Schema = m
				switch len(p) {
				case 5:
					para.Required, _ = strconv.ParseBool(p[3])
					para.Description = strings.Trim(p[4], `" `)
				case 6:
					para.Default, _ = str2RealType(p[3], para.Type)
					para.Required, _ = strconv.ParseBool(p[4])
					para.Description = strings.Trim(p[5], `" `)
				default:
					para.Description = strings.Trim(p[3], `" `)
				}
				opt.Parameters = append(opt.Parameters, para)
			case tagTrimPrefixAndSpace(&c, methodFailure):
				rs := swagger.Response{}
				var cd []rune
				var start bool
				for i, s := range c {
					if unicode.IsSpace(s) {
						if start {
							rs.Description = strings.TrimSpace(c[i+1:])
							break
						} else {
							continue
						}
					}
					start = true
					cd = append(cd, s)
				}
				opt.Responses[string(cd)] = rs
			case tagTrimPrefixAndSpace(&c, methodDeprecated):
				opt.Deprecated, _ = strconv.ParseBool(c)
			case tagTrimPrefixAndSpace(&c, methodAccept):
				for _, a := range strings.Split(c, ",") {
					switch a {
					case jsonType:
						opt.Consumes = append(opt.Consumes, appJson)
					case xmlType:
						opt.Consumes = append(opt.Consumes, appXml)
					case plainType:
						opt.Consumes = append(opt.Consumes, textPlain)
					case htmlType:
						opt.Consumes = append(opt.Consumes, textHtml)
					}
				}
			case tagTrimPrefixAndSpace(&c, methodProduce):
				for _, p := range strings.Split(c, ",") {
					switch p {
					case jsonType:
						opt.Produces = append(opt.Produces, appJson)
					case xmlType:
						opt.Produces = append(opt.Produces, appXml)
					case plainType:
						opt.Produces = append(opt.Produces, textPlain)
					case htmlType:
						opt.Produces = append(opt.Produces, textHtml)
					}
				}
			case tagTrimPrefixAndSpace(&c, methodRouter):
				// @Router / [post]
				elements := strings.Split(c, " ")
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
		if routerPath != "" && !private {
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
