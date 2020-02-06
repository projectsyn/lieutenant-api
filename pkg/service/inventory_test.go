package service

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/labstack/echo/v4"
	"github.com/projectsyn/lieutenant-api/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestQueryInventory(t *testing.T) {
	e := setupTest(t)

	query := "SELECT LAST(version,cloud) FROM mycluster"
	result := testutil.NewRequest().
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Get(APIBasePath+"/inventory?q="+url.QueryEscape(query)).
		Go(t, e)
	assert.Equal(t, http.StatusInternalServerError, result.Code())
}

func TestUpdateInventory(t *testing.T) {
	e := setupTest(t)

	updateInventory := api.Inventory{
		Cluster: "cluster-a",
		Inventory: &map[string]interface{}{
			"fact":    "one",
			"another": "fact",
		},
	}
	result := testutil.NewRequest().
		Post(APIBasePath+"/inventory").
		WithJsonBody(updateInventory).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusInternalServerError, result.Code())
}
