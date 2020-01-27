//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen -package=api -o=pkg/api/openapi.go -generate=types,server,spec openapi.yaml

package main

import (
	"fmt"
	"os"

	_ "github.com/cosmtrek/air/runner" // used for hot reload
	"github.com/projectsyn/lieutenant-api/pkg/service"
)

// Version is the lieutenant-api version (set during build)
var Version = "unreleased"

func main() {
	e, err := service.NewAPIServer()
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
	fmt.Println("Start server")
	e.Logger.Fatal(e.Start(":8080"))
}
