package service

import (
	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	"github.com/projectsyn/lieutenant/pkg/api"
	"net/http"
)

var sampleTenant = api.Tenant{
	TenantId: api.NewTenantID("ut0uaVae"),
	TenantProperties: api.TenantProperties{
		Name:        pointer.ToString("tenant-a"),
		DisplayName: pointer.ToString("Tenant A corp."),
		GitRepo:     pointer.ToString("ssh://git@github.com/projectsyn/cluster-catalog.git"),
	},
}

// ListTenants lists all tenants
func (s *APIImpl) ListTenants(ctx echo.Context) error {
	return ctx.JSON(http.StatusOK, []api.Tenant{sampleTenant})
}

// CreateTenant creates a new tenant
func (s *APIImpl) CreateTenant(ctx echo.Context) error {
	return ctx.JSON(http.StatusCreated, sampleTenant)
}

// DeleteTenant deletes a tenant
func (s *APIImpl) DeleteTenant(ctx echo.Context, tenantID api.TenantIdParameter) error {
	return ctx.NoContent(http.StatusNoContent)
}

// GetTenant gets a tenant
func (s *APIImpl) GetTenant(ctx echo.Context, tenantID api.TenantIdParameter) error {
	t := sampleTenant
	t.Id = api.Id(tenantID)
	return ctx.JSON(http.StatusOK, t)
}

// UpdateTenant udpates a tenant
func (s *APIImpl) UpdateTenant(ctx echo.Context, tenantID api.TenantIdParameter) error {
	return ctx.NoContent(http.StatusNoContent)
}
