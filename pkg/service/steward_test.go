package service

import (
	"net/http"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/projectsyn/lieutenant-api/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestInstallSteward(t *testing.T) {
	e := setupTest(t)

	result := testutil.NewRequest().
		Get(APIBasePath+"/install/steward.json?token=haevechee2ethot").
		Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	manifests := map[string]interface{}{}
	err = result.UnmarshalJsonToObject(&manifests)
	assert.NoError(t, err)
	assert.NotEmpty(t, manifests["apiVersion"])
}

func TestInstallStewardNoToken(t *testing.T) {
	e := setupTest(t)

	result := testutil.NewRequest().
		Get(APIBasePath+"/install/steward.json").
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.NotEmpty(t, reason.Reason)
}
