package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"github.com/teambition/swaggo/swagger"
)

func TestErrorPath(t *testing.T) {
	assert := assert.New(t)
	// error test
	swaggerGo := "../example/swagger.go.err"
	as, err := NewAppSuite(swaggerGo)
	assert.Nil(as)
	assert.NotNil(err)
}

func TestAppSuite(t *testing.T) {
	assert := assert.New(t)
	// error test
	swaggerGo := "../example/swagger.go"
	as, err := NewAppSuite(swaggerGo)
	assert.Nil(err)
	assert.NotNil(as)
	suite.Run(t, as)
}

type AppSuite struct {
	suite.Suite
	*swagger.Swagger
}

func NewAppSuite(swaggerGo string) (*AppSuite, error) {
	as := &AppSuite{Swagger: swagger.NewV2()}
	if err := doc2Swagger(swaggerGo, as.Swagger); err != nil {
		return nil, err
	}
	return as, nil
}

func (suite *AppSuite) TestSwagger() {
	assert := assert.New(suite.T())
	assert.Equal("2.0", suite.SwaggerVersion)
	assert.Equal("Swagger Example API", suite.Infos.Title)
	assert.Equal("Swagger Example API", suite.Infos.Description)
	assert.Equal("1.0.0", suite.Infos.Version)
	assert.Equal("http://teambition.com/", suite.Infos.TermsOfService)
	// contact
	assert.Equal("swagger", suite.Infos.Contact.Name)
	assert.Equal("swagger@teambition.com", suite.Infos.Contact.EMail)
	assert.Equal("teambition.com", suite.Infos.Contact.URL)
	// license
	assert.Equal("Apache", suite.Infos.License.Name)
	assert.Equal("http://teambition.com/", suite.Infos.License.URL)
	// schemes
	assert.Equal([]string{"http", "wss"}, suite.Schemes)
	// consumes and produces
	assert.Equal([]string{"application/json", "text/plain", "application/xml", "text/html"}, suite.Consumes)
	assert.Equal([]string{"application/json", "text/plain", "application/xml", "text/html"}, suite.Produces)

	assert.Equal("127.0.0.1:3000", suite.Host)
	assert.Equal("/api", suite.BasePath)
	assert.Equal(7, len(suite.Paths))
	router := suite.Paths["/testapi/get-string-by-int/{some_id}"]
	assert.NotNil(router)
	assert.NotNil(router.Get)
	assert.Equal([]string{"testapi"}, router.Get.Tags)
	assert.Equal("get string by ID summary\nmulti line", router.Get.Summary)
	assert.Equal("get string by ID desc\nmulti line", router.Get.Description)
	assert.Equal("testapi.GetStringByInt", router.Get.OperationID)
	assert.Equal([]string{"application/json", "text/plain", "application/xml", "text/html"}, router.Get.Consumes)
	assert.Equal([]string{"application/json", "text/plain", "application/xml", "text/html"}, router.Get.Produces)

	assert.Equal("path", router.Get.Parameters[0].In)
	assert.Equal("some_id", router.Get.Parameters[0].Name)
	assert.Equal("Some ID", router.Get.Parameters[0].Description)
	assert.Equal(true, router.Get.Parameters[0].Required)
	assert.Equal("integer", router.Get.Parameters[0].Type)
	assert.Equal("int32", router.Get.Parameters[0].Format)
	assert.Equal(123, router.Get.Parameters[0].Default)

	// 200
	assert.NotNil(router.Get.Responses["200"])
	assert.Equal("string", router.Get.Responses["200"].Schema.Type)

	// 400
	assert.NotNil(router.Get.Responses["400"])
	assert.Equal("We need ID!!", router.Get.Responses["400"].Description)
	assert.Equal("#/definitions/APIError", router.Get.Responses["400"].Schema.Ref)
	assert.Equal("object", router.Get.Responses["400"].Schema.Type)

	// 404
	assert.NotNil(router.Get.Responses["404"])
	assert.Equal("Can not find ID", router.Get.Responses["404"].Description)
	assert.Equal("#/definitions/APIError", router.Get.Responses["404"].Schema.Ref)
	assert.Equal("object", router.Get.Responses["404"].Schema.Type)

	// definitions
	// APIError
	apiError := suite.Definitions["APIError"]
	assert.NotNil(apiError)
	assert.Equal("APIError", apiError.Title)
	assert.Equal("object", apiError.Type)
	assert.Equal("integer", apiError.Properties["ErrorCode"].Type)
	assert.Equal("int32", apiError.Properties["ErrorCode"].Format)
	assert.Equal("string", apiError.Properties["ErrorMessage"].Type)

	// inherit
	simpleStruct := suite.Definitions["SimpleStructure_1"]
	assert.NotNil(simpleStruct)
	assert.Equal("SimpleStructure", simpleStruct.Title)
	assert.Equal("object", simpleStruct.Type)
	assert.Equal([]string{"id", "name", "age", "ctime", "sub", "i"}, simpleStruct.Required)
	assert.Equal("the user age", simpleStruct.Properties["age"].Description)
	assert.Equal("18", simpleStruct.Properties["age"].Default)
	assert.Equal("integer", simpleStruct.Properties["age"].Type)
	assert.Equal("int32", simpleStruct.Properties["age"].Format)

	assert.Equal("#/definitions/SimpleStructure", simpleStruct.Properties["sub"].Ref)
	assert.Equal("object", simpleStruct.Properties["sub"].Type)

	// tags
	assert.Equal(1, len(suite.Tags))
	assert.Equal("testapi", suite.Tags[0].Name)
	assert.Equal("test apis", suite.Tags[0].Description)
}
