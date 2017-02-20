package main

import (
	"os"

	"github.com/teambition/swaggo/generator"
)

func main() {
	generator.NewCli().Run(os.Args)
}
