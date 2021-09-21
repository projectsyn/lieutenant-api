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

// CreateClusterJSONRequestBody defines body for CreateCluster for application/json ContentType.
type CreateClusterJSONRequestBody CreateClusterJSONBody

// PutClusterJSONRequestBody defines body for PutCluster for application/json ContentType.
type PutClusterJSONRequestBody PutClusterJSONBody

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

}

// Base64 encoded, gzipped, json marshaled Swagger object
var swaggerSpec = []string{

	"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",
	"AAA=",
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
