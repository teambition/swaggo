package main

import (
	"log"
	"os"

	"github.com/teambition/swaggo/parserv3"
	//"github.com/teambition/swaggo/parser"
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
			Value: "yaml",
			Usage: "the type of swagger file (json or yaml)",
		},
	}
	app.Action = func(c *cli.Context) error {
		if err := parserv3.Parse(c.String("project"),
			c.String("swagger"),
			c.String("output"),
			c.String("type")); err != nil {
			return err
		}
		return nil
	}
	if err := app.Run(os.Args); err != nil {
		log.Printf("[Error] %v", err)
	}
}
