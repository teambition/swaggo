package parser

const (
	json   = "json"
	ajson  = "application/json"
	xml    = "xml"
	axml   = "application/xml"
	plain  = "plain"
	tplain = "text/plain"
	html   = "html"
	thtml  = "text/html"
)

const (
	docPrefix = "@"
	// controller tag
	ctrlName = "@Name"
	ctrlDesc = "@Description"
	// controller item
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
