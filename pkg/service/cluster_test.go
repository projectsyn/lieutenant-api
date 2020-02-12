package service

import (
	"net/http"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/labstack/echo/v4"
	"github.com/projectsyn/lieutenant-api/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestListCluster(t *testing.T) {
	e := setupTest(t)

	result := testutil.NewRequest().
		Get("/clusters").
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	clusters := []api.Cluster{}
	err := result.UnmarshalJsonToObject(&clusters)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(clusters), 1)
	found := false
	for _, cluster := range clusters {
		if string(cluster.ClusterId.Id) == clusterA.Name {
			found = true
			break
		}
	}
	assert.Truef(t, found, "Cluster not found in result list", clusterA.Name)
}

func TestListClusterMissingBearer(t *testing.T) {
	e := setupTest(t)

	result := testutil.NewRequest().
		Get("/clusters").
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
}

func TestListClusterWrongToken(t *testing.T) {
	e := setupTest(t)

	result := testutil.NewRequest().
		Get("/clusters").
		WithHeader(echo.HeaderAuthorization, "asdf").
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
}

func TestCreateCluster(t *testing.T) {
	e := setupTest(t)

	newCluster := api.ClusterProperties{
		DisplayName: pointer.ToString("My test cluster"),
		Tenant:      tenantA.Name,
	}
	result := testutil.NewRequest().
		Post("/clusters").
		WithJsonBody(newCluster).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusCreated, result.Code())
	cluster := &api.Cluster{}
	err := result.UnmarshalJsonToObject(cluster)
	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Contains(t, cluster.Id, api.ClusterIDPrefix)
	assert.Equal(t, cluster.DisplayName, newCluster.DisplayName)
}

func TestCreateClusterNoJSON(t *testing.T) {
	e := setupTest(t)

	result := testutil.NewRequest().
		Post("/clusters/").
		WithJsonContentType().
		WithBody([]byte("invalid-body")).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.NotEmpty(t, reason.Reason)
}

func TestCreateClusterEmpty(t *testing.T) {
	e := setupTest(t)

	result := testutil.NewRequest().
		Post("/clusters/").
		WithJsonContentType().
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.NotEmpty(t, reason.Reason)
}

func TestClusterDelete(t *testing.T) {
	e := setupTest(t)

	result := testutil.NewRequest().
		Delete("/clusters/"+clusterA.Name).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusNoContent, result.Code())
}

func TestClusterGet(t *testing.T) {
	e := setupTest(t)

	result := testutil.NewRequest().
		Get("/clusters/"+clusterA.Name).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	cluster := &api.Cluster{}
	err := result.UnmarshalJsonToObject(cluster)
	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Equal(t, clusterA.Name, string(cluster.ClusterId.Id))
	assert.Equal(t, tenantA.Name, cluster.Tenant)
	assert.Equal(t, clusterA.Spec.GitHostKeys, *cluster.GitRepo.HostKeys)
}

func TestClusterGetNotFound(t *testing.T) {
	e := setupTest(t)

	result := testutil.NewRequest().
		Get("/clusters/not-existing").
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusNotFound, result.Code())
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.NotEmpty(t, reason.Reason)
}

func TestClusterUpdateEmpty(t *testing.T) {
	e := setupTest(t)

	result := testutil.NewRequest().
		Patch("/clusters/1").
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.NotEmpty(t, reason.Reason)
}

func TestClusterUpdateTenant(t *testing.T) {
	e := setupTest(t)

	updateCluster := &api.ClusterProperties{
		Tenant: "changed-tenant",
	}

	result := testutil.NewRequest().
		Patch("/clusters/"+clusterA.Name).
		WithJsonBody(updateCluster).
		WithContentType(api.ContentJSONPatch).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.NotEmpty(t, reason.Reason)
}

func TestClusterUpdateIllegalDeployKey(t *testing.T) {
	e := setupTest(t)

	updateCluster := &api.ClusterProperties{
		GitRepo: &api.GitRepo{
			DeployKey: pointer.ToString("Some illegal key"),
		},
	}

	result := testutil.NewRequest().
		Patch("/clusters/"+clusterB.Name).
		WithJsonBody(updateCluster).
		WithContentType(api.ContentJSONPatch).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.NotEmpty(t, reason.Reason)
}

func TestClusterUpdateNotManagedDeployKey(t *testing.T) {
	e := setupTest(t)

	updateCluster := &api.ClusterProperties{
		GitRepo: &api.GitRepo{
			DeployKey: pointer.ToString("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPEx4k5NQ46DA+m49Sb3aIyAAqqbz7TdHbArmnnYqwjf"),
		},
	}

	result := testutil.NewRequest().
		Patch("/clusters/"+clusterA.Name).
		WithJsonBody(updateCluster).
		WithContentType(api.ContentJSONPatch).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.NotEmpty(t, reason.Reason)
}

func TestClusterUpdate(t *testing.T) {
	e := setupTest(t)

	updateCluster := &api.ClusterProperties{
		DisplayName: pointer.ToString("New Name"),
		GitRepo: &api.GitRepo{
			DeployKey: pointer.ToString("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPEx4k5NQ46DA+m49Sb3aIyAAqqbz7TdHbArmnnYqwjf"),
			Url:       pointer.ToString("https://github.com/some/repo.git"),
		},
		Facts: &api.ClusterFacts{
			"some": "fact",
		},
	}
	result := testutil.NewRequest().
		Patch("/clusters/"+clusterB.Name).
		WithJsonBody(updateCluster).
		WithContentType(api.ContentJSONPatch).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusNoContent, result.Code())
}
