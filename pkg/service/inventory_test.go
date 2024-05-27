package service

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/labstack/echo/v4"

	"github.com/projectsyn/lieutenant-api/pkg/api"
)

func TestQueryInventory(t *testing.T) {
	e, _ := setupTest(t)

	query := "SELECT LAST(version,cloud) FROM mycluster"
	result := testutil.NewRequest().
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Get("/inventory?q="+url.QueryEscape(query)).
		Go(t, e)
	requireHTTPCode(t, http.StatusInternalServerError, result)
}

func TestUpdateInventory(t *testing.T) {
	e, _ := setupTest(t)

	updateInventory := api.Inventory{
		Cluster: "cluster-a",
		Inventory: &map[string]interface{}{
			"fact":    "one",
			"another": "fact",
		},
	}
	result := testutil.NewRequest().
		Post("/inventory").
		WithJsonBody(updateInventory).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	requireHTTPCode(t, http.StatusInternalServerError, result)
}
