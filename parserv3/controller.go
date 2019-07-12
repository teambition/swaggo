package parserv3

import (
	"go/ast"
	"strings"

	"github.com/teambition/swaggo/swaggerv3"
)

type controller struct {
	doc      *ast.CommentGroup
	r        *resource
	noStruct bool // there is no controller struct
	// struct name
	name string
	// tag name of comment
	tagName  string
	filename string
	methods  []*method
}

func (ctrl *controller) parse(s *swaggerv3.Swagger) (err error) {
	tag := swaggerv3.Tag{}
	for _, c := range strings.Split(ctrl.doc.Text(), "\n") {
		switch {
		case tagTrimPrefixAndSpace(&c, ctrlName):
			ctrl.tagName = c
		case tagTrimPrefixAndSpace(&c, ctrlDesc):
			if tag.Description != "" {
				tag.Description += "<br>" + c
			} else {
				tag.Description = c
			}
		}
	}
	if ctrl.tagName == "" {
		if ctrl.noStruct {
			// TODO
			// means no controller struct for methods
			return
		}
		// 如果没有显式指定名称，则用struct name
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
