package service

import (
	"github.com/labstack/echo/v4"
	"github.com/projectsyn/lieutenant/pkg/api"
	"net/http"
)

// ListClusters lists all clusters
func (s *APIImpl) ListClusters(ctx echo.Context, params api.ListClustersParams) error {
	dispName := "Clustere Name"
	cluster := api.Cluster{}
	cluster.ClusterId = api.ClusterId{Id: api.Id("ClusterID")}
	cluster.DisplayName = &dispName
	if params.Tenant != nil {
		cluster.Tenant = *params.Tenant
	}
	return ctx.JSON(http.StatusOK, []api.Cluster{cluster})
}

// CreateCluster creates a new cluster
func (s *APIImpl) CreateCluster(ctx echo.Context) error {
	return nil
}

// DeleteCluster deletes a cluster
func (s *APIImpl) DeleteCluster(ctx echo.Context, clusterID api.ClusterIdParameter) error {
	return nil
}

// GetCluster gets a cluster
func (s *APIImpl) GetCluster(ctx echo.Context, clusterID api.ClusterIdParameter) error {
	return nil
}

// UpdateCluster updates a cluster
func (s *APIImpl) UpdateCluster(ctx echo.Context, clusterID api.ClusterIdParameter) error {
	return nil
}
