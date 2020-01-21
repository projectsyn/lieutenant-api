package service

import (
	"net/http"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/projectsyn/lieutenant/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestListTenants(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

	result := testutil.NewRequest().
		Get(APIBasePath+"/tenants").
		Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	tenants := &[]api.Tenant{}
	err = result.UnmarshalJsonToObject(tenants)
	assert.NoError(t, err)
	assert.NotNil(t, tenants)
	assert.GreaterOrEqual(t, len(*tenants), 1)
}

func TestCreateTenant(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

	newTenant := api.TenantProperties{
		Name:        "tenant-a",
		DisplayName: pointer.ToString("My test Tenant"),
	}
	result := testutil.NewRequest().
		Post(APIBasePath+"/tenants").
		WithJsonBody(newTenant).
		Go(t, e)
	assert.Equal(t, http.StatusCreated, result.Code())
	tenant := &api.Tenant{}
	err = result.UnmarshalJsonToObject(tenant)
	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.NotEmpty(t, tenant.Id)
	assert.Equal(t, tenant.Name, newTenant.Name)
}

func TestCreateTenantFail(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

	result := testutil.NewRequest().
		Post(APIBasePath+"/tenants/").
		WithJsonContentType().
		WithBody([]byte("invalid-body")).
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err = result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.NotEmpty(t, reason.Reason)
}

func TestCreateTenantEmpty(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

	result := testutil.NewRequest().
		Post(APIBasePath+"/tenants/").
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err = result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.NotEmpty(t, reason.Reason)
}

func TestTenantDelete(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

	result := testutil.NewRequest().
		Delete(APIBasePath+"/tenants/1").
		Go(t, e)
	assert.Equal(t, http.StatusNoContent, result.Code())
}

func TestTenantGet(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)
	id := "haevechee2ethot"
	result := testutil.NewRequest().
		Get(APIBasePath+"/tenants/"+id).
		Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	tenant := &api.Tenant{}
	err = result.UnmarshalJsonToObject(tenant)
	assert.NoError(t, err)
	assert.NotNil(t, tenant)
	assert.NotEmpty(t, tenant.Id)
	assert.NotEmpty(t, tenant.Name)
	assert.Equal(t, api.NewTenantID(id), tenant.TenantId)
}

func TestTenantUpdateEmpty(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

	result := testutil.NewRequest().
		Patch(APIBasePath+"/tenants/1").
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err = result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.NotEmpty(t, reason.Reason)
}

func TestTenantUpdate(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

	updateTenant := &api.TenantProperties{
		DisplayName: pointer.ToString("New Name"),
	}
	result := testutil.NewRequest().
		Patch(APIBasePath+"/tenants/1/").
		WithJsonBody(updateTenant).
		WithContentType("application/merge-patch+json").
		Go(t, e)
	assert.Equal(t, http.StatusNoContent, result.Code())
}
