package generator

import (
	"encoding/json"
	"go/parser"
	"go/token"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/astaxie/beego/utils"
	swaggoParser "github.com/teambition/swaggo/parser"
	"github.com/teambition/swaggo/swagger"
	"github.com/urfave/cli"
	"gopkg.in/yaml.v2"
)

func NewCli() *cli.App {
	app := cli.NewApp()
	app.Version = AppVersion
	app.Name = "swaggo"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "swagger, s",
			Value: "./swagger.go",
			Usage: "where is the swagger.go file",
		},
		cli.StringFlag{
			Name:  "output, o",
			Value: "./",
			Usage: "the output of the swagger file that was generated",
		},
		cli.StringFlag{
			Name:  "type, t",
			Value: "json",
			Usage: "the type of swagger file (json or yaml)",
		},
	}
	app.Action = func(c *cli.Context) error {
		// TODO: move this implement to parser
		swaggerGo := c.String("swagger")
		output := c.String("output")
		t := c.String("type")

		f, err := parser.ParseFile(token.NewFileSet(), swaggerGo, nil, parser.ParseComments)
		if err != nil {
			return err
		}

		sw := &swagger.Swagger{SwaggerVersion: SwaggerVersion}
		// Analyse API comments
		if f.Comments != nil {
			for _, c := range f.Comments {
				for _, s := range strings.Split(c.Text(), "\n") {
					switch {
					case strings.HasPrefix(s, apiVersion):
						sw.Infos.Version = strings.TrimSpace(s[len(apiVersion):])
					case strings.HasPrefix(s, apiTitle):
						sw.Infos.Title = strings.TrimSpace(s[len(apiTitle):])
					case strings.HasPrefix(s, apiDesc):
						sw.Infos.Description = strings.TrimSpace(s[len(apiDesc):])
					case strings.HasPrefix(s, apiTermsOfServiceUrl):
						sw.Infos.TermsOfService = strings.TrimSpace(s[len(apiTermsOfServiceUrl):])
					case strings.HasPrefix(s, apiContact):
						sw.Infos.Contact.EMail = strings.TrimSpace(s[len(apiContact):])
					case strings.HasPrefix(s, apiName):
						sw.Infos.Contact.Name = strings.TrimSpace(s[len(apiName):])
					case strings.HasPrefix(s, apiURL):
						sw.Infos.Contact.URL = strings.TrimSpace(s[len(apiURL):])
					case strings.HasPrefix(s, apiLicenseUrl):
						if sw.Infos.License == nil {
							sw.Infos.License = &swagger.License{URL: strings.TrimSpace(s[len(apiLicenseUrl):])}
						} else {
							sw.Infos.License.URL = strings.TrimSpace(s[len(apiLicenseUrl):])
						}
					case strings.HasPrefix(s, apiLicense):
						if sw.Infos.License == nil {
							sw.Infos.License = &swagger.License{Name: strings.TrimSpace(s[len(apiLicense):])}
						} else {
							sw.Infos.License.Name = strings.TrimSpace(s[len(apiLicense):])
						}
					case strings.HasPrefix(s, apiSchemes):
						sw.Schemes = strings.Split(strings.TrimSpace(s[len(apiSchemes):]), ",")
					case strings.HasPrefix(s, apiHost):
						sw.Host = strings.TrimSpace(s[len(apiHost):])
					case strings.HasPrefix(s, apiBasePath):
						sw.BasePath = strings.TrimSpace(s[len(apiBasePath):])
					case strings.HasPrefix(s, apiConsumes):
						sw.Consumes = strings.Split(strings.TrimSpace(s[len(apiConsumes):]), ",")
					case strings.HasPrefix(s, apiProduces):
						sw.Produces = strings.Split(strings.TrimSpace(s[len(apiProduces):]), ",")
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
			p, err := swaggoParser.NewResoucre(importPath, systemPackageFilter)
			if err != nil {
				return err
			}
			if err = p.Run(sw); err != nil {
				return err
			}
		}

		var (
			data     []byte
			filename string
		)

		switch t {
		case "json":
			filename = SwaggerJsonFile
			data, err = json.Marshal(sw)
		case "yaml":
			filename = SwaggerYamlFile
			data, err = yaml.Marshal(sw)
		}
		if err != nil {
			return err
		}
		return ioutil.WriteFile(filepath.Join(output, filename), data, 0644)
	}

	return app
}

func systemPackageFilter(importPath string) bool {
	goRoot := os.Getenv("GOROOT")
	if goRoot == "" {
		panic("GOROOT environment variable is not set or empty")
	}
	wg, _ := filepath.EvalSymlinks(filepath.Join(goRoot, "src", importPath))
	return !utils.FileExists(wg)
}
