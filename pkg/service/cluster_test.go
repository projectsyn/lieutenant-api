package service

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	"github.com/oapi-codegen/testutil"
	synv1alpha1 "github.com/projectsyn/lieutenant-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/projectsyn/lieutenant-api/pkg/api"
)

func TestListCluster(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Get("/clusters").
		WithHeader(echo.HeaderAuthorization, bearerToken).
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusOK, result)
	clusters := make([]api.Cluster, 0)
	err := result.UnmarshalJsonToObject(&clusters)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(clusters), 2)
	assert.Equal(t, clusterA.Spec.DisplayName, *clusters[0].DisplayName)
	assert.Equal(t, clusterB.Spec.DisplayName, *clusters[1].DisplayName)
	assert.Equal(t, string(clusterB.Spec.GitRepoTemplate.RepoType), *clusters[1].GitRepo.Type)
	assert.NotNil(t, clusters[0].InstallURL)
	assert.Nil(t, clusters[1].InstallURL)
	assert.True(t, strings.HasSuffix(*clusters[0].InstallURL, clusterA.Status.BootstrapToken.Token))
	assert.Contains(t, *clusters[0].Annotations, "some")
}

func TestListCluster_FilteredByTenant(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Get("/clusters?tenant="+tenantA.Name).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusOK, result)
	clusters := make([]api.Cluster, 0)
	err := result.UnmarshalJsonToObject(&clusters)
	assert.NoError(t, err)
	assert.Len(t, clusters, 1)
	assert.Equal(t, clusterA.Spec.DisplayName, *clusters[0].DisplayName)
}
func TestListCluster_Sort(t *testing.T) {

	clusterC := clusterA.DeepCopy()
	clusterC.Name = "sample-cluster-c"
	clusterC.Spec.DisplayName = "Z Cluster c"
	clusterC.Spec.TenantRef.Name = "c-tenant"

	tcs := map[string]struct {
		sortBy string
		order  []string
	}{
		"sort_by id": {
			sortBy: "id",
			order: []string{
				clusterA.Name,
				clusterB.Name,
				clusterC.Name,
			},
		},
		"sort_by tenant": {
			sortBy: "tenant",
			order: []string{
				clusterC.Name,
				clusterA.Name,
				clusterB.Name,
			},
		},
		"sort_by displayName": {
			sortBy: "displayName",
			order: []string{
				clusterB.Name,
				clusterA.Name,
				clusterC.Name,
			},
		},
	}

	e, _ := setupTest(t, clusterC)
	for name, tc := range tcs {
		t.Run(name, func(t *testing.T) {
			result := testutil.NewRequest().
				Get(fmt.Sprintf("/clusters?sort_by=%s", tc.sortBy)).
				WithHeader(echo.HeaderAuthorization, bearerToken).
				GoWithHTTPHandler(t, e)
			requireHTTPCode(t, http.StatusOK, result)
			clusters := make([]api.Cluster, 0)
			err := result.UnmarshalJsonToObject(&clusters)
			assert.NoError(t, err)
			assert.Len(t, clusters, 3)
			for i := range tc.order {
				assert.Equal(t, tc.order[i], clusters[i].Id.String())
			}
		})
	}
}

func TestListClusterMissingBearer(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Get("/clusters").
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusBadRequest, result)
}

func TestListClusterWrongToken(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Get("/clusters").
		WithHeader(echo.HeaderAuthorization, "asdf").
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusBadRequest, result)
}

