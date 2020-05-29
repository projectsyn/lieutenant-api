package service

import (
	"context"
	"net/http"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/labstack/echo/v4"
	"github.com/projectsyn/lieutenant-api/pkg/api"
	synv1alpha1 "github.com/projectsyn/lieutenant-operator/pkg/apis/syn/v1alpha1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
)

func TestListTenants(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Get("/tenants/").
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	tenants := []api.Tenant{}
	err := result.UnmarshalJsonToObject(&tenants)
	assert.NoError(t, err)
	assert.NotNil(t, tenants)
	assert.GreaterOrEqual(t, len(tenants), 2)
	assert.Equal(t, tenantA.Spec.DisplayName, *tenants[0].DisplayName)
	assert.Equal(t, tenantB.Spec.DisplayName, *tenants[1].DisplayName)
	assert.Equal(t, string(tenantB.Spec.GitRepoTemplate.RepoType), *tenants[1].GitRepo.Type)
}

func TestCreateTenant(t *testing.T) {
	e, _ := setupTest(t)

	newTenant := api.TenantProperties{
		DisplayName: pointer.ToString("My test Tenant"),
	}
	result := testutil.NewRequest().
		Post("/tenants").
		WithJsonBody(newTenant).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusCreated, result.Code())
	tenant := &api.Tenant{}
	err := result.UnmarshalJsonToObject(tenant)
	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Contains(t, tenant.Id, api.TenantIDPrefix)
	assert.Equal(t, newTenant.DisplayName, tenant.DisplayName)
}

func TestCreateTenantFail(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Post("/tenants/").
		WithJsonContentType().
		WithBody([]byte("invalid-body")).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.Contains(t, reason.Reason, "invalid character")
}

func TestCreateTenantEmpty(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Post("/tenants/").
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.Contains(t, reason.Reason, "must have a value")
}

func TestTenantDelete(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Delete("/tenants/"+tenantA.Name).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusNoContent, result.Code())
}

func TestTenantGet(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Get("/tenants/"+tenantA.Name).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	tenant := &api.Tenant{}
	err := result.UnmarshalJsonToObject(tenant)
	assert.NoError(t, err)
	assert.Equal(t, tenantA.Name, string(tenant.Id))
	assert.Equal(t, tenantA.Spec.DisplayName, *tenant.DisplayName)
	assert.Equal(t, tenantA.Spec.GitRepoURL, *tenant.GitRepo.Url)
}

func TestTenantUpdateEmpty(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Patch("/tenants/1").
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.Contains(t, reason.Reason, "must have a value")
}

func TestTenantUpdateUnknown(t *testing.T) {
	e, _ := setupTest(t)

	updateTenant := map[string]string{
		"displayName": "newName",
		"some":        "field",
		"unknown":     "true",
	}

	result := testutil.NewRequest().
		Patch("/tenants/1").
		WithJsonBody(updateTenant).
		WithContentType(api.ContentJSONPatch).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.Contains(t, reason.Reason, "unknown field")
}

func TestTenantUpdate(t *testing.T) {
	e, _ := setupTest(t)
	newDisplayName := "New Tenant Name"

	updateTenant := map[string]interface{}{
		"displayName": newDisplayName,
		"gitRepo": map[string]string{
			"url": "newURL",
		},
	}
	result := testutil.NewRequest().
		Patch("/tenants/"+tenantB.Name).
		WithJsonBody(updateTenant).
		WithContentType(api.ContentJSONPatch).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	tenant := &api.Tenant{}
	err := result.UnmarshalJsonToObject(tenant)
	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.Contains(t, string(tenant.Id), tenantB.Name)
	assert.Equal(t, newDisplayName, *tenant.DisplayName)
}

func TestTenantUpdateDisplayName(t *testing.T) {
	e, client := setupTest(t)
	newDisplayName := "New Tenant Name"

	updateTenant := map[string]string{
		"displayName": newDisplayName,
	}
	assert.NotEqual(t, newDisplayName, tenantB.Spec.DisplayName)
	result := testutil.NewRequest().
		Patch("/tenants/"+tenantB.Name).
		WithJsonBody(updateTenant).
		WithContentType(api.ContentJSONPatch).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	tenant := &api.Tenant{}
	err := result.UnmarshalJsonToObject(tenant)
	assert.NoError(t, err)
	assert.Equal(t, newDisplayName, *tenant.DisplayName)
	tenantObj := &synv1alpha1.Tenant{}
	err = client.Get(context.TODO(), types.NamespacedName{
		Namespace: "default",
		Name:      tenantB.Name,
	}, tenantObj)
	assert.NoError(t, err)
	assert.Equal(t, newDisplayName, tenantObj.Spec.DisplayName)
}
