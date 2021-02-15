package service

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/projectsyn/lieutenant-api/pkg/api"
	synv1alpha1 "github.com/projectsyn/lieutenant-operator/pkg/apis/syn/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	// LieutenantInstanceFact defines the name of the fact which specifies the Lieutenant instance
	// a cluster was created on
	LieutenantInstanceFact = "lieutenant-instance"

	// LieutenantInstanceFactEnvVar is the env var name that's used to get the instance name
	LieutenantInstanceFactEnvVar = "LIEUTENANT_INSTANCE"
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
		apiCluster := api.NewAPIClusterFromCRD(cluster)
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
	apiCluster := api.Cluster(*newCluster)
	if !strings.HasPrefix(string(apiCluster.Id), api.ClusterIDPrefix) {
		if apiCluster.Id == "" {
			id, err := api.GenerateClusterID()
			if err != nil {
				return err
			}
			apiCluster.ClusterId = id
		} else {
			apiCluster.Id = api.ClusterIDPrefix+apiCluster.Id
		}
	}
	cluster := api.NewCRDFromAPICluster(apiCluster)
	cluster.Namespace = s.namespace
	if cluster.Spec.Facts == nil {
		cluster.Spec.Facts = &synv1alpha1.Facts{}
	}
	(*cluster.Spec.Facts)[LieutenantInstanceFact] = os.Getenv(LieutenantInstanceFactEnvVar)

	if err := ctx.client.Create(ctx.context, cluster); err != nil {
		return err
	}
	apiCluster = *api.NewAPIClusterFromCRD(*cluster)
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
	apiCluster := api.NewAPIClusterFromCRD(*cluster)
	if cluster.Status.BootstrapToken != nil &&
		cluster.Status.BootstrapToken.TokenValid {
		installURL := ctx.Scheme() + "://" + ctx.Request().Host + "/install/steward.json?token=" + cluster.Status.BootstrapToken.Token
		apiCluster.InstallURL = &installURL
	}
	return ctx.JSON(http.StatusOK, apiCluster)
}

// UpdateCluster updates a cluster
func (s *APIImpl) UpdateCluster(c echo.Context, clusterID api.ClusterIdParameter) error {
	ctx := c.(*APIContext)

	var patchCluster api.ClusterProperties
	dec := json.NewDecoder(ctx.Request().Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(&patchCluster); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	existingCluster := &synv1alpha1.Cluster{}
	if err := ctx.client.Get(ctx.context, client.ObjectKey{Name: string(clusterID), Namespace: s.namespace}, existingCluster); err != nil {
		return err
	}

	if patchCluster.GitRepo != nil && patchCluster.GitRepo.DeployKey != nil {
		k := strings.Split(*patchCluster.GitRepo.DeployKey, " ")
		if len(k) != 2 {
			return echo.NewHTTPError(http.StatusBadRequest, "Illegal deploy key format. Expected '<type> <public key>'")
		}
		if existingCluster.Spec.GitRepoTemplate == nil {
			return echo.NewHTTPError(http.StatusBadRequest, "Cannot update deploy key for unmanaged git repo")
		}
	}

	api.SyncCRDFromAPICluster(patchCluster, existingCluster)

	if err := ctx.client.Update(ctx.context, existingCluster); err != nil {
		return err
	}
	apiCluster := api.NewAPIClusterFromCRD(*existingCluster)
	return ctx.JSON(http.StatusOK, apiCluster)
}
