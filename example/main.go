package main

import (
	"net/http"

	"github.com/teambition/swaggo/example/pkg"
)

func main() {
	router := pkg.New()
	http.ListenAndServe("localhost:3000", router)
}
