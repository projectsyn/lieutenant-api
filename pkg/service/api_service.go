package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	oapimiddleware "github.com/deepmap/oapi-codegen/pkg/middleware"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/getkin/kin-openapi/openapi3filter"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/projectsyn/lieutenant-api/pkg/api"
	swaggerui "github.com/projectsyn/lieutenant-api/swagger-ui"
)

// APIImpl implements the API interface
type APIImpl struct {
	namespace string

	// Metadata on the API itself
	metadata api.Metadata
}

// APIConfig holds the config options for the API
type APIConfig struct {
	APIVersion string

	Namespace string

	OidcDiscoveryURL string
	OidcCLientID     string
}

// APIContext is a custom echo context
type APIContext struct {
	echo.Context
	client client.Client
}

var (
	swaggerJSON []byte
)

// NewAPIServer instantiates a new Echo API server
func NewAPIServer(conf APIConfig, k8sMiddleware ...KubernetesAuth) (*echo.Echo, error) {
	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("error loading swagger spec: %w", err)
	}

	// Clear out the servers array in the swagger spec, that skips validating
	// that server names match. We don't know how/where this thing will be run.
	swagger.Servers = nil

	swaggerJSON, err = swagger.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("error marshalling swagger spec: %w", err)
	}

	namespace := conf.Namespace
	if len(namespace) == 0 {
		namespace = "default"
	}

	apiImpl := &APIImpl{
		namespace: namespace,
		metadata: api.Metadata{
			ApiVersion: conf.APIVersion,
		},
	}
	if conf.OidcCLientID != "" || conf.OidcDiscoveryURL != "" {
		apiImpl.metadata.Oidc = &api.OIDCConfig{
			ClientId:     conf.OidcCLientID,
			DiscoveryUrl: conf.OidcDiscoveryURL,
		}
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

// Discovery implements the API dicovery endpoint
func (s *APIImpl) Discovery(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, &s.metadata)
}

// Healthz implements the API health check
func (s *APIImpl) Healthz(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "ok")
}

// Docs serves the swagger UI
func (s *APIImpl) Docs(ctx echo.Context) error {
	return ctx.HTMLBlob(http.StatusOK, swaggerui.SwaggerHTML)
}

// Openapi serves the JSON spec
func (s *APIImpl) Openapi(ctx echo.Context) error {
	return ctx.JSONBlob(http.StatusOK, swaggerJSON)
}
