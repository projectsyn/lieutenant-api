package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	oapimiddleware "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/projectsyn/lieutenant-api/pkg/api"

	"github.com/rakyll/statik/fs"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	// Import swagger-ui static files
	_ "github.com/projectsyn/lieutenant-api/pkg/swaggerui"
)

// APIImpl implements the API interface
type APIImpl struct {
	namespace string
}

// APIContext is a custom echo context
type APIContext struct {
	echo.Context
	client  client.Client
}

var (
	statikFS    http.FileSystem
	swaggerJSON []byte
)

// NewAPIServer instantiates a new Echo API server
func NewAPIServer(k8sMiddleware ...KubernetesAuth) (*echo.Echo, error) {
	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("Error loading swagger spec: %w", err)
	}
	swaggerJSON, err = swagger.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("Error marshalling swagger spec: %w", err)
	}

	namespace := os.Getenv("NAMESPACE")
	if len(namespace) == 0 {
		namespace = "default"
	}

	apiImpl := &APIImpl{
		namespace: namespace,
	}

	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool { return strings.HasSuffix(c.Path(), "/healthz") },
	}))
	e.Use(middleware.Recover())
	e.Pre(middleware.RemoveTrailingSlash())

	e.HTTPErrorHandler = customHTTPErrorHandler

	openapi3filter.RegisterBodyDecoder(api.ContentJSONPatch, func(body io.Reader, header http.Header, schema *openapi3.SchemaRef, encFn openapi3filter.EncodingFn) (interface{}, error) {
		var value interface{}
		if err := json.NewDecoder(body).Decode(&value); err != nil {
			return nil, &openapi3filter.ParseError{Kind: openapi3filter.KindInvalidFormat, Cause: err}
		}
		return value, nil
	})

	options := &oapimiddleware.Options{
		Options: openapi3filter.Options{
			AuthenticationFunc: func(c context.Context, input *openapi3filter.AuthenticationInput) error {
				return nil
			},
		},
	}
	e.Use(oapimiddleware.OapiRequestValidatorWithOptions(swagger, options))
	if len(k8sMiddleware) == 0 {
		e.Use(DefaultKubernetesAuth.JWTAuth)
	} else {
		for _, middle := range k8sMiddleware {
			e.Use(middle.JWTAuth)
		}
	}
	api.RegisterHandlers(e, apiImpl)

	statikFS, err = fs.New()
	if err != nil {
		return nil, err
	}
	return e, nil
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	message := err.Error()
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if m, ok := he.Message.(string); ok {
			message = m
		}
	} else if apiErr, ok := err.(errors.APIStatus); ok {
		code = int(apiErr.Status().Code)
		message = strings.ReplaceAll(apiErr.Status().Message, "\"", "'")
	}
	reason := api.Reason{
		Reason: message,
	}
	c.JSON(code, reason)
	c.Logger().Error(err)
}

// Healthz implements the API health check
func (s *APIImpl) Healthz(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "ok")
}

// Docs serves the swagger UI
func (s *APIImpl) Docs(ctx echo.Context) error {
	file, err := fs.ReadFile(statikFS, "/index.html")
	if err != nil {
		return err
	}
	return ctx.HTMLBlob(http.StatusOK, file)
}

// Openapi serves the JSON spec
func (s *APIImpl) Openapi(ctx echo.Context) error {
	return ctx.JSONBlob(http.StatusOK, swaggerJSON)
}