func TestCreateCluster(t *testing.T) {
	e, _ := setupTest(t)

	err := os.Setenv(LieutenantInstanceFactEnvVar, "")
	require.NoError(t, err)

	newCluster := api.Cluster{
		ClusterProperties: api.ClusterProperties{
			DisplayName: pointer.ToString("My test cluster"),
			Facts: &api.ClusterFacts{
				"cloud":                "cloudscale",
				"region":               "test",
				LieutenantInstanceFact: "",
			},
			DynamicFacts: &api.DynamicClusterFacts{
				"kubernetesVersion": "1.16",
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
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusCreated, result)

	cluster := &api.Cluster{}
	err = result.UnmarshalJsonToObject(cluster)

	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Contains(t, cluster.Id.String(), api.ClusterIDPrefix)
	assert.Equal(t, cluster.DisplayName, newCluster.DisplayName)
	assert.Equal(t, newCluster.Facts, cluster.Facts)
	assert.Equal(t, newCluster.DynamicFacts, cluster.DynamicFacts)
	assert.Equal(t, newCluster.Tenant, cluster.Tenant)
	assert.Equal(t, *newCluster.Annotations, *cluster.Annotations)
}

var createClusterWithIDTests = map[string]struct {
	request  api.Id
	response api.Id
}{
	"requested ID gets accepted": {
		request:  "c-my-custom-id",
		response: "c-my-custom-id",
	},
	"ID without prefix gets prefixed": {
		request:  "my-custom-id",
		response: "c-my-custom-id",
	},
}

func TestCreateClusterWithId(t *testing.T) {
	for name, tt := range createClusterWithIDTests {
		t.Run(name, func(t *testing.T) {
			e, _ := setupTest(t)

			request := api.Cluster{
				ClusterId: api.ClusterId{
					Id: pointer.To(tt.request),
				},
				ClusterProperties: api.ClusterProperties{
					DisplayName: pointer.ToString("My test cluster"),
				},
				ClusterTenant: api.ClusterTenant{Tenant: tenantA.Name},
			}
			result := testutil.NewRequest().
				Post("/clusters").
				WithJsonBody(request).
				WithHeader(echo.HeaderAuthorization, bearerToken).
				GoWithHTTPHandler(t, e)
			requireHTTPCode(t, http.StatusCreated, result)
			cluster := &api.Cluster{}
			err := result.UnmarshalJsonToObject(cluster)
			assert.NoError(t, err)
			assert.Equal(t, tt.response.String(), cluster.Id.String())
		})
	}
}

func TestCreateClusterInstanceFact(t *testing.T) {
	e, _ := setupTest(t)

	instanceName := "lieutenant-dev"
	err := os.Setenv(LieutenantInstanceFactEnvVar, instanceName)
	require.NoError(t, err)

	newCluster := api.Cluster{
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
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusCreated, result)

	cluster := &api.Cluster{}
	err = result.UnmarshalJsonToObject(cluster)
	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Equal(t, instanceName, (*cluster.Facts)[LieutenantInstanceFact])

	(*newCluster.Facts)[LieutenantInstanceFact] = "test"
	result = testutil.NewRequest().
		Post("/clusters").
		WithJsonBody(newCluster).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusCreated, result)
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
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusBadRequest, result)
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.Contains(t, reason.Reason, "invalid character")
}

func TestCreateClusterNoTenant(t *testing.T) {
	e, _ := setupTest(t)

	createCluster := map[string]string{
		"id":          "c-1234",
		"displayName": "cluster-name",
	}
	result := testutil.NewRequest().
		Post("/clusters/").
		WithJsonBody(createCluster).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusBadRequest, result)
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.Contains(t, reason.Reason, "property \"tenant\" is missing")
}

func TestCreateClusterEmpty(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Post("/clusters/").
		WithJsonContentType().
		WithHeader(echo.HeaderAuthorization, bearerToken).
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusBadRequest, result)
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.Contains(t, reason.Reason, "value is required but missing")
}

func TestClusterDelete(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Delete("/clusters/"+clusterA.Name).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusNoContent, result)
}

func TestClusterGet(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Get("/clusters/"+clusterA.Name).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusOK, result)
	cluster := &api.Cluster{}
	err := result.UnmarshalJsonToObject(cluster)
	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Equal(t, clusterA.Name, cluster.Id.String())
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
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusOK, result)
	cluster := &api.Cluster{}
	err := result.UnmarshalJsonToObject(cluster)
	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Equal(t, clusterB.Name, cluster.Id.String())
	assert.Equal(t, tenantB.Name, cluster.Tenant)
	assert.Nil(t, cluster.InstallURL)
}

func TestClusterGetNotFound(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Get("/clusters/not-existing").
		WithHeader(echo.HeaderAuthorization, bearerToken).
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusNotFound, result)
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
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusBadRequest, result)
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.Contains(t, reason.Reason, "value is required but missing")
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
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusBadRequest, result)
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
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusBadRequest, result)
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
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusBadRequest, result)
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
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusBadRequest, result)
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.Contains(t, reason.Reason, "Cannot set deploy key for unmanaged git repo")
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
		DynamicFacts: &api.DynamicClusterFacts{
			"dynamic": "fact",
			"complex": struct{ name string }{name: "fact"},
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
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusOK, result)
	cluster := &api.Cluster{}
	err := result.UnmarshalJsonToObject(cluster)
	require.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.Equal(t, clusterB.Name, cluster.Id.String())
	assert.Equal(t, newDisplayName, *cluster.DisplayName)
	assert.Equal(t, *updateCluster.GitRepo.DeployKey, *cluster.GitRepo.DeployKey)
	assert.Empty(t, (*cluster.Annotations)["existing"])
	assert.Contains(t, *cluster.Annotations, "additional")
	assert.Len(t, *cluster.Annotations, 2)
	assert.Equal(t, "my-global-revision", pointer.GetString(cluster.GlobalGitRepoRevision))
	assert.Equal(t, "my-tenant-revision", pointer.GetString(cluster.TenantGitRepoRevision))

	require.NotNil(t, cluster.DynamicFacts)
	assert.Contains(t, *cluster.DynamicFacts, "dynamic")
	assert.Contains(t, *cluster.DynamicFacts, "complex")
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
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusOK, result)
	cluster := &api.Cluster{}
	err := result.UnmarshalJsonToObject(cluster)
	require.NoError(t, err)
	assert.Equal(t, newDisplayName, *cluster.DisplayName)
	clusterObj := &synv1alpha1.Cluster{}
	err = client.Get(context.TODO(), types.NamespacedName{
		Namespace: "default",
		Name:      clusterB.Name,
	}, clusterObj)
	assert.NoError(t, err)
	assert.Equal(t, newDisplayName, clusterObj.Spec.DisplayName)
}

