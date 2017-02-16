package main

import (
	"flag"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"encoding/json"

	"github.com/teambition/swaggo/generate"
	"github.com/teambition/swaggo/swagger"
	"github.com/teambition/swaggo/utils"
)

var swaggerFile = flag.String("swagger", "./swagger.go", "swagger.go path")

const (
	swaggerVersion = "2.0"
)

func main() {
	flag.Parse()

	f, err := parser.ParseFile(token.NewFileSet(), *swaggerFile, nil, parser.ParseComments)
	if err != nil {
		fmt.Println(err)
		return
	}

	sw := &swagger.Swagger{SwaggerVersion: swaggerVersion}
	// Analyse API comments
	if f.Comments != nil {
		for _, c := range f.Comments {
			for _, s := range strings.Split(c.Text(), "\n") {
				switch {
				case strings.HasPrefix(s, "@Version"):
					sw.Infos.Version = strings.TrimSpace(s[len("@APIVersion"):])
				case strings.HasPrefix(s, "@Title"):
					sw.Infos.Title = strings.TrimSpace(s[len("@Title"):])
				case strings.HasPrefix(s, "@Description"):
					sw.Infos.Description = strings.TrimSpace(s[len("@Description"):])
				case strings.HasPrefix(s, "@TermsOfServiceUrl"):
					sw.Infos.TermsOfService = strings.TrimSpace(s[len("@TermsOfServiceUrl"):])
				case strings.HasPrefix(s, "@Contact"):
					sw.Infos.Contact.EMail = strings.TrimSpace(s[len("@Contact"):])
				case strings.HasPrefix(s, "@Name"):
					sw.Infos.Contact.Name = strings.TrimSpace(s[len("@Name"):])
				case strings.HasPrefix(s, "@URL"):
					sw.Infos.Contact.URL = strings.TrimSpace(s[len("@URL"):])
				case strings.HasPrefix(s, "@LicenseUrl"):
					if sw.Infos.License == nil {
						sw.Infos.License = &swagger.License{URL: strings.TrimSpace(s[len("@LicenseUrl"):])}
					} else {
						sw.Infos.License.URL = strings.TrimSpace(s[len("@LicenseUrl"):])
					}
				case strings.HasPrefix(s, "@License"):
					if sw.Infos.License == nil {
						sw.Infos.License = &swagger.License{Name: strings.TrimSpace(s[len("@License"):])}
					} else {
						sw.Infos.License.Name = strings.TrimSpace(s[len("@License"):])
					}
				case strings.HasPrefix(s, "@Schemes"):
					sw.Schemes = strings.Split(strings.TrimSpace(s[len("@Schemes"):]), ",")
				case strings.HasPrefix(s, "@Host"):
					sw.Host = strings.TrimSpace(s[len("@Host"):])
				case strings.HasPrefix(s, "@BasePath"):
					sw.BasePath = strings.TrimSpace(s[len("@BasePath"):])
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
		p, err := generate.NewCtrlPackage(importPath, filter)
		if err != nil {
			fmt.Println(err)
			return
		}
		if err = p.Run(sw); err != nil {
			fmt.Println(err)
			return
		}
	}
	data, _ := json.Marshal(sw)
	fmt.Println(string(data))
}

func filter(importPath string) bool {
	goRoot := os.Getenv("GOROOT")
	if goRoot == "" {
		panic("GOROOT environment variable is not set or empty")
	}
	wg, _ := filepath.EvalSymlinks(filepath.Join(goRoot, "src", importPath))
	return !utils.FileExists(wg)
}
