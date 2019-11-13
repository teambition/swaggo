package parserv3

import (
	"errors"
	"fmt"

	"github.com/teambition/swaggo/swaggerv3"
)

type result struct {
	kind       kind
	buildin    string      // golang type
	def        interface{} // default value
	desc       string
	ref        string
	item       *result
	required   []string
	deprecated []string
	items      map[string]*result

	sType   string // swagger type
	sFormat string // swagger format
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
		ss.Deprecated = r.deprecated
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
	switch r.kind {
	case innerKind:
		sp.Default = r.def
		sp.Type = r.sType
		sp.Description = r.desc
	case arrayKind:
		sp.Type = "array"
		sp.Items = &swaggerv3.Propertie{}
		sp.Description = r.desc
		r.item.parsePropertie(sp.Items)
	case mapKind:
		sp.Type = "object"
		sp.AdditionalProperties = &swaggerv3.Propertie{}
		r.item.parsePropertie(sp.AdditionalProperties)
	case objectKind:
		if r.ref != "" {
			sp.Ref = r.ref
			return
		}
		sp.Description = r.desc
		sp.Required = r.required
		sp.Deprecated = r.deprecated
		if sp.Properties == nil {
			sp.Properties = make(map[string]*swaggerv3.Propertie)
		}
		for k, v := range r.items {
			tmpSp := &swaggerv3.Propertie{}
			v.parsePropertie(tmpSp)
			sp.Properties[k] = tmpSp
		}
	case interfaceKind:
		sp.Type = "object"
		// TODO
	}
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
