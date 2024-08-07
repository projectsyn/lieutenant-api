package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	synv1alpha1 "github.com/projectsyn/lieutenant-operator/api/v1alpha1"
	"go.uber.org/multierr"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
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
func (s *APIImpl) ListClusters(c echo.Context, p api.ListClustersParams) error {
	ctx := c.(*APIContext)

	filterOptions := []client.ListOption{client.InNamespace(s.namespace)}
	if p.Tenant != nil && *p.Tenant != "" {
		filterOptions = append(filterOptions, client.MatchingLabels{synv1alpha1.LabelNameTenant: *p.Tenant})
	}

	clusterList := &synv1alpha1.ClusterList{}
	err := ctx.client.List(ctx.Request().Context(), clusterList, filterOptions...)
	if err != nil {
		return err
	}

	clusters := make([]api.Cluster, 0, len(clusterList.Items))
	errs := make([]error, 0, len(clusterList.Items))
	for _, cluster := range clusterList.Items {
		apiCluster, err := apiClusterWithInstallURL(ctx, &cluster)
		clusters = append(clusters, *apiCluster)
		errs = append(errs, err)
	}
	if err := multierr.Combine(errs...); err != nil {
		return fmt.Errorf("failed to translate CRD to API representation: %w", err)
	}
	sortClustersBy(clusters, p.SortBy)
	return ctx.JSON(http.StatusOK, clusters)
}

func sortClustersBy(clusters []api.Cluster, by *api.ListClustersParamsSortBy) {
	sortBy := "id"
	if by != nil {
		sortBy = string(*by)
	}
	sort.Slice(clusters, func(i, j int) bool {
		switch sortBy {
		case "tenant":
			return clusters[i].Tenant < clusters[j].Tenant
		case "displayName":
			di := ""
			if clusters[i].DisplayName != nil {
				di = *clusters[i].DisplayName
			}
			dj := ""
			if clusters[j].DisplayName != nil {
				dj = *clusters[j].DisplayName
			}

			return di < dj
		default:
			return clusters[i].Id.String() < clusters[j].Id.String()
		}
	})
}

// CreateCluster creates a new cluster
func (s *APIImpl) CreateCluster(c echo.Context) error {
	ctx := c.(*APIContext)

	var newCluster *api.CreateClusterJSONRequestBody
	if err := ctx.Bind(&newCluster); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	apiCluster := api.Cluster(*newCluster)

	cluster, err := api.NewCRDFromAPICluster(apiCluster)
	if err != nil {
		return err
	}

	return s.createCluster(ctx, cluster)
}

func (s *APIImpl) createCluster(ctx *APIContext, cluster *synv1alpha1.Cluster) error {
	cluster.Namespace = s.namespace
	if cluster.Spec.Facts == nil {
		cluster.Spec.Facts = synv1alpha1.Facts{}
	}
	cluster.Spec.Facts[LieutenantInstanceFact] = os.Getenv(LieutenantInstanceFactEnvVar)

	// Need to copy status as Create will modify it
	status := cluster.Status.DeepCopy()
	if err := ctx.client.Create(ctx.Request().Context(), cluster); err != nil {
		return err
	}
	cluster.Status = *status
	if err := ctx.client.Status().Update(ctx.Request().Context(), cluster); err != nil {
		return err
	}
	ac, err := apiClusterWithInstallURL(ctx, cluster)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusCreated, ac)
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

	ac, err := apiClusterWithInstallURL(ctx, cluster)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, ac)
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

	err := api.SyncCRDFromAPICluster(patchCluster, existingCluster)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return s.updateCluster(ctx, existingCluster)
}

func (s *APIImpl) updateCluster(ctx *APIContext, existingCluster *synv1alpha1.Cluster) error {
	// Need to copy status as the update will modify it
	status := existingCluster.Status.DeepCopy()
	if err := ctx.client.Update(ctx.Request().Context(), existingCluster); err != nil {
		return err
	}
	existingCluster.Status = *status
	if err := ctx.client.Status().Update(ctx.Request().Context(), existingCluster); err != nil {
		return err
	}

	ac, err := apiClusterWithInstallURL(ctx, existingCluster)
	if err != nil {
		return err
	}
	return ctx.JSON(http.StatusOK, ac)
}

// PutCluster updates the cluster or cleates it if it does not exist
func (s *APIImpl) PutCluster(c echo.Context, clusterID api.ClusterIdParameter) error {
	ctx := c.(*APIContext)

	body := &api.PutClusterJSONRequestBody{}
	if err := ctx.Bind(body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}
	apiCluster := api.Cluster(*body)
	apiCluster.Id = pointer.To(api.Id(clusterID))

	cluster, err := api.NewCRDFromAPICluster(apiCluster)
	if err != nil {
		return err
	}

	found := &synv1alpha1.Cluster{}
	err = ctx.client.Get(ctx.Request().Context(), client.ObjectKey{Name: string(clusterID), Namespace: s.namespace}, found)
	if err != nil && !errors.IsNotFound(err) {
		return err
	}

	if errors.IsNotFound(err) {
		return s.createCluster(ctx, cluster)
	}

	found.Spec = cluster.Spec
	found.Annotations = cluster.Annotations
	return s.updateCluster(ctx, found)
}

// PostClusterCompileMeta compiles the meta data of a cluster
func (s *APIImpl) PostClusterCompileMeta(c echo.Context, clusterID api.ClusterIdParameter) error {
	ctx := c.(*APIContext)

	body := &api.ClusterCompileMeta{}
	if err := ctx.Bind(body); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err)
	}

	toPatch := &synv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      string(clusterID),
			Namespace: s.namespace,
		},
	}
	patch := []map[string]any{
		{"op": "replace", "path": "/status/compileMeta", "value": body},
	}
	rawPatch, err := json.Marshal(patch)
	if err != nil {
		return fmt.Errorf("failed to marshal patch: %w", err)
	}

	if err := retryOnConflict(func() error {
		return ctx.client.Status().Patch(ctx.Request().Context(), toPatch, client.RawPatch(types.JSONPatchType, rawPatch))
	}); err != nil {
		return err
	}

	return ctx.NoContent(http.StatusNoContent)
}

func apiClusterWithInstallURL(ctx *APIContext, cluster *synv1alpha1.Cluster) (*api.Cluster, error) {
	apiCluster, err := api.NewAPIClusterFromCRD(*cluster)
	if err != nil {
		return nil, err
	}

	token, tokenValid := bootstrapToken(cluster)
	if tokenValid {
		installURL := fmt.Sprintf("%s://%s/install/steward.json?token=%s", ctx.Scheme(), ctx.Request().Host, token)
		apiCluster.InstallURL = &installURL
	}

	return apiCluster, nil
}

func bootstrapToken(cluster *synv1alpha1.Cluster) (token string, valid bool) {
	if cluster.Status.BootstrapToken == nil {
		return "", false
	}

	return cluster.Status.BootstrapToken.Token, cluster.Status.BootstrapToken.TokenValid
}

// retryOnConflict retries the given function if the returned error is a conflict (HTTP 409) error once
func retryOnConflict(f func() error) error {
	err := f()
	if errors.IsConflict(err) {
		return f()
	}
	return err
}
