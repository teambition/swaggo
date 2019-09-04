package parserv3

import (
	"fmt"
	"go/ast"
	"log"
	"strconv"
	"strings"

	"github.com/teambition/swaggo/swaggerv3"
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
func (m *method) parse(s *swaggerv3.Swagger) (err error) {
	var routerPath, HTTPMethod string
	tagName := m.ctrl.tagName
	opt := swaggerv3.Operation{
		Responses: make(map[string]*swaggerv3.Response),
		Tags:      []string{tagName},
	}
	for _, c := range strings.Split(m.doc.Text(), "\n") {
		switch {
		case tagTrimPrefixAndSpace(&c, methodRouter):
			// @Router get [post]
			elements := strings.Split(c, " ")
			if len(elements) < 2 {
				return m.prettyErr("should has HTTPMethod and Router information")
			}
			HTTPMethod = strings.ToUpper(elements[0])
			routerPath = elements[1]
		case tagTrimPrefixAndSpace(&c, methodDeprecated):
			opt.Deprecated, _ = strconv.ParseBool(c)
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
		case tagTrimPrefixAndSpace(&c, methodPermission):
			// @Permission member get
			elements := strings.Split(c, " ")
			if len(elements) == 0 {
				return m.prettyErr("should has Permission information")
			}
			permission := swaggerv3.Permission{
				Resource: elements[0],
				Action:   elements[1],
			}
			opt.Permissions = append(opt.Permissions, permission)
		case tagTrimPrefixAndSpace(&c, methodSuccess), tagTrimPrefixAndSpace(&c, methodFailure):
			sr := &swaggerv3.Response{}
			p := getparams(c)
			respCode := ""
			for idx, v := range p {
				switch idx {
				case 0: // code
					respCode = v
				case 1: // type
					if v != "-" {
						sr.Content = swaggerv3.Content{
							ApplicationJSON: swaggerv3.ApplicationJSON{
								Schema: &swaggerv3.PathSchema{},
							},
						}
						if err = m.ctrl.r.parseSchema(s, m.filename, v, sr.Content.ApplicationJSON.Schema); err != nil {
							return
						}
					}
				case 2: // Description
					sr.Description = v
				default:
					err = m.prettyErr("response (%s) format error, need(code, type, description)", c)
					return
				}
			}
			opt.Responses[respCode] = sr
		case tagTrimPrefixAndSpace(&c, methodParam):
			err = m.handleMethodParam(s, &opt, c)
			if err != nil {
				return
			}
		}
	}
	if routerPath == "" {
		return
	}
	m.paramCheck(&opt)
	if s.Paths == nil {
		s.Paths = map[string]swaggerv3.PathItem{}
	}
	item, ok := s.Paths[routerPath]
	if !ok {
		item = swaggerv3.PathItem{}
	}
	var oldOpt *swaggerv3.Operation
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
		log.Println("[Warning]", m.prettyErr("router(%s %s) has existed in controller(%s)", HTTPMethod, routerPath, oldOpt.Tags[0]))
	}
	s.Paths[routerPath] = item

	return
}
func (m *method) handleMethodParam(s *swaggerv3.Swagger, opt *swaggerv3.Operation, c string) (err error) {
	para := swaggerv3.Parameter{}
	p := getparams(c)
	if len(p) < 4 {
		err = m.prettyErr("comments %s shuold have 4 params at least", c)
		return
	}
	para.Name = p[0]
	if param, ok := paramType[p[1]]; ok {
		para.In = param
	} else {
		log.Printf("unknown param(%s) type(%s), type must in(query, header, path, form, body) \n", p[0], p[1])
		return
	}
	if para.In == "body" {
		// @Param body body subpackage.SimpleStructure true ""
		body := swaggerv3.RequestBody{}
		if err = m.ctrl.r.parseBodyParam(s, &body, m.filename, p[2]); err != nil {
			return
		}
		for idx, v := range p {
			switch idx {
			case 3:
				// required
				if v != "-" {
					body.Required, _ = strconv.ParseBool(v)
				}
			case 4:
				// description
				body.Description = strings.Trim(v, `" `)
			}
		}
		opt.RequestBody = &body
	} else {
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
			}
		}
		opt.Parameters = append(opt.Parameters, &para)
	}
	return
}

// paramCheck Verify the validity of parametes
func (m *method) paramCheck(opt *swaggerv3.Operation) {
	hasFile, hasBody, hasForm, bodyWarn := false, false, false, false
	for _, v := range opt.Parameters {
		if v.Schema.Type == "file" && !hasFile {
			hasFile = true
		}
		switch v.In {

		case paramType[body]:
			if hasBody {
				if !bodyWarn {
					log.Println("[Warning]", m.prettyErr("more than one body-type existed in this method"))
					bodyWarn = true
				}
			} else {
				hasBody = true
			}
		case paramType[path]:
			if !v.Required {
				// path-type parameter must be required
				v.Required = true
				log.Println("[Warning]", m.prettyErr("path-type parameter(%s) must be required", v.Name))
			}
		}
	}
	if hasBody && hasForm {
		log.Println("[Warning]", m.prettyErr("body-type and form-type cann't coexist"))
	}
	// If type is "file", the consumes MUST be
	// either "multipart/form-data", " application/x-www-form-urlencoded"
	// or both and the parameter MUST be in "formData".
	if hasFile {
		if hasBody {
			log.Println("[Warning]", m.prettyErr("file-data-type and body-type cann't coexist"))
		}
		if !(len(opt.Consumes) == 0 || subset(opt.Consumes, []string{contentType[formType], contentType[formDataType]})) {
			log.Println("[Warning]", m.prettyErr("file-data-type existed and this api's consumes must in(form, formData)"))
		}
	}
}
