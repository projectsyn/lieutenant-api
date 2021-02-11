package service

import (
	"context"
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
	assert.Contains(t, *clusters[0].Annotations, "some")
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
			Annotations: &api.Annotations{
				"new": "annotation",
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
	assert.Equal(t, *newCluster.Annotations, *cluster.Annotations)
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
	assert.Contains(t, reason.Reason, "invalid character")
}

func TestCreateClusterNoTenant(t *testing.T) {
	e, _ := setupTest(t)

	createCluster := map[string]string{
		"displayName": "cluster-name",
	}
	result := testutil.NewRequest().
		Post("/clusters/").
		WithJsonBody(createCluster).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.Contains(t, reason.Reason, "Property 'tenant' is missing")
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
	assert.Contains(t, reason.Reason, "must have a value")
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
	assert.Equal(t, clusterA.Annotations["some"], (*cluster.Annotations)["some"])
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
	assert.Contains(t, reason.Reason, "not found")
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
	assert.Contains(t, reason.Reason, "must have a value")
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
	assert.Contains(t, reason.Reason, "unknown field")
}

func TestClusterUpdateUnknown(t *testing.T) {
	e, _ := setupTest(t)

	updateCluster := map[string]string{
		"displayName": "newName",
		"some":        "field",
		"unknown":     "true",
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
	assert.Contains(t, reason.Reason, "unknown field")
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
	assert.Contains(t, reason.Reason, "Illegal deploy key format")
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
	assert.Contains(t, reason.Reason, "Cannot update depoy key for not-managed")
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
		Annotations: &api.Annotations{
			"existing":   "",
			"additional": "value",
		},
		GlobalGitRepoRevision: pointer.ToString("my-global-revision"),
		TenantGitRepoRevision: pointer.ToString("my-tenant-revision"),
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
	assert.Empty(t, (*cluster.Annotations)["existing"])
	assert.Contains(t, *cluster.Annotations, "additional")
	assert.Len(t, *cluster.Annotations, 2)
	assert.Equal(t, "my-global-revision", pointer.GetString(cluster.GlobalGitRepoRevision))
	assert.Equal(t, "my-tenant-revision", pointer.GetString(cluster.TenantGitRepoRevision))
}

func TestClusterUpdateDisplayName(t *testing.T) {
	e, client := setupTest(t)
	newDisplayName := "New Cluster Name"

	updateCluster := map[string]string{
		"displayName": newDisplayName,
	}
	assert.NotEqual(t, newDisplayName, clusterB.Spec.DisplayName)
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
	assert.Equal(t, newDisplayName, *cluster.DisplayName)
	clusterObj := &synv1alpha1.Cluster{}
	err = client.Get(context.TODO(), types.NamespacedName{
		Namespace: "default",
		Name:      clusterB.Name,
	}, clusterObj)
	assert.NoError(t, err)
	assert.Equal(t, newDisplayName, clusterObj.Spec.DisplayName)
}
