//go:generate go run github.com/deepmap/oapi-codegen/v2/cmd/oapi-codegen --config oapi-codegen.conf openapi.yaml

package main

import (
	"fmt"
	"os"

	_ "github.com/cosmtrek/air/runner" // used for hot reload
	"github.com/go-logr/logr"
	"github.com/go-logr/logr/funcr"
	"github.com/projectsyn/lieutenant-api/pkg/service"
	crlog "sigs.k8s.io/controller-runtime/pkg/log"
)

// Version is the lieutenant-api version (set during build)
var (
	Version   = "unreleased"
	BuildDate = "now"
)

func main() {
	crlog.SetLogger(newStdoutLogger())

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

func newStdoutLogger() logr.Logger {
	return funcr.New(func(prefix, args string) {
		if prefix != "" {
			fmt.Printf("%s: %s\n", prefix, args)
		} else {
			fmt.Println(args)
		}
	}, funcr.Options{})
}
