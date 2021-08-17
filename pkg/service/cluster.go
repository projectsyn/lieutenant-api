package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/labstack/echo/v4"
	synv1alpha1 "github.com/projectsyn/lieutenant-operator/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/projectsyn/lieutenant-api/pkg/api"
)

const (
	// LieutenantInstanceFact defines the name of the fact which specifies the Lieutenant instance
	// a cluster was created on
	LieutenantInstanceFact = "lieutenant-instance"

	// LieutenantInstanceFactEnvVar is the env var name that's used to get the instance name
	LieutenantInstanceFactEnvVar = "LIEUTENANT_INSTANCE"
)

// ListClusters lists all clusters
func (s *APIImpl) ListClusters(c echo.Context, _ api.ListClustersParams) error {
	ctx := c.(*APIContext)

	clusterList := &synv1alpha1.ClusterList{}
	err := ctx.client.List(ctx.Request().Context(), clusterList, client.InNamespace(s.namespace))
	if err != nil {
		return err
	}

	clusters := make([]api.Cluster, 0)
	for _, cluster := range clusterList.Items {
		apiCluster := apiClusterWithInstallURL(ctx, &cluster)
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
			apiCluster.Id = api.ClusterIDPrefix + apiCluster.Id
		}
	}

	cluster := api.NewCRDFromAPICluster(apiCluster)
	cluster.Namespace = s.namespace
	if cluster.Spec.Facts == nil {
		cluster.Spec.Facts = synv1alpha1.Facts{}
	}
	cluster.Spec.Facts[LieutenantInstanceFact] = os.Getenv(LieutenantInstanceFactEnvVar)

	if err := ctx.client.Create(ctx.Request().Context(), cluster); err != nil {
		return err
	}
	return ctx.JSON(http.StatusCreated, apiClusterWithInstallURL(ctx, cluster))
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

	if err := ctx.client.Delete(ctx.Request().Context(), deleteCluster); err != nil {
		return err
	}
	return ctx.NoContent(http.StatusNoContent)
}

// GetCluster gets a cluster
func (s *APIImpl) GetCluster(c echo.Context, clusterID api.ClusterIdParameter) error {
	ctx := c.(*APIContext)
	cluster := &synv1alpha1.Cluster{}

	err := ctx.client.Get(ctx.Request().Context(), client.ObjectKey{Name: string(clusterID), Namespace: s.namespace}, cluster)
	if err != nil {
		return err
	}

	return ctx.JSON(http.StatusOK, apiClusterWithInstallURL(ctx, cluster))
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
	if err := ctx.client.Get(ctx.Request().Context(), client.ObjectKey{Name: string(clusterID), Namespace: s.namespace}, existingCluster); err != nil {
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

	if err := ctx.client.Update(ctx.Request().Context(), existingCluster); err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, apiClusterWithInstallURL(ctx, existingCluster))
}

func apiClusterWithInstallURL(ctx *APIContext, cluster *synv1alpha1.Cluster) *api.Cluster {
	apiCluster := api.NewAPIClusterFromCRD(*cluster)

	token, tokenValid := bootstrapToken(cluster)
	if tokenValid {
		installURL := fmt.Sprintf("%s://%s/install/steward.json?token=%s", ctx.Scheme(), ctx.Request().Host, token)
		apiCluster.InstallURL = &installURL
	}

	return apiCluster
}

func bootstrapToken(cluster *synv1alpha1.Cluster) (token string, valid bool) {
	if cluster.Status.BootstrapToken == nil {
		return "", false
	}

	return cluster.Status.BootstrapToken.Token, cluster.Status.BootstrapToken.TokenValid
}
