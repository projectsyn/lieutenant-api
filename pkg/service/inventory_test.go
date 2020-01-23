package service

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/projectsyn/lieutenant/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestQueryInventory(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

	query := "SELECT LAST(version,cloud) FROM mycluster"
	result := testutil.NewRequest().
		Get(APIBasePath+"/inventory?q="+url.QueryEscape(query)).
		Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	inventory := api.Inventory{}
	err = result.UnmarshalJsonToObject(&inventory)
	assert.NoError(t, err)
	assert.NotEmpty(t, inventory.Cluster)
}

func TestUpdateInventory(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

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
		Go(t, e)
	assert.Equal(t, http.StatusCreated, result.Code())
}
