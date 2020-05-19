package service

import (
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/labstack/echo/v4"
	"github.com/projectsyn/lieutenant-api/pkg/api"
	synv1alpha1 "github.com/projectsyn/lieutenant-operator/pkg/apis/syn/v1alpha1"
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/types"
)

func TestListCluster(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Get("/clusters").
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	clusters := []api.Cluster{}
	err := result.UnmarshalJsonToObject(&clusters)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(clusters), 2)
	assert.Equal(t, clusterA.Spec.DisplayName, *clusters[0].DisplayName)
	assert.Equal(t, clusterB.Spec.DisplayName, *clusters[1].DisplayName)
	assert.Equal(t, string(clusterB.Spec.GitRepoTemplate.RepoType), *clusters[1].GitRepo.Type)
}

func TestListClusterMissingBearer(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Get("/clusters").
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
}

func TestListClusterWrongToken(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Get("/clusters").
		WithHeader(echo.HeaderAuthorization, "asdf").
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
}

func TestCreateCluster(t *testing.T) {
	e, _ := setupTest(t)

	os.Setenv(LieutenantInstanceFactEnvVar, "")
	newCluster := api.CreateCluster{
		ClusterProperties: api.ClusterProperties{
			DisplayName: pointer.ToString("My test cluster"),
			Facts: &api.ClusterFacts{
				"cloud":                "cloudscale",
				"region":               "test",
				LieutenantInstanceFact: "",
			},
		},
		ClusterTenant: api.ClusterTenant{Tenant: tenantA.Name},
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
	assert.Equal(t, newCluster.Facts, cluster.Facts)
	assert.Equal(t, newCluster.Tenant, cluster.Tenant)
}

func TestCreateClusterInstanceFact(t *testing.T) {
	e, _ := setupTest(t)

	instanceName := "lieutenant-dev"
	os.Setenv(LieutenantInstanceFactEnvVar, instanceName)
	newCluster := api.CreateCluster{
		ClusterProperties: api.ClusterProperties{
			DisplayName: pointer.ToString("My test cluster"),
			Facts: &api.ClusterFacts{
				"cloud":  "cloudscale",
				"region": "test",
			},
		},
		ClusterTenant: api.ClusterTenant{Tenant: tenantA.Name},
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
	assert.Equal(t, instanceName, (*cluster.Facts)[LieutenantInstanceFact])

	(*newCluster.Facts)[LieutenantInstanceFact] = "test"
	result = testutil.NewRequest().
		Post("/clusters").
		WithJsonBody(newCluster).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusCreated, result.Code())
	cluster = &api.Cluster{}
	err = result.UnmarshalJsonToObject(cluster)
	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Equal(t, instanceName, (*cluster.Facts)[LieutenantInstanceFact])
}

func TestCreateClusterNoJSON(t *testing.T) {
	e, _ := setupTest(t)

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
	e, _ := setupTest(t)

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
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Delete("/clusters/"+clusterA.Name).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusNoContent, result.Code())
}

func TestClusterGet(t *testing.T) {
	e, _ := setupTest(t)

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
	assert.True(t, strings.HasSuffix(*cluster.InstallURL, clusterA.Status.BootstrapToken.Token))
}

func TestClusterGetNoToken(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Get("/clusters/"+clusterB.Name).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	cluster := &api.Cluster{}
	err := result.UnmarshalJsonToObject(cluster)
	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Equal(t, clusterB.Name, string(cluster.ClusterId.Id))
	assert.Equal(t, tenantB.Name, cluster.Tenant)
	assert.Nil(t, cluster.InstallURL)
}

func TestClusterGetNotFound(t *testing.T) {
	e, _ := setupTest(t)

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
	e, _ := setupTest(t)

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
	e, _ := setupTest(t)

	updateCluster := map[string]string{
		"tenant": "changed-tenant",
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
	e, _ := setupTest(t)

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
	e, _ := setupTest(t)

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
	e, _ := setupTest(t)
	newDisplayName := "New Cluster Name"

	updateCluster := &api.ClusterProperties{
		DisplayName: &newDisplayName,
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
	assert.Equal(t, http.StatusOK, result.Code())
	cluster := &api.Cluster{}
	err := result.UnmarshalJsonToObject(cluster)
	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Equal(t, clusterB.Name, string(cluster.Id))
	assert.Equal(t, newDisplayName, *cluster.DisplayName)
	assert.Equal(t, *updateCluster.GitRepo.DeployKey, *cluster.GitRepo.DeployKey)
}
