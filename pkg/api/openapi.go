// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.9.0 DO NOT EDIT.
package api

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"

	"github.com/deepmap/oapi-codegen/pkg/runtime"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/labstack/echo/v4"
)

const (
	BearerAuthScopes = "BearerAuth.Scopes"
)

// Unstructured key value map containing arbitrary metadata
type Annotations map[string]interface{}

// Cluster defines model for Cluster.
type Cluster struct {
	// Embedded struct due to allOf(#/components/schemas/ClusterId)
	ClusterId `yaml:",inline"`
	// Embedded struct due to allOf(#/components/schemas/ClusterTenant)
	ClusterTenant `yaml:",inline"`
	// Embedded struct due to allOf(#/components/schemas/ClusterProperties)
	ClusterProperties `yaml:",inline"`
}

// Facts about a cluster object. Statically configured key/value pairs.
type ClusterFacts map[string]interface{}

// ClusterId defines model for ClusterId.
type ClusterId struct {
	// A unique object identifier string. Automatically generated by the API on creation (in the form "<letter>-<adjective>-<noun>-<digits>" where all letters are lowercase, max 63 characters in total).
	Id Id `json:"id"`
}

// A cluster defition object.
// The Git repository is usually managed by the API and autogenerated.
// The sshDeployKey will be managed by Steward
type ClusterProperties struct {
	// Unstructured key value map containing arbitrary metadata
	Annotations *Annotations `json:"annotations,omitempty"`

	// Display Name of the cluster
	DisplayName *string `json:"displayName,omitempty"`

	// Dynamic facts about a cluster object. Are periodically udpated by Project Syn and should not be set manually.
	DynamicFacts *DynamicClusterFacts `json:"dynamicFacts,omitempty"`

	// Facts about a cluster object. Statically configured key/value pairs.
	Facts *ClusterFacts `json:"facts,omitempty"`

	// Configuration Git repository, usually generated by the API
	GitRepo *GitRepo `json:"gitRepo,omitempty"`

	// Git revision to use with the global configruation git repository.
	// This takes precedence over the revision configured on the Tenant.
	GlobalGitRepoRevision *string `json:"globalGitRepoRevision,omitempty"`

	// URL to fetch install manifests for Steward cluster agent. This will only be set if the cluster's token is still valid.
	InstallURL *string `json:"installURL,omitempty"`

	// Git revision to use with the tenant configruation git repository.
	// This takes precedence over the revision configured on the Tenant.
	TenantGitRepoRevision *string `json:"tenantGitRepoRevision,omitempty"`
}

// ClusterTenant defines model for ClusterTenant.
type ClusterTenant struct {
	// Id of the tenant this cluster belongs to
	Tenant string `json:"tenant"`
}

// Dynamic facts about a cluster object. Are periodically udpated by Project Syn and should not be set manually.
type DynamicClusterFacts map[string]interface{}

// Configuration Git repository, usually generated by the API
type GitRepo struct {
	// SSH public key / deploy key for clusterconfiguration catalog Git repository. This property is managed by Steward.
	DeployKey *string `json:"deployKey,omitempty"`

	// SSH known hosts of the git server (multiline possible for multiple keys)
	HostKeys *string `json:"hostKeys,omitempty"`

	// Specifies if a repo should be managed by the git controller. A value of 'unmanaged' means it's not manged by the controller
	Type *string `json:"type,omitempty"`

	// Full URL of the git repo
	Url *string `json:"url,omitempty"`
}

// A unique object identifier string. Automatically generated by the API on creation (in the form "<letter>-<adjective>-<noun>-<digits>" where all letters are lowercase, max 63 characters in total).
type Id string

// Inventory data of a cluster
type Inventory struct {
	Cluster   string                  `json:"cluster"`
	Inventory *map[string]interface{} `json:"inventory,omitempty"`
}

// Metadata defines model for Metadata.
type Metadata struct {
	ApiVersion string      `json:"apiVersion"`
	Oidc       *OIDCConfig `json:"oidc,omitempty"`
}

// OIDCConfig defines model for OIDCConfig.
type OIDCConfig struct {
	ClientId     string `json:"clientId"`
	DiscoveryUrl string `json:"discoveryUrl"`
}

// A reason for responses
type Reason struct {
	// The reason message
	Reason string `json:"reason"`
}

// Revision defines model for Revision.
type Revision struct {
	// Revision to use with a git repository.
	Revision *string `json:"revision,omitempty"`
}

