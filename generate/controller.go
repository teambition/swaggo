package generate

import (
	"fmt"
	"go/ast"
	"go/token"
	"strconv"
	"strings"
	"unicode"

	"github.com/teambition/swaggo/pkg"
	"github.com/teambition/swaggo/swagger"
	"github.com/teambition/swaggo/utils"
)

const (
	ajson  = "application/json"
	axml   = "application/xml"
	aplain = "text/plain"
	ahtml  = "text/html"
)

const (
	docPrefix = "@"
	// controller tag
	_ctrlName = "@Name"
	_ctrlDesc = "@Description"
	// controller item
	_itemTitle      = "@Title"
	_itemDesc       = "@Description"
	_itemSummary    = "@Summary"
	_itemSuccess    = "@Success"
	_itemParam      = "@Param"
	_itemFailure    = "@Failure"
	_itemDeprecated = "@Deprecated"
	_itemAccept     = "@Accept"
	_itemProduce    = "@Produce"
	_itemRouter     = "@Router"
)

type method struct {
	*ast.FuncDecl
	name     string
	filename string
}

type controller struct {
	*ast.TypeSpec
	doc     *ast.CommentGroup
	pkg     *CtrlPackage
	name    string
	methods []*method
}

func (ctrl *controller) parse(s *swagger.Swagger) (err error) {
	ctrlName := ""
	tag := swagger.Tag{}
	for _, c := range strings.Split(ctrl.Doc.Text(), "\n") {
		switch {
		case strings.HasPrefix(c, _ctrlName):
			ctrlName = strings.TrimSpace(c[len(_ctrlName):])
		case strings.HasPrefix(c, _ctrlDesc):
			tag.Description = strings.TrimSpace(c[len(_ctrlDesc):])
		}
	}
	if ctrlName == "" {
		ctrlName = ctrl.name
	}
	tag.Name = ctrlName
	s.Tags = append(s.Tags, tag)

	var routerPath, HTTPMethod string
	opt := swagger.Operation{
		Responses: make(map[string]swagger.Response),
	}
	for _, method := range ctrl.methods {
		for _, c := range strings.Split(method.Doc.Text(), "\n") {
			switch {
			case strings.HasPrefix(c, _itemTitle):
				opt.OperationID = ctrlName + "." + strings.TrimSpace(c[len(_itemTitle):])
			case strings.HasPrefix(c, _itemDesc):
				opt.Description = strings.TrimSpace(c[len(_itemDesc):])
			case strings.HasPrefix(c, _itemSummary):
				opt.Summary = strings.TrimSpace(c[len(_itemSummary):])
			case strings.HasPrefix(c, _itemSuccess):
				ss := strings.TrimSpace(c[len(_itemSuccess):])
				rs := swagger.Response{}
				m := swagger.Schema{}
				respCode, pos := peekNextSplitString(ss)
				ss = strings.TrimSpace(ss[pos:])
				schemaName, pos := peekNextSplitString(ss)
				rs.Description = strings.TrimSpace(ss[pos:])
				if err = ctrl.pkg.Parse(s, &m, method.filename, schemaName); err != nil {
					return
				}
				opt.Responses[respCode] = rs
			case strings.HasPrefix(c, _itemParam):
				para := swagger.Parameter{}
				p := getparams(strings.TrimSpace(c[len(_itemParam):]))
				if len(p) < 4 {
					err = fmt.Errorf("(%s.%s) comments %s shuold have 4 params at least", ctrlName, method.name)
					return
				}
				para.Name = p[0]
				switch p[1] {
				case "query":
					fallthrough
				case "header":
					fallthrough
				case "path":
					fallthrough
				case "formData":
					fallthrough
				case "body":
					break
				default:
					err = fmt.Errorf("(%s.%s) unknown param(%s). Maybe in(query, header, path, formData, body)", ctrlName, method.name)
					return
				}
				para.In = p[1]
				m := &swagger.Schema{}
				if err = ctrl.pkg.Parse(s, m, method.filename, p[2]); err != nil {
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
			case strings.HasPrefix(c, _itemFailure):
				rs := swagger.Response{}
				st := strings.TrimSpace(c[len(_itemFailure):])
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
			case strings.HasPrefix(c, _itemDeprecated):
				opt.Deprecated, _ = strconv.ParseBool(strings.TrimSpace(c[len(_itemDeprecated):]))
			case strings.HasPrefix(c, _itemAccept):
				accepts := strings.Split(strings.TrimSpace(strings.TrimSpace(c[len(_itemAccept):])), ",")
				for _, a := range accepts {
					switch a {
					case "json":
						opt.Consumes = append(opt.Consumes, ajson)
					case "xml":
						opt.Consumes = append(opt.Consumes, axml)
					case "plain":
						opt.Consumes = append(opt.Consumes, aplain)
					case "html":
						opt.Consumes = append(opt.Consumes, ahtml)
					}
				}
			case strings.HasPrefix(c, _itemProduce):
				produces := strings.Split(strings.TrimSpace(strings.TrimSpace(c[len(_itemProduce):])), ",")
				for _, p := range produces {
					switch p {
					case "json":
						opt.Produces = append(opt.Produces, ajson)
					case "xml":
						opt.Produces = append(opt.Produces, axml)
					case "plain":
						opt.Produces = append(opt.Produces, aplain)
					case "html":
						opt.Produces = append(opt.Produces, ahtml)
					}
				}
			case strings.HasPrefix(c, _itemRouter):
				// @Router / [post]
				elements := strings.Split(strings.TrimSpace(c[len(_itemRouter):]), " ")
				if len(elements) == 0 {
					return fmt.Errorf("method(%s) should has Router information", method.Name.String())
				}
				routerPath = elements[0]
				if len(elements) > 1 {
					HTTPMethod = strings.ToUpper(strings.Trim(elements[1], "[]"))
				} else {
					HTTPMethod = "GET"
				}
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
	return
}

type CtrlPackage struct {
	*pkg.Package
	// ctrl name -> ctrl
	// ctrlName -> ctrl
	controllers map[string]*controller
}

func NewCtrlPackage(importPath string, filter func(string) bool) (*CtrlPackage, error) {
	p, err := pkg.NewPackage("_", importPath, filter)
	if err != nil {
		return nil, err
	}

	cp := &CtrlPackage{
		Package:     p,
		controllers: map[string]*controller{},
	}
	for filename, f := range p.Files {
		for _, d := range f.Decls {
			switch specDecl := d.(type) {
			case *ast.FuncDecl:
				if specDecl.Recv != nil && len(specDecl.Recv.List) != 0 {
					if t, ok := specDecl.Recv.List[0].Type.(*ast.StarExpr); ok {
						if isDocComments(specDecl.Doc) {
							ctrlName := fmt.Sprint(t.X)
							if ctrl, ok := cp.controllers[ctrlName]; !ok {
								cp.controllers[ctrlName] = &controller{
									pkg: cp,
									methods: []*method{
										&method{
											FuncDecl: specDecl,
											filename: filename,
											name:     specDecl.Name.Name,
										},
									},
								}

							} else {
								ctrl.methods = append(ctrl.methods, &method{
									FuncDecl: specDecl,
									filename: filename,
									name:     specDecl.Name.Name,
								})
							}
						}
					}
				}
			case *ast.GenDecl:
				if specDecl.Tok == token.TYPE {
					for _, s := range specDecl.Specs {
						t := s.(*ast.TypeSpec)
						switch t.Type.(type) {
						case *ast.StructType:
							if isDocComments(specDecl.Doc) {
								ctrlName := t.Name.String()
								if ctrl, ok := cp.controllers[ctrlName]; !ok {
									cp.controllers[ctrlName] = &controller{
										TypeSpec: t,
										doc:      specDecl.Doc,
										pkg:      cp,
										name:     t.Name.Name,
										methods:  []*method{},
									}
								} else if ctrl.TypeSpec == nil {
									ctrl.TypeSpec = t
									ctrl.doc = specDecl.Doc
									ctrl.name = t.Name.Name
								}
							}
						}
					}
				}
			}
		}
	}
	return cp, nil
}

func (c *CtrlPackage) Run(s *swagger.Swagger) error {
	// parse controllers
	for _, ctrl := range c.controllers {
		if err := ctrl.parse(s); err != nil {
			return err
		}
	}
	return nil

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
