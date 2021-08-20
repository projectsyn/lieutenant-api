// Package api provides primitives to interact with the openapi HTTP API.
//
// Code generated by github.com/deepmap/oapi-codegen version v1.7.0 DO NOT EDIT.
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
}

// CreateClusterJSONBody defines parameters for CreateCluster.
type CreateClusterJSONBody Cluster

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

// CreateClusterJSONRequestBody defines body for CreateCluster for application/json ContentType.
type CreateClusterJSONRequestBody CreateClusterJSONBody

// UpdateInventoryJSONRequestBody defines body for UpdateInventory for application/json ContentType.
type UpdateInventoryJSONRequestBody UpdateInventoryJSONBody

// CreateTenantJSONRequestBody defines body for CreateTenant for application/json ContentType.
type CreateTenantJSONRequestBody CreateTenantJSONBody

// ServerInterface represents all server handlers.
type ServerInterface interface {
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
}

// ServerInterfaceWrapper converts echo contexts to parameters.
type ServerInterfaceWrapper struct {
	Handler ServerInterface
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

	router.GET(baseURL+"/clusters", wrapper.ListClusters)
	router.POST(baseURL+"/clusters", wrapper.CreateCluster)
	router.DELETE(baseURL+"/clusters/:clusterId", wrapper.DeleteCluster)
	router.GET(baseURL+"/clusters/:clusterId", wrapper.GetCluster)
	router.PATCH(baseURL+"/clusters/:clusterId", wrapper.UpdateCluster)
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

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"AAAA//+DLo2f7D8AAA==",
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