// RevisionedGitRepo defines model for RevisionedGitRepo.
type RevisionedGitRepo struct {
	// Embedded struct due to allOf(#/components/schemas/GitRepo)
	GitRepo `yaml:",inline"`
	// Embedded struct due to allOf(#/components/schemas/Revision)
	Revision `yaml:",inline"`
}

// Tenant defines model for Tenant.
type Tenant struct {
	// Embedded struct due to allOf(#/components/schemas/TenantId)
	TenantId `yaml:",inline"`
	// Embedded struct due to allOf(#/components/schemas/TenantProperties)
	TenantProperties `yaml:",inline"`
}

// TenantId defines model for TenantId.
type TenantId struct {
	// A unique object identifier string. Automatically generated by the API on creation (in the form "<letter>-<adjective>-<noun>-<digits>" where all letters are lowercase, max 63 characters in total).
	Id Id `json:"id"`
}

// A tenant definition object.
// The Git repository is usually managed by the API and autogenerated.
// All properties except name are optional on creation.
type TenantProperties struct {
	// Unstructured key value map containing arbitrary metadata
	Annotations *Annotations `json:"annotations,omitempty"`

	// Display name of the tenant
	DisplayName *string            `json:"displayName,omitempty"`
	GitRepo     *RevisionedGitRepo `json:"gitRepo,omitempty"`

	// Git revision to use with the global configruation git repository.
	GlobalGitRepoRevision *string `json:"globalGitRepoRevision,omitempty"`

	// Full URL of the global configuration git repo
	GlobalGitRepoURL *string `json:"globalGitRepoURL,omitempty"`
}

// A unique object identifier string. Automatically generated by the API on creation (in the form "<letter>-<adjective>-<noun>-<digits>" where all letters are lowercase, max 63 characters in total).
type ClusterIdParameter Id

// A unique object identifier string. Automatically generated by the API on creation (in the form "<letter>-<adjective>-<noun>-<digits>" where all letters are lowercase, max 63 characters in total).
type TenantIdParameter Id

// A reason for responses
type Default Reason

// ListClustersParams defines parameters for ListClusters.
type ListClustersParams struct {
	// Filter clusters by tenant id
	Tenant *string `json:"tenant,omitempty"`

	// Sort list by field
	SortBy *ListClustersParamsSortBy `json:"sort_by,omitempty"`
}

// ListClustersParamsSortBy defines parameters for ListClusters.
type ListClustersParamsSortBy string

// CreateClusterJSONBody defines parameters for CreateCluster.
type CreateClusterJSONBody Cluster

// PutClusterJSONBody defines parameters for PutCluster.
type PutClusterJSONBody Cluster

// InstallStewardParams defines parameters for InstallSteward.
type InstallStewardParams struct {
	// Initial bootstrap token
	Token *string `json:"token,omitempty"`
}

// QueryInventoryParams defines parameters for QueryInventory.
type QueryInventoryParams struct {
	// InfluxQL query string
	Q *string `json:"q,omitempty"`
}

// UpdateInventoryJSONBody defines parameters for UpdateInventory.
type UpdateInventoryJSONBody Inventory

// CreateTenantJSONBody defines parameters for CreateTenant.
type CreateTenantJSONBody Tenant

// PutTenantJSONBody defines parameters for PutTenant.
type PutTenantJSONBody Tenant

// CreateClusterJSONRequestBody defines body for CreateCluster for application/json ContentType.
type CreateClusterJSONRequestBody CreateClusterJSONBody

// PutClusterJSONRequestBody defines body for PutCluster for application/json ContentType.
type PutClusterJSONRequestBody PutClusterJSONBody

// UpdateInventoryJSONRequestBody defines body for UpdateInventory for application/json ContentType.
type UpdateInventoryJSONRequestBody UpdateInventoryJSONBody

// CreateTenantJSONRequestBody defines body for CreateTenant for application/json ContentType.
type CreateTenantJSONRequestBody CreateTenantJSONBody

// PutTenantJSONRequestBody defines body for PutTenant for application/json ContentType.
type PutTenantJSONRequestBody PutTenantJSONBody

