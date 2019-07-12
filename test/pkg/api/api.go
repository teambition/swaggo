package api

import (
	"fmt"
	_ "os"

	"github.com/gocraft/web"
	"github.com/teambition/swaggo/test/pkg/api/subpackage"
	sub "github.com/teambition/swaggo/test/pkg/api/subpackage_alias"
	. "github.com/teambition/swaggo/test/pkg/api/subpackage_dot"
)

var (
	_ = sub.SubStructAlias{}
	_ = SubStructDot{}
)

// @Name testapi
// @Description test apis
type Context struct {
	Response interface{}
}

func (c *Context) WriteResponse(response interface{}) {
	c.Response = response
}

// Title unique id
// @Title GetStringByInt
//
// Deprecated show if this method has been deprecated
// @Deprecated true
//
// Summary short explain it's action
// @Summary get string by ID summary
// @Summary multi line
//
// Description long explain about implement
// @Description get string by ID desc
// @Description multi line
//
// @Param Authorization header string true "oauth token"
// Consumes type include(json,plain,xml)
// @Consumes json,plain,xml,html
//
// Produces type include(json,plain,xml,html)
// @Produces json,plain,xml,html
//
// Param:param_name/param_type/data_type/required(optional)/describtion(optional)/defaul_value(optional)
// value == "-" means optional
// form and body params cann't coexist
// path param must be required
// if file type param exsited, all params must be form except path and query
// @Param path_param path int - "Some ID" 123
// @Param form_param form file - "Request Form"
// @Param query_param query []string - "Array"
// @Param query_param_2 query [][]string - "Array Array"
//
// Success:response_code/data_type(optional)/describtion(optional)
// @Success 200 string "Success"
// @Success 201 SubStructDot "Success"
// @Success 202 sub.SubStructAlias "Success"
// @Success 203 StructureWithAnonymousStructure
// @Success 204 map[string]string
//
// Failure:response_code/data_type(optional)/describtion(optional)
// @Failure 400 APIError "We need ID!!"
// @Failure 404 APIError "Can not find ID"
//
// Router:http_method/api_path
// @Router GET /testapi/get-string-by-int/some_id
func (c *Context) GetStringByInt(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse(fmt.Sprint("Some data for %s ID", req.PathParams["some_id"]))
}

// @Title GetStructByInt
// @Summary get struct by ID
// @Description get struct by ID
// @Permission member get
// @Consumes json
// @Produces json
// @Param some_id path int true "Some ID"
// @Param offset query int true "Offset"
// @Param limit query int true "Limit"
// @Success 200  StructureWithEmbededStructure "Success"
// @Failure 400  APIError "We need ID!!"
// @Failure 404  APIError "Can not find ID"
// @Router GET /testapi/get-struct-by-int/some_id
func (c *Context) GetStructByInt(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse(StructureWithEmbededStructure{})
}

// @Title GetStruct2ByInt
// @Summary get struct2 by ID
// @Description get struct2 by ID
// @Consumes json
// @Produces json
// @Param some_id path int true "Some ID"
// @Param offset query int true "Offset"
// @Param limit query int true "Limit"
// @Success 200 StructureWithEmbededPointer "Success"
// @Failure 400 APIError "We need ID!!"
// @Failure 404 APIError "Can not find ID"
// @Router GET /testapi/get-struct2-by-int/some_id
func (c *Context) GetStruct2ByInt(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse(StructureWithEmbededPointer{})
}

// @Title GetSimpleArrayByString
// @Summary get simple array by ID
// @Description get simple array by ID
// @Consumes json
// @Produces json
// @Param some_id path string true "Some ID"
// @Param offset query int true "Offset"
// @Param limit query int true "Limit"
// @Success 200 []string "Success"
// @Failure 400 APIError "We need ID!!"
// @Failure 404 APIError "Can not find ID"
// @Router POST /testapi/get-simple-array-by-string/some_id
func (c *Context) GetSimpleArrayByString(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse([]string{"one", "two", "three"})
}

// @Title GetStructArrayByString
// @Summary get struct array by ID
// @Description get struct array by ID
// @Consumes json
// @Produces json
// @Param some_id path string true "Some ID" "hello world"
// @Param body body subpackage.SimpleStructure true
// @Param limit query int true "Limit"
// @Success 200 []subpackage.SimpleStructure "Success"
// @Failure 400 APIError "We need ID!!"
// @Failure 404 APIError "Can not find ID"
// @Router PUT /testapi/get-struct-array-by-string/some_id
func (c *Context) GetStructArrayByString(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse([]subpackage.SimpleStructure{
		subpackage.SimpleStructure{},
		subpackage.SimpleStructure{},
		subpackage.SimpleStructure{},
	})
}

// @Title SameStruct
// @Summary get struct array by ID
// @Description get struct array by ID
// @Consumes json
// @Produces json
// @Param some_id path string true "Some ID"
// @Param offset query int true "Offset"
// @Param body body []SimpleStructure true "Body"
// @Param limit query int true "Limit"
// @Success 200 []SimpleStructure "Success"
// @Failure 400 APIError "We need ID!!"
// @Failure 404 APIError "Can not find ID"
// @Router PUT /testapi/get-same-struct-array-by-string/some_id
func (c *Context) GetSameStructArraryByString(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse([]SimpleStructure{
		SimpleStructure{},
		SimpleStructure{},
		SimpleStructure{},
	})
}

// @Title GetStruct3
// @Summary get struct3 summary
// @Description get struct3 desc
// @Consumes json
// @Produces json
// @Success 200 SimpleStructure "Success"
// @Success 201 SimpleStructure "Success"
// @Failure 400 APIError "We need ID!!"
// @Failure 404 APIError "Can not find ID"
// @Router DELETE /testapi/get-struct3
func (c *Context) DelStruct3(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse(StructureWithSlice{})
}

// @Title GetStruct3
// @Summary get struct3
// @Description get struct3
// @Consumes json
// @Produces json
// @Success 204 - "null"
// @Success 200 StructureWithSlice "Success"
// @Success 201 TypeInterface "Success"
// @Failure 400 APIError "We need ID!!"
// @Failure 404 APIError "Can not find ID"
// @Router POST /testapi/get-struct3
func (c *Context) PostStruct3(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse(StructureWithSlice{})
}
