package service

import (
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
	"github.com/projectsyn/lieutenant-api/pkg/api"
)

// APIBasePath the base path of the API
const APIBasePath = "/api"

// APIImpl implements the API interface
type APIImpl struct{}

// NewAPIServer instantiates a new Echo API server
func NewAPIServer() (*echo.Echo, error) {
	swagger, err := api.GetSwagger()
	if err != nil {
		return nil, fmt.Errorf("Error loading swagger spec\n: %s", err)
	}

	apiImpl := &APIImpl{}

	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Skipper: func(c echo.Context) bool { return strings.HasSuffix(c.Path(), "/healthz") },
	}))
	e.Use(middleware.Recover())
	e.Pre(middleware.RemoveTrailingSlash())

	e.HTTPErrorHandler = customHTTPErrorHandler

	apiGroup := e.Group(APIBasePath)
	openapi3filter.RegisterBodyDecoder("application/merge-patch+json", func(body io.Reader, header http.Header, schema *openapi3.SchemaRef, encFn openapi3filter.EncodingFn) (interface{}, error) {
		var value interface{}
		if err := json.NewDecoder(body).Decode(&value); err != nil {
			return nil, &openapi3filter.ParseError{Kind: openapi3filter.KindInvalidFormat, Cause: err}
		}
		return value, nil
	})

	apiGroup.Use(oapimiddleware.OapiRequestValidator(swagger))
	api.RegisterHandlers(apiGroup, apiImpl)

	return e, nil
}

func customHTTPErrorHandler(err error, c echo.Context) {
	code := http.StatusInternalServerError
	msg := err.Error()
	if he, ok := err.(*echo.HTTPError); ok {
		code = he.Code
		if m, ok := he.Message.(string); ok {
			msg = m
		}
	}
	reason := api.Reason{
		Reason: msg,
	}
	c.JSON(code, reason)
	c.Logger().Error(err)
}

// Healthz implements the API health check
func (s *APIImpl) Healthz(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "ok")
}