// ServerInterface represents all server handlers.
type ServerInterface interface {
	// Lieutenant API Root
	// (GET /)
	Discovery(ctx echo.Context) error
	// Returns a list of clusters
	// (GET /clusters)
	ListClusters(ctx echo.Context, params ListClustersParams) error
	// Creates a new cluster
	// (POST /clusters)
	CreateCluster(ctx echo.Context) error
	// Deletes a cluster
	// (DELETE /clusters/{clusterId})
	DeleteCluster(ctx echo.Context, clusterId ClusterIdParameter) error
	// Returns all values of a cluster
	// (GET /clusters/{clusterId})
	GetCluster(ctx echo.Context, clusterId ClusterIdParameter) error
	// Updates a cluster
	// (PATCH /clusters/{clusterId})
	UpdateCluster(ctx echo.Context, clusterId ClusterIdParameter) error
	// Updates or creates a cluster
	// (PUT /clusters/{clusterId})
	PutCluster(ctx echo.Context, clusterId ClusterIdParameter) error
	// API documentation
	// (GET /docs)
	Docs(ctx echo.Context) error
	// API health check
	// (GET /healthz)
	Healthz(ctx echo.Context) error
	// Returns the Steward JSON installation manifest
	// (GET /install/steward.json)
	InstallSteward(ctx echo.Context, params InstallStewardParams) error
	// Returns inventory data according to query
	// (GET /inventory)
	QueryInventory(ctx echo.Context, params QueryInventoryParams) error
	// Write inventory data
	// (POST /inventory)
	UpdateInventory(ctx echo.Context) error
	// OpenAPI JSON spec
	// (GET /openapi.json)
	Openapi(ctx echo.Context) error
	// Returns a list of tenants
	// (GET /tenants)
	ListTenants(ctx echo.Context) error
	// Creates a new tenant
	// (POST /tenants)
	CreateTenant(ctx echo.Context) error
	// Deletes a tenant
	// (DELETE /tenants/{tenantId})
	DeleteTenant(ctx echo.Context, tenantId TenantIdParameter) error
	// Returns all values of a tenant
	// (GET /tenants/{tenantId})
	GetTenant(ctx echo.Context, tenantId TenantIdParameter) error
	// Updates a tenant
	// (PATCH /tenants/{tenantId})
	UpdateTenant(ctx echo.Context, tenantId TenantIdParameter) error
	// Updates or creates a tenant
	// (PUT /tenants/{tenantId})
	PutTenant(ctx echo.Context, tenantId TenantIdParameter) error
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
}

// Discovery converts echo context to params.
func (w *ServerInterfaceWrapper) Discovery(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.Discovery(ctx)
	return err
}

// ListClusters converts echo context to params.
func (w *ServerInterfaceWrapper) ListClusters(ctx echo.Context) error {
	var err error

	ctx.Set(BearerAuthScopes, []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params ListClustersParams
	// ------------- Optional query parameter "tenant" -------------

	err = runtime.BindQueryParameter("form", true, false, "tenant", ctx.QueryParams(), &params.Tenant)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter tenant: %s", err))
	}

	// ------------- Optional query parameter "sort_by" -------------

	err = runtime.BindQueryParameter("form", true, false, "sort_by", ctx.QueryParams(), &params.SortBy)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter sort_by: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ListClusters(ctx, params)
	return err
}

