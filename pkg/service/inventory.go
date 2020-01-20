package service

import (
	"github.com/labstack/echo/v4"
	"github.com/projectsyn/lieutenant/pkg/api"
	"net/http"
)

// QueryInventory queries the inventory
func (s *APIImpl) QueryInventory(ctx echo.Context, params api.QueryInventoryParams) error {
	return ctx.JSON(http.StatusOK, api.Inventory{
		Cluster: sampleCluster.Name,
		Inventory: &map[string]interface{}{
			"some": "info",
		},
	})
}

// UpdateInventory updates an inventory entry
func (s *APIImpl) UpdateInventory(ctx echo.Context) error {
	return ctx.NoContent(http.StatusCreated)
}
