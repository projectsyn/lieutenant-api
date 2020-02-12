package service

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/projectsyn/lieutenant-api/pkg/api"
)

// QueryInventory queries the inventory
func (s *APIImpl) QueryInventory(ctx echo.Context, params api.QueryInventoryParams) error {
	return echo.NewHTTPError(http.StatusInternalServerError, "Not implemented")
}

// UpdateInventory updates an inventory entry
func (s *APIImpl) UpdateInventory(ctx echo.Context) error {
	return echo.NewHTTPError(http.StatusInternalServerError, "Not implemented")
}