// CreateCluster converts echo context to params.
func (w *ServerInterfaceWrapper) CreateCluster(ctx echo.Context) error {
	var err error

	ctx.Set(BearerAuthScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateCluster(ctx)
	return err
}

// DeleteCluster converts echo context to params.
func (w *ServerInterfaceWrapper) DeleteCluster(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "clusterId" -------------
	var clusterId ClusterIdParameter

	err = runtime.BindStyledParameterWithLocation("simple", false, "clusterId", runtime.ParamLocationPath, ctx.Param("clusterId"), &clusterId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter clusterId: %s", err))
	}

	ctx.Set(BearerAuthScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DeleteCluster(ctx, clusterId)
	return err
}

// GetCluster converts echo context to params.
func (w *ServerInterfaceWrapper) GetCluster(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "clusterId" -------------
	var clusterId ClusterIdParameter

	err = runtime.BindStyledParameterWithLocation("simple", false, "clusterId", runtime.ParamLocationPath, ctx.Param("clusterId"), &clusterId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter clusterId: %s", err))
	}

	ctx.Set(BearerAuthScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetCluster(ctx, clusterId)
	return err
}

// UpdateCluster converts echo context to params.
func (w *ServerInterfaceWrapper) UpdateCluster(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "clusterId" -------------
	var clusterId ClusterIdParameter

	err = runtime.BindStyledParameterWithLocation("simple", false, "clusterId", runtime.ParamLocationPath, ctx.Param("clusterId"), &clusterId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter clusterId: %s", err))
	}

	ctx.Set(BearerAuthScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdateCluster(ctx, clusterId)
	return err
}

// PutCluster converts echo context to params.
func (w *ServerInterfaceWrapper) PutCluster(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "clusterId" -------------
	var clusterId ClusterIdParameter

	err = runtime.BindStyledParameterWithLocation("simple", false, "clusterId", runtime.ParamLocationPath, ctx.Param("clusterId"), &clusterId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter clusterId: %s", err))
	}

	ctx.Set(BearerAuthScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PutCluster(ctx, clusterId)
	return err
}

// Docs converts echo context to params.
func (w *ServerInterfaceWrapper) Docs(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.Docs(ctx)
	return err
}

// Healthz converts echo context to params.
func (w *ServerInterfaceWrapper) Healthz(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.Healthz(ctx)
	return err
}

// InstallSteward converts echo context to params.
func (w *ServerInterfaceWrapper) InstallSteward(ctx echo.Context) error {
	var err error

	// Parameter object where we will unmarshal all parameters from the context
	var params InstallStewardParams
	// ------------- Optional query parameter "token" -------------

	err = runtime.BindQueryParameter("form", true, false, "token", ctx.QueryParams(), &params.Token)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter token: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.InstallSteward(ctx, params)
	return err
}

// QueryInventory converts echo context to params.
func (w *ServerInterfaceWrapper) QueryInventory(ctx echo.Context) error {
	var err error

	ctx.Set(BearerAuthScopes, []string{""})

	// Parameter object where we will unmarshal all parameters from the context
	var params QueryInventoryParams
	// ------------- Optional query parameter "q" -------------

	err = runtime.BindQueryParameter("form", true, false, "q", ctx.QueryParams(), &params.Q)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter q: %s", err))
	}

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.QueryInventory(ctx, params)
	return err
}

// UpdateInventory converts echo context to params.
func (w *ServerInterfaceWrapper) UpdateInventory(ctx echo.Context) error {
	var err error

	ctx.Set(BearerAuthScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdateInventory(ctx)
	return err
}

// Openapi converts echo context to params.
func (w *ServerInterfaceWrapper) Openapi(ctx echo.Context) error {
	var err error

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.Openapi(ctx)
	return err
}

// ListTenants converts echo context to params.
func (w *ServerInterfaceWrapper) ListTenants(ctx echo.Context) error {
	var err error

	ctx.Set(BearerAuthScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.ListTenants(ctx)
	return err
}

// CreateTenant converts echo context to params.
func (w *ServerInterfaceWrapper) CreateTenant(ctx echo.Context) error {
	var err error

	ctx.Set(BearerAuthScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.CreateTenant(ctx)
	return err
}

// DeleteTenant converts echo context to params.
func (w *ServerInterfaceWrapper) DeleteTenant(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "tenantId" -------------
	var tenantId TenantIdParameter

	err = runtime.BindStyledParameterWithLocation("simple", false, "tenantId", runtime.ParamLocationPath, ctx.Param("tenantId"), &tenantId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter tenantId: %s", err))
	}

	ctx.Set(BearerAuthScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.DeleteTenant(ctx, tenantId)
	return err
}

// GetTenant converts echo context to params.
func (w *ServerInterfaceWrapper) GetTenant(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "tenantId" -------------
	var tenantId TenantIdParameter

	err = runtime.BindStyledParameterWithLocation("simple", false, "tenantId", runtime.ParamLocationPath, ctx.Param("tenantId"), &tenantId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter tenantId: %s", err))
	}

	ctx.Set(BearerAuthScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.GetTenant(ctx, tenantId)
	return err
}

// UpdateTenant converts echo context to params.
func (w *ServerInterfaceWrapper) UpdateTenant(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "tenantId" -------------
	var tenantId TenantIdParameter

	err = runtime.BindStyledParameterWithLocation("simple", false, "tenantId", runtime.ParamLocationPath, ctx.Param("tenantId"), &tenantId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter tenantId: %s", err))
	}

	ctx.Set(BearerAuthScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.UpdateTenant(ctx, tenantId)
	return err
}

// PutTenant converts echo context to params.
func (w *ServerInterfaceWrapper) PutTenant(ctx echo.Context) error {
	var err error
	// ------------- Path parameter "tenantId" -------------
	var tenantId TenantIdParameter

	err = runtime.BindStyledParameterWithLocation("simple", false, "tenantId", runtime.ParamLocationPath, ctx.Param("tenantId"), &tenantId)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, fmt.Sprintf("Invalid format for parameter tenantId: %s", err))
	}

	ctx.Set(BearerAuthScopes, []string{""})

	// Invoke the callback with all the unmarshalled arguments
	err = w.Handler.PutTenant(ctx, tenantId)
	return err
}

// This is a simple interface which specifies echo.Route addition functions which
// are present on both echo.Echo and echo.Group, since we want to allow using
// either of them for path registration
type EchoRouter interface {
	CONNECT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	DELETE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	GET(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	HEAD(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	OPTIONS(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PATCH(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	POST(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	PUT(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
	TRACE(path string, h echo.HandlerFunc, m ...echo.MiddlewareFunc) *echo.Route
}

// RegisterHandlers adds each server route to the EchoRouter.
func RegisterHandlers(router EchoRouter, si ServerInterface) {
	RegisterHandlersWithBaseURL(router, si, "")
}

// Registers handlers, and prepends BaseURL to the paths, so that the paths
// can be served under a prefix.
func RegisterHandlersWithBaseURL(router EchoRouter, si ServerInterface, baseURL string) {

	wrapper := ServerInterfaceWrapper{
		Handler: si,
	}

	router.GET(baseURL+"/", wrapper.Discovery)
	router.GET(baseURL+"/clusters", wrapper.ListClusters)
	router.POST(baseURL+"/clusters", wrapper.CreateCluster)
	router.DELETE(baseURL+"/clusters/:clusterId", wrapper.DeleteCluster)
	router.GET(baseURL+"/clusters/:clusterId", wrapper.GetCluster)
	router.PATCH(baseURL+"/clusters/:clusterId", wrapper.UpdateCluster)
	router.PUT(baseURL+"/clusters/:clusterId", wrapper.PutCluster)
	router.GET(baseURL+"/docs", wrapper.Docs)
	router.GET(baseURL+"/healthz", wrapper.Healthz)
	router.GET(baseURL+"/install/steward.json", wrapper.InstallSteward)
	router.GET(baseURL+"/inventory", wrapper.QueryInventory)
	router.POST(baseURL+"/inventory", wrapper.UpdateInventory)
	router.GET(baseURL+"/openapi.json", wrapper.Openapi)
	router.GET(baseURL+"/tenants", wrapper.ListTenants)
	router.POST(baseURL+"/tenants", wrapper.CreateTenant)
	router.DELETE(baseURL+"/tenants/:tenantId", wrapper.DeleteTenant)
	router.GET(baseURL+"/tenants/:tenantId", wrapper.GetTenant)
	router.PATCH(baseURL+"/tenants/:tenantId", wrapper.UpdateTenant)
	router.PUT(baseURL+"/tenants/:tenantId", wrapper.PutTenant)

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"2nXOskVp+/fD5cP/BQAA//9RbZ2ux0kAAA==",
}

// GetSwagger returns the content of the embedded swagger specification file
// or error if failed to decode
func decodeSpec() ([]byte, error) {
	zipped, err := base64.StdEncoding.DecodeString(strings.Join(swaggerSpec, ""))
	if err != nil {
		return nil, fmt.Errorf("error base64 decoding spec: %s", err)
	}
	zr, err := gzip.NewReader(bytes.NewReader(zipped))
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(zr)
	if err != nil {
		return nil, fmt.Errorf("error decompressing spec: %s", err)
	}

	return buf.Bytes(), nil
}

var rawSpec = decodeSpecCached()

// a naive cached of a decoded swagger spec
func decodeSpecCached() func() ([]byte, error) {
	data, err := decodeSpec()
	return func() ([]byte, error) {
		return data, err
	}
}

// Constructs a synthetic filesystem for resolving external references when loading openapi specifications.
func PathToRawSpec(pathToFile string) map[string]func() ([]byte, error) {
	var res = make(map[string]func() ([]byte, error))
	if len(pathToFile) > 0 {
		res[pathToFile] = rawSpec
	}

	return res
}

// GetSwagger returns the Swagger specification corresponding to the generated code
// in this file. The external references of Swagger specification are resolved.
// The logic of resolving external references is tightly connected to "import-mapping" feature.
// Externally referenced files must be embedded in the corresponding golang packages.
// Urls can be supported but this task was out of the scope.
func GetSwagger() (swagger *openapi3.T, err error) {
	var resolvePath = PathToRawSpec("")

	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = true
	loader.ReadFromURIFunc = func(loader *openapi3.Loader, url *url.URL) ([]byte, error) {
		var pathToFile = url.String()
		pathToFile = path.Clean(pathToFile)
		getSpec, ok := resolvePath[pathToFile]
		if !ok {
			err1 := fmt.Errorf("path not found: %s", pathToFile)
			return nil, err1
		}
		return getSpec()
	}
	var specData []byte
	specData, err = rawSpec()
	if err != nil {
		return
	}
	swagger, err = loader.LoadFromData(specData)
	if err != nil {
		return
	}
	return
}