var putClusterTestCases = map[string]struct {
	cluster *api.Cluster
	code    int
	valid   func(t *testing.T, act *api.Cluster) bool
}{
	"put unchanged object": {
		cluster: api.NewAPIClusterFromCRD(*clusterB),
		code:    http.StatusOK,
		valid: func(t *testing.T, act *api.Cluster) bool {
			return true
		},
	},
	"put updated object": {
		cluster: func() *api.Cluster {
			cluster := api.NewAPIClusterFromCRD(*clusterB)
			(*cluster.Facts)["foo"] = "bar"
			return cluster
		}(),
		code: http.StatusOK,
		valid: func(t *testing.T, act *api.Cluster) bool {
			require.Contains(t, *act.Facts, "cloud")
			assert.Equal(t, clusterB.Spec.Facts["cloud"], (*act.Facts)["cloud"])
			require.Contains(t, *act.Facts, "foo")
			assert.Equal(t, (*act.Facts)["foo"], "bar")
			return true
		},
	},
	"put new object": {
		cluster: &api.Cluster{
			ClusterId: api.ClusterId{
				Id: pointer.To(api.Id("c-new-2379")),
			},
			ClusterProperties: api.ClusterProperties{
				DisplayName: pointer.ToString("My new cluster"),
				Facts: &api.ClusterFacts{
					"cloud":                "cloudscale",
					"region":               "test",
					LieutenantInstanceFact: "",
				},
				DynamicFacts: &api.DynamicClusterFacts{
					"kubernetesVersion": "1.16",
				},
				Annotations: &api.Annotations{
					"new": "annotation",
				},
			},
			ClusterTenant: api.ClusterTenant{Tenant: tenantA.Name},
		},
		code: http.StatusCreated,
		valid: func(t *testing.T, act *api.Cluster) bool {
			assert.Contains(t, act.Id.String(), api.ClusterIDPrefix)
			assert.Equal(t, pointer.ToString("My new cluster"), act.DisplayName)
			return true
		},
	},
}

func TestClusterPut(t *testing.T) {
	e, client := setupTest(t)

	for k, tc := range putClusterTestCases {
		t.Run(k, func(t *testing.T) {
			result := testutil.NewRequest().
				Put("/clusters/"+tc.cluster.Id.String()).
				WithJsonBody(tc.cluster).
				WithHeader(echo.HeaderAuthorization, bearerToken).
				GoWithHTTPHandler(t, e)
			requireHTTPCode(t, tc.code, result)

			res := &api.Cluster{}
			err := result.UnmarshalJsonToObject(res)
			require.NoError(t, err)
			assert.True(t, tc.valid(t, res))

			clusterObj := &synv1alpha1.Cluster{}
			err = client.Get(context.TODO(), types.NamespacedName{
				Namespace: "default",
				Name:      res.Id.String(),
			}, clusterObj)
			require.NotNil(t, clusterObj)
			require.NotEmpty(t, clusterObj.Name)
			assert.True(t, tc.valid(t, api.NewAPIClusterFromCRD(*clusterObj)))
		})
	}

}

func TestClusterPutCreateNameMissmatch(t *testing.T) {
	e, client := setupTest(t)
	cluster := &api.Cluster{
		ClusterId: api.ClusterId{
			Id: pointer.To(api.Id("c-new-2379")),
		},
		ClusterProperties: api.ClusterProperties{
			DisplayName: pointer.ToString("My new cluster"),
		},
		ClusterTenant: api.ClusterTenant{Tenant: tenantA.Name},
	}
	result := testutil.NewRequest().
		Put("/clusters/c-foo").
		WithJsonBody(cluster).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusCreated, result)

	res := &api.Cluster{}
	err := result.UnmarshalJsonToObject(res)
	require.NoError(t, err)
	require.Equal(t, "c-foo", res.Id.String())
	require.NotEmpty(t, res.Facts)

	clusterObj := &synv1alpha1.Cluster{}
	err = client.Get(context.TODO(), types.NamespacedName{
		Namespace: "default",
		Name:      res.Id.String(),
	}, clusterObj)
	require.NotNil(t, clusterObj)
}

