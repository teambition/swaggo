package parser

import (
	"go/ast"
	"strings"

	"github.com/teambition/swaggo/swagger"
)

type controller struct {
	doc      *ast.CommentGroup
	r        *resource
	noStruct bool // there is no controller struct
	name     string
	tagName  string
	filename string
	methods  []*method
}

func (ctrl *controller) parse(s *swagger.Swagger) (err error) {
	tag := swagger.Tag{}
	for _, c := range strings.Split(ctrl.doc.Text(), "\n") {
		switch {
		case tagTrimPrefixAndSpace(&c, ctrlName):
			ctrl.tagName = c
		case tagTrimPrefixAndSpace(&c, ctrlDesc):
			tag.Description = c
		case tagTrimPrefixAndSpace(&c, ctrlPrivate):
			// private controller
			return
		}
	}
	if ctrl.tagName == "" {
		if ctrl.noStruct {
			// TODO
			// means no controller struct for methods
			return
		}
		ctrl.tagName = ctrl.name
	}
	tag.Name = ctrl.tagName
	s.Tags = append(s.Tags, tag)

	for _, m := range ctrl.methods {
		if err = m.parse(s); err != nil {
			return
		}
	}
	return
}
