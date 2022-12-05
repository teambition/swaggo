package main

import (
	"log"
	"os"

	"github.com/teambition/swaggo/parser"
	"github.com/urfave/cli"
)

const (
	version = "v0.2.8"
)

func main() {
	app := cli.NewApp()
	app.Version = version
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
		cli.BoolFlag{
			Name:  "mod, m",
			Usage: "whether use the mod ",
		},
	}
	app.Action = func(c *cli.Context) error {
		if err := parser.Parse(c.String("project"),
			c.String("swagger"),
			c.String("output"),
			c.String("type"),
			c.Bool("dev"),
			c.Bool("mod"),
		); err != nil {
			return err
		}
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		log.Printf("[Error] %v", err)
	}
}
