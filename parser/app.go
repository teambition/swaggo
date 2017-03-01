package parser

import (
	"encoding/json"
	"fmt"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/teambition/swaggo/swagger"
	yaml "gopkg.in/yaml.v2"
)

func doc2Swagger(swaggerGo string, sw *swagger.Swagger) error {
	f, err := parser.ParseFile(token.NewFileSet(), swaggerGo, nil, parser.ParseComments)
	if err != nil {
		return err
	}
	// Analyse API comments
	if f.Comments != nil {
		for _, c := range f.Comments {
			for _, s := range strings.Split(c.Text(), "\n") {
				switch {
				case tagTrimPrefixAndSpace(&s, apiVersion):
					sw.Infos.Version = s
				case tagTrimPrefixAndSpace(&s, apiTitle):
					sw.Infos.Title = s
				case tagTrimPrefixAndSpace(&s, apiDesc):
					sw.Infos.Description = s
				case tagTrimPrefixAndSpace(&s, apiTermsOfServiceUrl):
					sw.Infos.TermsOfService = s
				case tagTrimPrefixAndSpace(&s, apiContact):
					sw.Infos.Contact.EMail = s
				case tagTrimPrefixAndSpace(&s, apiName):
					sw.Infos.Contact.Name = s
				case tagTrimPrefixAndSpace(&s, apiURL):
					sw.Infos.Contact.URL = s
				case tagTrimPrefixAndSpace(&s, apiLicenseUrl):
					sw.Infos.License.URL = s
				case tagTrimPrefixAndSpace(&s, apiLicense):
					sw.Infos.License.Name = s
				case tagTrimPrefixAndSpace(&s, apiSchemes):
					sw.Schemes = strings.Split(s, ",")
				case tagTrimPrefixAndSpace(&s, apiHost):
					sw.Host = s
				case tagTrimPrefixAndSpace(&s, apiBasePath):
					sw.BasePath = s
				case tagTrimPrefixAndSpace(&s, apiConsumes):
					sw.Consumes = contentTypeByDoc(s)
				case tagTrimPrefixAndSpace(&s, apiProduces):
					sw.Produces = contentTypeByDoc(s)
				}
			}
		}
	}

	// Analyse controller package
	// like:
	// swagger.go
	// import (
	//     _ "path/to/ctrl1"
	//     _ "path/to/ctrl2"
	//     _ "path/to/ctrl3"
	// )
	// // @APIVersion xxx
	// // @....
	for _, im := range f.Imports {
		importPath := strings.Trim(im.Path.Value, "\"")
		p, err := newResoucre(importPath, true)
		if err != nil {
			return err
		}
		if err = p.run(sw); err != nil {
			return err
		}
	}
	return nil
}

func Parser(swaggerGo, output, t string) error {
	sw := swagger.NewV2()
	err := doc2Swagger(swaggerGo, sw)
	if err != nil {
		return err
	}
	var (
		data     []byte
		filename string
	)

	switch t {
	case "json":
		filename = jsonFile
		data, err = json.Marshal(sw)
	case "yaml":
		filename = yamlFile
		data, err = yaml.Marshal(sw)
	default:
		err = fmt.Errorf("missing swagger file type(%s), only support in (json, yaml)", t)
	}
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(output, filename), data, 0644)
}
