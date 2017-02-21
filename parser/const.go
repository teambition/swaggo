package parser

const (
	jsonType  = "json"
	appJson   = "application/json"
	xmlType   = "xml"
	appXml    = "application/xml"
	plainType = "plain"
	textPlain = "text/plain"
	htmlType  = "html"
	textHtml  = "text/html"
)

const (
	jsonFile = "swagger.json"
	yamlFile = "swagger.yaml"
)

const (
	docPrefix = "@"
	// api tag
	apiVersion           = "@Version"
	apiTitle             = "@Title"
	apiDesc              = "@Description"
	apiTermsOfServiceUrl = "@TermsOfServiceUrl"
	apiContact           = "@Contact"
	apiName              = "@Name"
	apiURL               = "@URL"
	apiLicenseUrl        = "@LicenseUrl"
	apiLicense           = "@License"
	apiSchemes           = "@Schemes"
	apiHost              = "@Host"
	apiBasePath          = "@BasePath"
	apiConsumes          = "@Consumes"
	apiProduces          = "@Produces"
	// controller tag
	ctrlPrivate = "@Private"
	ctrlName    = "@Name"
	ctrlDesc    = "@Description"
	// method tag
	methodPrivate    = "@Private" // @Private
	methodTitle      = "@Title"
	methodDesc       = "@Description"
	methodSummary    = "@Summary"
	methodSuccess    = "@Success"
	methodParam      = "@Param"
	methodFailure    = "@Failure"
	methodDeprecated = "@Deprecated"
	methodAccept     = "@Accept"
	methodProduce    = "@Produce"
	methodRouter     = "@Router"
)

const (
	query    = "query"
	header   = "header"
	path     = "path"
	formData = "formData"
	body     = "body"
)

// inner type and swagger type
var basicTypes = map[string]string{
	"bool":       "boolean:",
	"uint":       "integer:int32",
	"uint8":      "integer:int32",
	"uint16":     "integer:int32",
	"uint32":     "integer:int32",
	"uint64":     "integer:int64",
	"int":        "integer:int32",
	"int8":       "integer:int32",
	"int16":      "integer:int32",
	"int32":      "integer:int32",
	"int64":      "integer:int64",
	"uintptr":    "integer:int64",
	"float32":    "number:float",
	"float64":    "number:double",
	"string":     "string:",
	"complex64":  "number:float",
	"complex128": "number:double",
	"byte":       "string:byte",
	"rune":       "string:byte",
}
