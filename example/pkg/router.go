package pkg

import (
	"encoding/json"

	"github.com/gocraft/web"
	"github.com/teambition/swaggo/example/pkg/api"
)

func New() *web.Router {
	router := web.New(api.Context{}).
		Middleware(web.LoggerMiddleware).
		Middleware(web.ShowErrorsMiddleware).
		Middleware(func(c *api.Context, rw web.ResponseWriter, req *web.Request, next web.NextMiddlewareFunc) {
			resultJSON, _ := json.Marshal(c.Response)
			rw.Write(resultJSON)
		}).
		Get("/testapi/get-string-by-int/{some_id}", (*api.Context).GetStringByInt).
		Get("/testapi/get-struct-by-int/{some_id}", (*api.Context).GetStructByInt).
		Get("/testapi/get-simple-array-by-string/{some_id}", (*api.Context).GetSimpleArrayByString).
		Get("/testapi/get-struct-array-by-string/{some_id}", (*api.Context).GetStructArrayByString).
		Post("/testapi/get-struct3", (*api.Context).PostStruct3).
		Delete("/testapi/get-struct3", (*api.Context).DelStruct3).
		Get("/testapi/get-struct2-by-int/{some_id}", (*api.Context).GetStruct2ByInt)
	return router
}
