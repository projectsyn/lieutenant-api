package service

import (
	"encoding/json"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/projectsyn/lieutenant-api/pkg/api"
	synv1alpha1 "github.com/projectsyn/lieutenant-operator/pkg/apis/syn/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListTenants lists all tenants
func (s *APIImpl) ListTenants(c echo.Context) error {
	ctx := c.(*APIContext)
	tenantList := &synv1alpha1.TenantList{}
	if err := ctx.client.List(ctx.context, tenantList, client.InNamespace(s.namespace)); err != nil {
		return err
	}
	tenants := []api.Tenant{}
	for _, tenant := range tenantList.Items {
		apiTenant := api.NewAPITenantFromCRD(tenant)
		tenants = append(tenants, *apiTenant)
	}
	return ctx.JSON(http.StatusOK, tenants)
}

// CreateTenant creates a new tenant
func (s *APIImpl) CreateTenant(c echo.Context) error {
	ctx := c.(*APIContext)
	var newTenant *api.CreateTenantJSONRequestBody
	if err := ctx.Bind(&newTenant); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	if newTenant.GitRepo == nil ||
		newTenant.GitRepo.Url == nil ||
		*newTenant.GitRepo.Url == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "GitRepo URL is required")
	}
	apiTenant := &api.Tenant{
		TenantProperties: api.TenantProperties(*newTenant),
	}
	id, err := api.GenerateTenantID()
	if err != nil {
		return err
	}
	apiTenant.TenantId = id
	tenant := api.NewCRDFromAPITenant(*apiTenant)
	tenant.Namespace = s.namespace
	if err := ctx.client.Create(ctx.context, tenant); err != nil {
		return err
	}
	apiTenant = api.NewAPITenantFromCRD(*tenant)
	return ctx.JSON(http.StatusCreated, apiTenant)
}

// DeleteTenant deletes a tenant
func (s *APIImpl) DeleteTenant(c echo.Context, tenantID api.TenantIdParameter) error {
	ctx := c.(*APIContext)

	deleteTenant := &synv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      string(tenantID),
			Namespace: s.namespace,
		},
	}
	if err := ctx.client.Delete(ctx.context, deleteTenant); err != nil {
		return err
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GetTenant gets a tenant
func (s *APIImpl) GetTenant(c echo.Context, tenantID api.TenantIdParameter) error {
	ctx := c.(*APIContext)

	tenant := &synv1alpha1.Tenant{}
	if err := ctx.client.Get(ctx.context, client.ObjectKey{Name: string(tenantID), Namespace: s.namespace}, tenant); err != nil {
		return err
	}
	apiTenant := api.NewAPITenantFromCRD(*tenant)
	return ctx.JSON(http.StatusOK, apiTenant)
}

// UpdateTenant udpates a tenant
func (s *APIImpl) UpdateTenant(c echo.Context, tenantID api.TenantIdParameter) error {
	ctx := c.(*APIContext)

	var patchTenant api.TenantProperties
	dec := json.NewDecoder(ctx.Request().Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&patchTenant); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	existingTenant := &synv1alpha1.Tenant{}
	if err := ctx.client.Get(ctx.context, client.ObjectKey{Name: string(tenantID), Namespace: s.namespace}, existingTenant); err != nil {
		return err
	}
	if patchTenant.DisplayName != nil {
		existingTenant.Spec.DisplayName = *patchTenant.DisplayName
	}
	if patchTenant.GitRepo != nil {
		if patchTenant.GitRepo.Url != nil {
			existingTenant.Spec.GitRepoURL = *patchTenant.GitRepo.Url
		}
	}
	if err := ctx.client.Update(ctx.context, existingTenant); err != nil {
		return err
	}
	apiTenant := api.NewAPITenantFromCRD(*existingTenant)
	return ctx.JSON(http.StatusOK, apiTenant)
}
