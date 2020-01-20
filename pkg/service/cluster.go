package service

import (
	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	"github.com/projectsyn/lieutenant/pkg/api"
	"net/http"
)

var sampleCluster = api.Cluster{
	ClusterId: api.NewClusterID("haevechee2ethot"),
	ClusterProperties: api.ClusterProperties{
		Name:         "some-cluster",
		Tenant:       "tenant-a",
		DisplayName:  pointer.ToString("Cluster Name"),
		ApiEndpoint:  pointer.ToString("https://api.example.com"),
		GitRepo:      pointer.ToString("ssh://git@github.com/projectsyn/cluster-catalog.git"),
		SshDeployKey: pointer.ToString("ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAAIPEx4k5NQ46DA+m49Sb3aIyAAqqbz7TdHbArmnnYqwjf"),
	},
}

// ListClusters lists all clusters
func (s *APIImpl) ListClusters(ctx echo.Context, params api.ListClustersParams) error {
	return ctx.JSON(http.StatusOK, []api.Cluster{sampleCluster})
}

// CreateCluster creates a new cluster
func (s *APIImpl) CreateCluster(ctx echo.Context) error {
	return ctx.JSON(http.StatusCreated, sampleCluster)
}

// DeleteCluster deletes a cluster
func (s *APIImpl) DeleteCluster(ctx echo.Context, clusterID api.ClusterIdParameter) error {
	return ctx.NoContent(http.StatusNoContent)
}

// GetCluster gets a cluster
func (s *APIImpl) GetCluster(ctx echo.Context, clusterID api.ClusterIdParameter) error {
	c := sampleCluster
	c.Id = api.Id(clusterID)
	return ctx.JSON(http.StatusOK, c)
}

// UpdateCluster updates a cluster
func (s *APIImpl) UpdateCluster(ctx echo.Context, clusterID api.ClusterIdParameter) error {
	return ctx.NoContent(http.StatusNoContent)
}
