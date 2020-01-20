package service

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// NewService creates a new API implemenation
func NewService() *APIImpl {
	return &APIImpl{}
}

// APIImpl implements the API interface
type APIImpl struct{}

// Healthz implements the API health check
func (s *APIImpl) Healthz(ctx echo.Context) error {
	return ctx.String(http.StatusOK, "ok")
}
