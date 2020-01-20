//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen -package=api -o=pkg/api/openapi.go -generate=types,server,spec openapi.yaml

package main

import (
	"fmt"
	"os"

	"github.com/labstack/echo/v4"
	echomiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/projectsyn/lieutenant/pkg/api"
	"github.com/projectsyn/lieutenant/pkg/service"
)

// Version is the lieutenant-api version (set during build)
var Version = "unreleased"

func main() {

	swagger, err := api.GetSwagger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading swagger spec\n: %s", err)
		os.Exit(1)
	}

	swagger.Servers = nil

	svc := service.NewService()

	e := echo.New()
	e.Use(echomiddleware.Logger())
	e.Pre(echomiddleware.RemoveTrailingSlash())

	apiGroup := e.Group("api")
	//apiGroup.Use(middleware.OapiRequestValidator(swagger))
	api.RegisterHandlers(apiGroup, svc)

	fmt.Println("Start server")
	e.Logger.Fatal(e.Start(":8080"))
}
