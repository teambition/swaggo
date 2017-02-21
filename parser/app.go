package parser

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/teambition/swaggo/swagger"
	yaml "gopkg.in/yaml.v2"
)

func Parser(swaggerGo, output, t string) error {
	f, err := parser.ParseFile(token.NewFileSet(), swaggerGo, nil, parser.ParseComments)
	if err != nil {
		return err
	}

	sw := swagger.NewV2()
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
					if sw.Infos.License == nil {
						sw.Infos.License = &swagger.License{URL: s}
					} else {
						sw.Infos.License.URL = s
					}
				case tagTrimPrefixAndSpace(&s, apiLicense):
					if sw.Infos.License == nil {
						sw.Infos.License = &swagger.License{Name: s}
					} else {
						sw.Infos.License.Name = s
					}
				case tagTrimPrefixAndSpace(&s, apiSchemes):
					sw.Schemes = strings.Split(s, ",")
				case tagTrimPrefixAndSpace(&s, apiHost):
					sw.Host = s
				case tagTrimPrefixAndSpace(&s, apiBasePath):
					sw.BasePath = s
				case tagTrimPrefixAndSpace(&s, apiConsumes):
					for _, a := range strings.Split(s, ",") {
						switch a {
						case jsonType:
							sw.Consumes = append(sw.Consumes, appJson)
						case xmlType:
							sw.Consumes = append(sw.Consumes, appXml)
						case plainType:
							sw.Consumes = append(sw.Consumes, textPlain)
						case htmlType:
							sw.Consumes = append(sw.Consumes, textHtml)
						}
					}
				case tagTrimPrefixAndSpace(&s, apiProduces):
					for _, p := range strings.Split(s, ",") {
						switch p {
						case jsonType:
							sw.Produces = append(sw.Produces, appJson)
						case xmlType:
							sw.Produces = append(sw.Produces, appXml)
						case plainType:
							sw.Produces = append(sw.Produces, textPlain)
						case htmlType:
							sw.Produces = append(sw.Produces, textHtml)
						}
					}
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
		if systemPackageFilter(importPath) {
			p, err := newResoucre(importPath)
			if err != nil {
				return err
			}
			if err = p.run(sw); err != nil {
				return err
			}
		}
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
	}
	if err != nil {
		return err
	}
	return ioutil.WriteFile(filepath.Join(output, filename), data, 0644)
}
