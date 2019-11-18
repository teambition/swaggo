package parserv3

import (
	"errors"
	"fmt"

	"github.com/teambition/swaggo/swaggerv3"
)

type schemaExtend struct {
	template string   // 支持泛型，内嵌结构替换
	allof    []string // 支持allOf操作符
}

type result struct {
	kind       kind
	buildin    string      // golang type
	def        interface{} // default value
	desc       string
	ref        string
	item       *result
	required   []string
	deprecated bool
	items      map[string]*result

	sType   string // swagger type
	sFormat string // swagger format
	extend  schemaExtend
	oneof   []*result
}

func (r *result) convertToPathSchema() (*swaggerv3.PathSchema, error) {
	ss := &swaggerv3.PathSchema{}
	switch r.kind {
	case objectKind:
		r.parsePathSchema(ss)
	default:
		return nil, errors.New("result need object kind")
	}
	return ss, nil
}

func (r *result) convertToSchema() (*swaggerv3.Schema, error) {
	ss := &swaggerv3.Schema{}
	switch r.kind {
	case objectKind:
		r.parseSchema(ss)
	default:
		return nil, errors.New("result need object kind")
	}
	return ss, nil
}

func (r *result) parsePathSchema(ss *swaggerv3.PathSchema) {
	// NOTE:
	// schema description not support now
	// ss.Description = r.desc
	switch r.kind {
	case innerKind:
		ss.Type = r.sType
	case objectKind:
		if r.ref != "" {
			ss.Ref = r.ref
			return
		}
	case arrayKind:
		ss.Type = "array"
		ss.Items = &swaggerv3.PathSchema{}
		r.item.parsePathSchema(ss.Items)
	case mapKind:
		ss.Type = "object"
	case interfaceKind:
		ss.Type = "object"
	}
}

func (r *result) parseSchema(ss *swaggerv3.Schema) {
	// NOTE:
	// schema description not support now
	// ss.Description = r.desc
	switch r.kind {
	case innerKind:
		ss.Type = r.sType
		ss.Format = r.sFormat
		ss.Default = r.def
	case objectKind:
		//ss.Properties = r.items
		ss.Required = r.required
		if ss.Properties == nil {
			ss.Properties = make(map[string]*swaggerv3.Propertie)
		}
		for k, v := range r.items {
			sp := &swaggerv3.Propertie{}
			v.parsePropertie(sp)
			ss.Properties[k] = sp
		}
	case arrayKind:
		ss.Type = "array"
		ss.Items = &swaggerv3.Schema{}
		r.item.parseSchema(ss.Items)
	case mapKind:
		ss.AdditionalProperties = &swaggerv3.Propertie{}
		r.item.parsePropertie(ss.AdditionalProperties)
	case interfaceKind:
		ss.Type = "object"
	}
}

func (r *result) parsePropertie(sp *swaggerv3.Propertie) {
	property := &swaggerv3.Propertie{}
	switch r.kind {
	case innerKind:
		property.Default = r.def
		property.Type = r.sType
		property.Description = r.desc
		property.Deprecated = r.deprecated
	case arrayKind:
		property.Type = "array"
		property.Items = &swaggerv3.Propertie{}
		property.Description = r.desc
		property.Deprecated = r.deprecated
		r.item.parsePropertie(property.Items)
	case mapKind:
		property.Type = "object"
		property.Deprecated = r.deprecated
		property.AdditionalProperties = &swaggerv3.Propertie{}
		r.item.parsePropertie(property.AdditionalProperties)
	case objectKind:
		if r.ref != "" {
			property.Ref = r.ref
		} else {
			property.Description = r.desc
			property.Required = r.required
			property.Deprecated = r.deprecated
			if property.Properties == nil {
				property.Properties = make(map[string]*swaggerv3.Propertie)
			}
			for k, v := range r.items {
				tmpSp := &swaggerv3.Propertie{}
				v.parsePropertie(tmpSp)
				property.Properties[k] = tmpSp
			}
		}
	case interfaceKind:
		property.Type = "object"
		// TODO
	}

	if len(r.oneof) > 0 {
		sp.OneOf = []*swaggerv3.Propertie{}
		sp.OneOf = append(sp.OneOf, property)

		for _, v := range r.oneof {
			p := swaggerv3.Propertie{}
			v.parsePropertie(&p)
			sp.OneOf = append(sp.OneOf, &p)
		}

		return
	}

	sp.AdditionalProperties = property.AdditionalProperties
	sp.Default = property.Default
	sp.Deprecated = property.Deprecated
	sp.Description = property.Description
	sp.Example = property.Example
	sp.Items = property.Items
	sp.OneOf = property.OneOf
	sp.Properties = property.Properties
	sp.ReadOnly = property.ReadOnly
	sp.Ref = property.Ref
	sp.Required = property.Required
	sp.Title = property.Title
	sp.Type = property.Type

}
func (r *result) parseBodyParam(sp *swaggerv3.RequestBody) error {
	if sp.Content.ApplicationJSON.Schema == nil {
		sp.Content.ApplicationJSON.Schema = &swaggerv3.PathSchema{}
	}
	r.parsePathSchema(sp.Content.ApplicationJSON.Schema)
	return nil
}

func (r *result) parseParam(sp *swaggerv3.Parameter) error {
	if sp.Schema == nil {
		sp.Schema = &swaggerv3.ParameterSchema{}
	}
	switch r.kind {
	case innerKind:
		sp.Schema.Type = r.sType
	case arrayKind:
		sp.Schema.Type = "array"
		sp.Schema.Items = &swaggerv3.ParameterSchema{}
		if err := r.item.parseParamItem(sp.Schema.Items); err != nil {
			return err
		}
	default:
		// TODO
		// not support object and array in any value other than "body"
		return fmt.Errorf("not support(%d) in(%s) any value other than `body`", r.kind, sp.In)
	}
	return nil
}

func (r *result) parseParamItem(sp *swaggerv3.ParameterSchema) error {
	switch r.kind {
	case innerKind:
		sp.Type = r.sType
		sp.Format = r.sFormat
	case arrayKind:
		sp.Type = "array"
		sp.Items = &swaggerv3.ParameterSchema{}
		if err := r.item.parseParamItem(sp.Items); err != nil {
			return err
		}
	default:
		// TODO
		// param not support object, map, interface
		return fmt.Errorf("not support(%d) in any value other than `body`", r.kind)
	}
	return nil
}
