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
	query  = "query"
	header = "header"
	path   = "path"
	form   = "form"
	body   = "body"
)
