package api

import (
	"fmt"

	"github.com/gocraft/web"
	"github.com/teambition/swaggo/example/pkg/api/subpackage"
)

// @Name testapi
// @Description test apis
type Context struct {
	Response interface{}
}

func (c *Context) WriteResponse(response interface{}) {
	c.Response = response
}

// @Title GetStringByInt
// @Summary get string by ID
// @Description get string by ID
// @Accept json,plain
// @Produce json,plain
// @Param some_id path int true "Some ID"
// @Success 200  string
// @Failure 400  APIError "We need ID!!"
// @Failure 404  APIError "Can not find ID"
// @Router GET /testapi/get-string-by-int/{some_id}
func (c *Context) GetStringByInt(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse(fmt.Sprint("Some data for %s ID", req.PathParams["some_id"]))
}

// @Title GetStructByInt
// @Summary get struct by ID
// @Description get struct by ID
// @Accept json
// @Produce json
// @Param some_id path int true "Some ID"
// @Param offset query int true "Offset"
// @Param limit query int true "Offset"
// @Success 200  StructureWithEmbededStructure
// @Failure 400  APIError "We need ID!!"
// @Failure 404  APIError "Can not find ID"
// @Router GET /testapi/get-struct-by-int/{some_id}
func (c *Context) GetStructByInt(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse(StructureWithEmbededStructure{})
}

// @Title GetStruct2ByInt
// @Summary get struct2 by ID
// @Description get struct2 by ID
// @Accept json
// @Produce json
// @Param some_id path int true "Some ID"
// @Param offset query int true "Offset"
// @Param limit query int true "Offset"
// @Success 200 StructureWithEmbededPointer
// @Failure 400 APIError "We need ID!!"
// @Failure 404 APIError "Can not find ID"
// @Router GET /testapi/get-struct2-by-int/{some_id}
func (c *Context) GetStruct2ByInt(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse(StructureWithEmbededPointer{})
}

// @Title GetSimpleArrayByString
// @Summary get simple array by ID
// @Description get simple array by ID
// @Accept json
// @Produce json
// @Param some_id path string true "Some ID"
// @Param offset query int true "Offset"
// @Param limit query int true "Offset"
// @Success 200 []string
// @Failure 400 APIError "We need ID!!"
// @Failure 404 APIError "Can not find ID"
// @Router POST /testapi/get-simple-array-by-string/{some_id}
func (c *Context) GetSimpleArrayByString(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse([]string{"one", "two", "three"})
}

// @Title GetStructArrayByString
// @Summary get struct array by ID
// @Description get struct array by ID
// @Accept json
// @Produce json
// @Param some_id path string true "Some ID"
// @Param offset query int true "Offset"
// @Param limit query int true "Offset"
// @Success 200 []subpackage.SimpleStructure
// @Failure 400 APIError "We need ID!!"
// @Failure 404 APIError "Can not find ID"
// @Router PUT /testapi/get-struct-array-by-string/{some_id}
func (c *Context) GetStructArrayByString(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse([]subpackage.SimpleStructure{
		subpackage.SimpleStructure{},
		subpackage.SimpleStructure{},
		subpackage.SimpleStructure{},
	})
}

// @Title GetStruct3
// @Summary get struct3
// @Description get struct3
// @Accept json
// @Produce json
// @Success 200 StructureWithSlice
// @Failure 400 APIError "We need ID!!"
// @Failure 404 APIError "Can not find ID"
// @Router DELETE /testapi/get-struct3
func (c *Context) DelStruct3(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse(StructureWithSlice{})
}

// @Title GetStruct3
// @Summary get struct3
// @Description get struct3
// @Accept json
// @Produce json
// @Success 200 StructureWithSlice
// @Failure 400 APIError "We need ID!!"
// @Failure 404 APIError "Can not find ID"
// @Router POST /testapi/get-struct3
func (c *Context) PostStruct3(rw web.ResponseWriter, req *web.Request) {
	c.WriteResponse(StructureWithSlice{})
}
