package main

import (
	"fmt"
	"os"

	"github.com/teambition/swaggo/parser"
	"github.com/urfave/cli"
)

const (
	AppVersion = "v0.2.4"
)

func main() {
	app := cli.NewApp()
	app.Version = AppVersion
	app.Name = "swaggo"
	app.HelpName = "swaggo"
	app.Usage = "a utility for convert go annotations to swagger-doc"
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "dev, d",
			Usage: "develop mode",
		},
		cli.StringFlag{
			Name:  "swagger, s",
			Value: "./swagger.go",
			Usage: "where is the swagger.go file",
		},
		cli.StringFlag{
			Name:  "project, p",
			Value: "./",
			Usage: "where is the project",
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
		if err := parser.Parse(c.String("project"),
			c.String("swagger"),
			c.String("output"),
			c.String("type"),
			c.Bool("dev")); err != nil {
			return fmt.Errorf("[Error] %v", err)
		}
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		fmt.Printf("[Error] %v", err)
	}
}
