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

// parse Parse the api method annotations
func (m *method) parse(s *swagger.Swagger) (err error) {
	var routerPath, HTTPMethod string
	tagName := m.ctrl.tagName
	opt := swagger.Operation{
		Responses: make(map[string]*swagger.Response),
		Tags:      []string{tagName},
	}
	private := false
	for _, c := range strings.Split(m.doc.Text(), "\n") {
		switch {
		case tagTrimPrefixAndSpace(&c, methodPrivate):
			if !devMode {
				private = true
				break
			}
		case tagTrimPrefixAndSpace(&c, methodTitle):
			opt.OperationID = tagName + "." + c
		case tagTrimPrefixAndSpace(&c, methodDesc):
			if opt.Description != "" {
				opt.Description += "<br>" + c
			} else {
				opt.Description = c
			}
		case tagTrimPrefixAndSpace(&c, methodSummary):
			if opt.Summary != "" {
				opt.Summary += "<br>" + c
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
			para.In = paramType[p[1]]
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
			opt.Parameters = append(opt.Parameters, &para)
		case tagTrimPrefixAndSpace(&c, methodSuccess), tagTrimPrefixAndSpace(&c, methodFailure):
			sr := &swagger.Response{}
			p := getparams(c)
			respCode := ""
			for idx, v := range p {
				switch idx {
				case 0:
					respCode = v
				case 1:
					if v != "-" {
						sr.Schema = &swagger.Schema{}
						if err = m.ctrl.r.parseSchema(s, sr.Schema, m.filename, v); err != nil {
							return
						}
					}
				case 2:
					sr.Description = v
				default:
					err = m.prettyErr("response (%s) format error, need(code, type, description)", c)
					return
				}
			}
			opt.Responses[respCode] = sr
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
		m.paramCheck(&opt)
		if s.Paths == nil {
			s.Paths = map[string]*swagger.Item{}
		}
		item, ok := s.Paths[routerPath]
		if !ok {
			item = &swagger.Item{}
		}
		var oldOpt *swagger.Operation
		switch HTTPMethod {
		case "GET":
			oldOpt = item.Get
			item.Get = &opt
		case "POST":
			oldOpt = item.Post
			item.Post = &opt
		case "PUT":
			oldOpt = item.Put
			item.Put = &opt
		case "PATCH":
			oldOpt = item.Patch
			item.Patch = &opt
		case "DELETE":
			oldOpt = item.Delete
			item.Delete = &opt
		case "HEAD":
			oldOpt = item.Head
			item.Head = &opt
		case "OPTIONS":
			oldOpt = item.Options
			item.Options = &opt
		}
		if oldOpt != nil {
			fmt.Println("[Warnning]", m.prettyErr("router(%s %s) has existed in controller(%s)", HTTPMethod, routerPath, oldOpt.Tags[0]))
		}
		s.Paths[routerPath] = item
	}
	return
}

// paramCheck Verify the validity of parametes
func (m *method) paramCheck(opt *swagger.Operation) {
	// swagger ui url (unique)
	if opt.OperationID == "" {
		opt.OperationID = m.ctrl.tagName + "." + m.name
	}

	hasFile, hasBody, hasForm, bodyWarn := false, false, false, false
	for _, v := range opt.Parameters {
		if v.Type == "file" && !hasFile {
			hasFile = true
		}
		switch v.In {
		case paramType[form]:
			hasForm = true
		case paramType[body]:
			if hasBody {
				if !bodyWarn {
					fmt.Println("[Warnning]", m.prettyErr("more than one body-type existed in this method"))
					bodyWarn = true
				}
			} else {
				hasBody = true
			}
		case paramType[path]:
			if !v.Required {
				// path-type parameter must be required
				v.Required = true
				fmt.Println("[Warnning]", m.prettyErr("path-type parameter(%s) must be required", v.Name))
			}
		}
	}
	if hasBody && hasForm {
		fmt.Println("[Warnning]", m.prettyErr("body-type and form-type cann't coexist"))
	}
	// If type is "file", the consumes MUST be
	// either "multipart/form-data", " application/x-www-form-urlencoded"
	// or both and the parameter MUST be in "formData".
	if hasFile {
		if hasBody {
			fmt.Println("[Warnning]", m.prettyErr("file-data-type and body-type cann't coexist"))
		}
		if !(len(opt.Consumes) == 0 || subset(opt.Consumes, []string{contentType[formType], contentType[formDataType]})) {
			fmt.Println("[Warnning]", m.prettyErr("file-data-type existed and this api's consumes must in(form, formData)"))
		}
	}
}
