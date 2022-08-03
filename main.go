//go:generate go run github.com/deepmap/oapi-codegen/cmd/oapi-codegen --config oapi-codegen.conf openapi.yaml

package main

import (
	"fmt"
	"os"

	_ "github.com/cosmtrek/air/runner" // used for hot reload
	"github.com/projectsyn/lieutenant-api/pkg/service"
)

// Version is the lieutenant-api version (set during build)
var (
	Version   = "unreleased"
	BuildDate = "now"
)

func main() {
	conf := service.APIConfig{
		APIVersion:       Version,
		Namespace:        os.Getenv("NAMESPACE"),
		OidcDiscoveryURL: os.Getenv("OIDC_DISCOVERY_URL"),
		OidcCLientID:     os.Getenv("OIDC_CLIENT_ID"),
	}

	e, err := service.NewAPIServer(conf)
	if err != nil {
		fmt.Fprintf(os.Stderr, err.Error()+"\n")
		os.Exit(1)
	}
	fmt.Println("Version: " + Version)
	fmt.Println("Build Date: " + BuildDate)

	e.Logger.Fatal(e.Start(":8080"))
}
