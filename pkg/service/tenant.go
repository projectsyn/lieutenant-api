package service

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	synv1alpha1 "github.com/projectsyn/lieutenant-operator/api/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/projectsyn/lieutenant-api/pkg/api"
)

// DefaultAPISecretRefNameEnvVar is the name of the env var which specifies the default APISecretRef name
const DefaultAPISecretRefNameEnvVar = "DEFAULT_API_SECRET_REF_NAME"

// ListTenants lists all tenants
func (s *APIImpl) ListTenants(c echo.Context) error {
	ctx := c.(*APIContext)
	tenantList := &synv1alpha1.TenantList{}
	if err := ctx.client.List(ctx.Request().Context(), tenantList, client.InNamespace(s.namespace)); err != nil {
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
	apiTenant := api.Tenant(*newTenant)

	tenant, err := api.NewCRDFromAPITenant(apiTenant)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return s.createTenant(ctx, tenant)
}

func (s *APIImpl) createTenant(ctx *APIContext, tenant *synv1alpha1.Tenant) error {
	tenant.Namespace = s.namespace
	if name, ok := os.LookupEnv(DefaultAPISecretRefNameEnvVar); ok &&
		tenant.Spec.GitRepoTemplate != nil &&
		tenant.Spec.GitRepoTemplate.RepoType == synv1alpha1.AutoRepoType {
		tenant.Spec.GitRepoTemplate.APISecretRef.Name = name
	}
	if err := ctx.client.Create(ctx.Request().Context(), tenant); err != nil {
		return err
	}
	apiTenant := *api.NewAPITenantFromCRD(*tenant)
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
	if err := ctx.client.Delete(ctx.Request().Context(), deleteTenant); err != nil {
		return err
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GetTenant gets a tenant
func (s *APIImpl) GetTenant(c echo.Context, tenantID api.TenantIdParameter) error {
	ctx := c.(*APIContext)

	tenant := &synv1alpha1.Tenant{}
	if err := ctx.client.Get(ctx.Request().Context(), client.ObjectKey{Name: string(tenantID), Namespace: s.namespace}, tenant); err != nil {
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
	if err := ctx.client.Get(ctx.Request().Context(), client.ObjectKey{Name: string(tenantID), Namespace: s.namespace}, existingTenant); err != nil {
		return err
	}

	api.SyncCRDFromAPITenant(patchTenant, existingTenant)

	return s.updateTenant(ctx, existingTenant)
}

func (s *APIImpl) updateTenant(ctx *APIContext, tenant *synv1alpha1.Tenant) error {

	if err := ctx.client.Update(ctx.Request().Context(), tenant); err != nil {
		return err
	}
	apiTenant := api.NewAPITenantFromCRD(*tenant)
	return ctx.JSON(http.StatusOK, apiTenant)
}

// PutTenant udpates or creates a tenant
func (s *APIImpl) PutTenant(c echo.Context, tenantID api.TenantIdParameter) error {
	ctx := c.(*APIContext)
	var newTenant *api.CreateTenantJSONRequestBody
	if err := ctx.Bind(&newTenant); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	apiTenant := api.Tenant(*newTenant)
	apiTenant.Id = pointer.To(api.Id(tenantID))

	tenant, err := api.NewCRDFromAPITenant(apiTenant)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	found := &synv1alpha1.Tenant{}
	err = ctx.client.Get(ctx.Request().Context(), client.ObjectKey{Name: string(tenantID), Namespace: s.namespace}, found)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}
	if errors.IsNotFound(err) {
		return s.createTenant(ctx, tenant)
	}

	found.Spec = tenant.Spec
	found.Annotations = tenant.Annotations
	return s.updateTenant(ctx, found)
}
