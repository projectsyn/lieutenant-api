package service

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/projectsyn/lieutenant-api/pkg/api"
	synv1alpha1 "github.com/projectsyn/lieutenant-operator/pkg/apis/syn/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ListClusters lists all clusters
func (s *APIImpl) ListClusters(c echo.Context, params api.ListClustersParams) error {
	ctx := c.(*APIContext)

	clusterList := &synv1alpha1.ClusterList{}
	if err := ctx.client.List(ctx.context, clusterList, client.InNamespace(s.namespace)); err != nil {
		return err
	}
	clusters := []api.Cluster{}
	for _, cluster := range clusterList.Items {
		apiCluster := api.NewAPIClusterFromCRD(&cluster)
		clusters = append(clusters, *apiCluster)
	}
	return ctx.JSON(http.StatusOK, clusters)
}

// CreateCluster creates a new cluster
func (s *APIImpl) CreateCluster(c echo.Context) error {
	ctx := c.(*APIContext)

	var newCluster *api.CreateClusterJSONRequestBody
	if err := ctx.Bind(&newCluster); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	apiCluster := &api.Cluster{
		ClusterProperties: api.ClusterProperties(*newCluster),
	}
	id, err := api.GenerateClusterID()
	if err != nil {
		return err
	}
	apiCluster.ClusterId = id
	cluster := api.NewCRDFromAPICluster(apiCluster)
	cluster.Namespace = s.namespace
	if err := ctx.client.Create(ctx.context, cluster); err != nil {
		return err
	}
	apiCluster = api.NewAPIClusterFromCRD(cluster)
	return ctx.JSON(http.StatusCreated, apiCluster)
}

// DeleteCluster deletes a cluster
func (s *APIImpl) DeleteCluster(c echo.Context, clusterID api.ClusterIdParameter) error {
	ctx := c.(*APIContext)

	deleteCluster := &synv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      string(clusterID),
			Namespace: s.namespace,
		},
	}

	if err := ctx.client.Delete(ctx.context, deleteCluster); err != nil {
		return err
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GetCluster gets a cluster
func (s *APIImpl) GetCluster(c echo.Context, clusterID api.ClusterIdParameter) error {
	ctx := c.(*APIContext)

	cluster := &synv1alpha1.Cluster{}
	if err := ctx.client.Get(ctx.context, client.ObjectKey{Name: string(clusterID), Namespace: s.namespace}, cluster); err != nil {
		return err
	}
	apiCluster := api.NewAPIClusterFromCRD(cluster)
	return ctx.JSON(http.StatusOK, apiCluster)
}

// UpdateCluster updates a cluster
func (s *APIImpl) UpdateCluster(c echo.Context, clusterID api.ClusterIdParameter) error {
	ctx := c.(*APIContext)

	var patchCluster api.ClusterProperties
	dec := json.NewDecoder(ctx.Request().Body)
	if err := dec.Decode(&patchCluster); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	existingCluster := &synv1alpha1.Cluster{}
	if err := ctx.client.Get(ctx.context, client.ObjectKey{Name: string(clusterID), Namespace: s.namespace}, existingCluster); err != nil {
		return err
	}

	if len(patchCluster.Tenant) > 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Changing tenant of a cluster is not yet implemented")
	}

	if patchCluster.DisplayName != nil {
		existingCluster.Spec.DisplayName = *patchCluster.DisplayName
	}
	if patchCluster.GitRepo != nil {
		if patchCluster.GitRepo.Url != nil {
			existingCluster.Spec.GitRepoURL = *patchCluster.GitRepo.Url
		}
		if patchCluster.GitRepo.HostKeys != nil {
			existingCluster.Spec.GitHostKeys = *patchCluster.GitRepo.HostKeys
		}

		if patchCluster.GitRepo.DeployKey != nil {
			k := strings.Split(*patchCluster.GitRepo.DeployKey, " ")
			if len(k) != 2 {
				return echo.NewHTTPError(http.StatusBadRequest, "Illegal deploy key format. Expected '<type> <public key>'")
			}
			stewardKey := synv1alpha1.DeployKey{
				Type:        k[0],
				Key:         k[1],
				WriteAccess: false,
			}
			if existingCluster.Spec.GitRepoTemplate == nil {
				return echo.NewHTTPError(http.StatusBadRequest, "Cannot update depoy key for not-managed git repo")
			}
			existingCluster.Spec.GitRepoTemplate.DeployKeys["steward"] = stewardKey
		}
	}
	if patchCluster.Facts != nil {
		if existingCluster.Spec.Facts == nil {
			existingCluster.Spec.Facts = &synv1alpha1.Facts{}
		}
		for key, value := range *patchCluster.Facts {
			if valueStr, ok := value.(string); ok {
				(*existingCluster.Spec.Facts)[key] = valueStr
			}
		}
	}
	if err := ctx.client.Update(ctx.context, existingCluster); err != nil {
		return err
	}
	return ctx.NoContent(http.StatusNoContent)
}
