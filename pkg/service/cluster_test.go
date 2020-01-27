package service

import (
	"net/http"
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/projectsyn/lieutenant-api/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestListCluster(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

	result := testutil.NewRequest().
		Get(APIBasePath+"/clusters").
		Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	clusters := &[]api.Cluster{}
	err = result.UnmarshalJsonToObject(clusters)
	assert.NoError(t, err)
	assert.NotNil(t, clusters)
	assert.GreaterOrEqual(t, len(*clusters), 1)
}

func TestCreateCluster(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

	newCluster := api.ClusterProperties{
		Name:        "test-cluster",
		DisplayName: pointer.ToString("My test cluster"),
	}
	result := testutil.NewRequest().
		Post(APIBasePath+"/clusters").
		WithJsonBody(newCluster).
		Go(t, e)
	assert.Equal(t, http.StatusCreated, result.Code())
	cluster := &api.Cluster{}
	err = result.UnmarshalJsonToObject(cluster)
	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.NotEmpty(t, cluster.Id)
	assert.Equal(t, cluster.Name, newCluster.Name)
}

func TestCreateClusterNoJSON(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

	result := testutil.NewRequest().
		Post(APIBasePath+"/clusters/").
		WithJsonContentType().
		WithBody([]byte("invalid-body")).
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err = result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.NotEmpty(t, reason.Reason)
}

func TestCreateClusterEmpty(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

	result := testutil.NewRequest().
		Post(APIBasePath+"/clusters/").
		WithJsonContentType().
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err = result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.NotEmpty(t, reason.Reason)
}

func TestClusterDelete(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

	result := testutil.NewRequest().
		Delete(APIBasePath+"/clusters/1").
		Go(t, e)
	assert.Equal(t, http.StatusNoContent, result.Code())
}

func TestClusterGet(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)
	id := "haevechee2ethot"
	result := testutil.NewRequest().
		Get(APIBasePath+"/clusters/"+id).
		Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	cluster := &api.Cluster{}
	err = result.UnmarshalJsonToObject(cluster)
	assert.NoError(t, err)
	assert.NotNil(t, cluster)
	assert.NotEmpty(t, cluster.Id)
	assert.NotEmpty(t, cluster.Name)
	assert.Equal(t, api.NewClusterID(id), cluster.ClusterId)
}

func TestClusterUpdateEmpty(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

	result := testutil.NewRequest().
		Patch(APIBasePath+"/clusters/1").
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err = result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.NotEmpty(t, reason.Reason)
}

func TestClusterUpdate(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

	updateCluster := &api.ClusterProperties{
		DisplayName: pointer.ToString("New Name"),
	}
	result := testutil.NewRequest().
		Patch(APIBasePath+"/clusters/1/").
		WithJsonBody(updateCluster).
		WithContentType("application/merge-patch+json").
		Go(t, e)
	assert.Equal(t, http.StatusNoContent, result.Code())
}
