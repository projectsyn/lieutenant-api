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
		Get(APIBasePath+"/clusters").
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
		Get(APIBasePath+"/clusters").
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
}

func TestListClusterWrongToken(t *testing.T) {
	e := setupTest(t)

	result := testutil.NewRequest().
		Get(APIBasePath+"/clusters").
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
		Post(APIBasePath+"/clusters").
		WithJsonBody(newCluster).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusCreated, result.Code())
	cluster := &api.Cluster{}
	err := result.UnmarshalJsonToObject(cluster)
	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.NotEmpty(t, cluster.Id)
	assert.Equal(t, cluster.DisplayName, newCluster.DisplayName)
}

func TestCreateClusterNoJSON(t *testing.T) {
	e := setupTest(t)

	result := testutil.NewRequest().
		Post(APIBasePath+"/clusters/").
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
		Post(APIBasePath+"/clusters/").
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
		Delete(APIBasePath+"/clusters/"+clusterA.Name).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusNoContent, result.Code())
}

func TestClusterGet(t *testing.T) {
	e := setupTest(t)

	result := testutil.NewRequest().
		Get(APIBasePath+"/clusters/"+clusterA.Name).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	cluster := &api.Cluster{}
	err := result.UnmarshalJsonToObject(cluster)
	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Equal(t, api.NewClusterID(clusterA.Name), cluster.ClusterId)
	assert.Equal(t, tenantA.Name, cluster.Tenant)
}

func TestClusterUpdateEmpty(t *testing.T) {
	e := setupTest(t)

	result := testutil.NewRequest().
		Patch(APIBasePath+"/clusters/1").
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
		Patch(APIBasePath+"/clusters/"+clusterA.Name).
		WithJsonBody(updateCluster).
		WithContentType("application/merge-patch+json").
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
		SshDeployKey: pointer.ToString("Some illegal key"),
	}

	result := testutil.NewRequest().
		Patch(APIBasePath+"/clusters/"+clusterB.Name).
		WithJsonBody(updateCluster).
		WithContentType("application/merge-patch+json").
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
		SshDeployKey: pointer.ToString("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPEx4k5NQ46DA+m49Sb3aIyAAqqbz7TdHbArmnnYqwjf"),
	}

	result := testutil.NewRequest().
		Patch(APIBasePath+"/clusters/"+clusterA.Name).
		WithJsonBody(updateCluster).
		WithContentType("application/merge-patch+json").
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
		DisplayName:  pointer.ToString("New Name"),
		SshDeployKey: pointer.ToString("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPEx4k5NQ46DA+m49Sb3aIyAAqqbz7TdHbArmnnYqwjf"),
		GitRepo:      pointer.ToString("https://github.com/some/repo.git"),
		Facts: &api.ClusterFacts{
			"some": "fact",
		},
	}
	result := testutil.NewRequest().
		Patch(APIBasePath+"/clusters/"+clusterB.Name).
		WithJsonBody(updateCluster).
		WithContentType("application/merge-patch+json").
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusNoContent, result.Code())
}