func TestClusterPostCompileMeta(t *testing.T) {
	e, c := setupTest(t)
	cluster := &synv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "c-compile-meta",
			Namespace: "default",
		},
	}
	require.NoError(t, c.Create(context.Background(), cluster))

	compileOutput := map[string]any{
		"commodoreBuildInfo": map[string]any{
			"version": "1.0.0",
		},
		"lastCompile": time.Date(2024, time.April, 14, 21, 5, 56, 0, time.UTC).Format(time.RFC3339),
		"global": map[string]any{
			"gitSha":  "68e5722a883f3044e287afe810ded53023175a32",
			"url":     "example.com/global.git",
			"version": "master",
		},
		"tenant": map[string]any{
			"gitSha":  "c12b5847133adc2a62e484bfa5da34e1c09d4baf",
			"url":     "example.com/tenant.git",
			"version": "master",
		},
		"instances": map[string]any{
			"operations-operator-dev": map[string]any{
				"component": "operations-operator",
				"gitSha":    "cb0b6e77e8a213c614716155efc2de929a200ec0",
				"url":       "example.com/operations-operator.git",
				"version":   "v0.1.0",
			},
		},
		"packages": map[string]any{
			"app1": map[string]any{
				"gitSha":  "3ab3bf74860045601645a37c170dfe04fe7eddd8",
				"url":     "example.com/app1.git",
				"version": "develop",
				"path":    "packages/main",
			},
		},
	}

	result := testutil.NewRequest().
		Post("/"+path.Join("clusters", cluster.Name, "compileMeta")).
		WithJsonBody(compileOutput).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusNoContent, result)

	require.NoError(t, c.Get(context.Background(), client.ObjectKeyFromObject(cluster), cluster))
	requireJSONMatch(t, compileOutput, cluster.Status.CompileMeta)
}

func TestClusterPostCompileMeta_OverridesExisting_NoMerge(t *testing.T) {
	e, c := setupTest(t)
	cluster := &synv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "c-compile-meta",
			Namespace: "default",
		},
		Status: synv1alpha1.ClusterStatus{
			CompileMeta: synv1alpha1.CompileMeta{
				CommodoreBuildInfo: map[string]string{
					"version":  "6.6.6",
					"otherkey": "othervalue",
				},
				Instances: map[string]synv1alpha1.CompileMetaInstanceVersionInfo{
					"operations-operator-dev": {
						Component: "operations-operator",
					},
				},
			},
		},
	}
	require.NoError(t, c.Create(context.Background(), cluster))

	compileOutput := map[string]any{
		"commodoreBuildInfo": map[string]any{
			"version": "7.0.0",
		},
		"instances": map[string]any{
			"operations-operator-prod": map[string]any{
				"component": "operations-operator",
			},
		},
	}

	result := testutil.NewRequest().
		Post("/"+path.Join("clusters", cluster.Name, "compileMeta")).
		WithJsonBody(compileOutput).
		WithHeader(echo.HeaderAuthorization, bearerToken).
		GoWithHTTPHandler(t, e)
	requireHTTPCode(t, http.StatusNoContent, result)

	require.NoError(t, c.Get(context.Background(), client.ObjectKeyFromObject(cluster), cluster))
	require.NotContains(t, cluster.Status.CompileMeta.CommodoreBuildInfo, "otherkey")
	require.NotContains(t, cluster.Status.CompileMeta.Instances, "operations-operator-dev")
	require.Contains(t, cluster.Status.CompileMeta.Instances, "operations-operator-prod")
}

// requireJSONMatch checks if the JSON representation of two objects are equal.
func requireJSONMatch(t *testing.T, expected, actual any) {
	t.Helper()
	expectedJSON, err := json.Marshal(expected)
	require.NoError(t, err)
	actualJSON, err := json.Marshal(actual)
	require.NoError(t, err)
	require.JSONEq(t, string(expectedJSON), string(actualJSON))
}

// requireHTTPCode is a helper function to check the HTTP status code of a response and log the response body if the code is not as expected.
func requireHTTPCode(t *testing.T, expected int, result *testutil.CompletedRequest) {
	t.Helper()
	require.Equalf(t, expected, result.Code(), "Unexpected response code: %d, body: %s", result.Code(), string(result.Recorder.Body.String()))
}
