package parser

const (
	jsonType     = "json"
	xmlType      = "xml"
	plainType    = "plain"
	htmlType     = "html"
	formType     = "form"
	formDataType = "formData"
	streamType   = "stream"
)

var contentType = map[string]string{
	jsonType:     "application/json",
	xmlType:      "application/xml",
	plainType:    "text/plain",
	htmlType:     "text/html",
	formType:     "application/x-www-form-urlencoded",
	formDataType: "multipart/form-data",
	streamType:   "application/octet-stream",
}

const (
	jsonFile = "swagger.json"
	yamlFile = "swagger.yaml"
)

const (
	docPrefix = "@"
	// app tag
	appVersion           = "@Version"
	appTitle             = "@Title"
	appDesc              = "@Description"
	appTermsOfServiceUrl = "@TermsOfServiceUrl"
	appContact           = "@Contact"
	appName              = "@Name"
	appURL               = "@URL"
	appLicenseUrl        = "@LicenseUrl"
	appLicense           = "@License"
	appSchemes           = "@Schemes"
	appHost              = "@Host"
	appBasePath          = "@BasePath"
	appConsumes          = "@Consumes"
	appProduces          = "@Produces"
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
	methodConsumes   = "@Consumes"
	methodProduces   = "@Produces"
	methodRouter     = "@Router"
)

const (
	query  = "query"
	header = "header"
	path   = "path"
	form   = "form"
	body   = "body"
)

var paramType = map[string]string{
	query:  "query",
	header: "header",
	path:   "path",
	form:   "formData",
	body:   "body",
}
