package service

import (
	"github.com/labstack/echo/v4"
	"github.com/projectsyn/lieutenant/pkg/api"
)

func (s *APIImpl) ListTenants(ctx echo.Context) error {
	return nil
}

func (s *APIImpl) CreateTenant(ctx echo.Context) error {
	return nil
}

func (s *APIImpl) DeleteTenant(ctx echo.Context, tenantId api.TenantIdParameter) error {
	return nil
}

func (s *APIImpl) GetTenant(ctx echo.Context, tenantId api.TenantIdParameter) error {
	return nil
}

func (s *APIImpl) UpdateTenant(ctx echo.Context, tenantId api.TenantIdParameter) error {
	return nil
}
