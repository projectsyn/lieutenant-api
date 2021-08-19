package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/taion809/haikunator"

	synv1alpha1 "github.com/projectsyn/lieutenant-operator/api/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ClusterIDPrefix is prefixed to all cluster IDs
	ClusterIDPrefix = "c-"
	// TenantIDPrefix is prefixed to all tenant IDs
	TenantIDPrefix = "t-"
	// ContentJSONPatch is the content type to do JSON updates
	ContentJSONPatch = "application/merge-patch+json"
)

var (
	h = haikunator.NewHaikunator()
)

// GenerateClusterID creates a new cluster id
func GenerateClusterID() (ClusterId, error) {
	id, err := generateID(ClusterIDPrefix)
	return ClusterId{
		Id: id,
	}, err
}

// GenerateTenantID creates a new tenant id
func GenerateTenantID() (TenantId, error) {
	id, err := generateID(TenantIDPrefix)
	return TenantId{
		Id: id,
	}, err
}

// GenerateID generates a new id from random alphanumeric characters
func generateID(prefix string) (Id, error) {
	retry := 10
	var id = ""
	var i int
	// let's try a few times
	for i = 0; i < retry; i++ {
		id = prefix + h.Haikunate()
		if len(id) <= 63 {
			return Id(id), nil
		}
	}
	return "", errors.New("could not generate a DNS-compatible ID")
}

// NewAPITenantFromCRD transforms a CRD tenant into the API representation
func NewAPITenantFromCRD(tenant synv1alpha1.Tenant) *Tenant {
	apiTenant := &Tenant{
		TenantId: TenantId{Id: Id(tenant.Name)},
		TenantProperties: TenantProperties{
			GitRepo: &RevisionedGitRepo{},
		},
	}

	if len(tenant.Annotations) > 0 {
		apiTenant.Annotations = &Annotations{}
		for key, val := range tenant.Annotations {
			(*apiTenant.Annotations)[key] = val
		}
	}

	if len(tenant.Spec.DisplayName) > 0 {
		apiTenant.DisplayName = &tenant.Spec.DisplayName
	}

	if len(tenant.Spec.GitRepoURL) > 0 {
		apiTenant.GitRepo.Url = &tenant.Spec.GitRepoURL
	}

	if len(tenant.Spec.GitRepoRevision) > 0 {
		apiTenant.GitRepo.Revision = Revision{&tenant.Spec.GitRepoRevision}
	}

	if len(tenant.Spec.GlobalGitRepoURL) > 0 {
		apiTenant.GlobalGitRepoURL = &tenant.Spec.GlobalGitRepoURL
	}

	if len(tenant.Spec.GlobalGitRepoRevision) > 0 {
		apiTenant.GlobalGitRepoRevision = &tenant.Spec.GlobalGitRepoRevision
	}

	if tenant.Spec.GitRepoTemplate != nil {
		if len(tenant.Spec.GitRepoTemplate.RepoType) > 0 {
			repoType := string(tenant.Spec.GitRepoTemplate.RepoType)
			apiTenant.GitRepo.Type = &repoType
		}
	}

	return apiTenant
}

// NewCRDFromAPITenant transforms an API tenant into the CRD representation
func NewCRDFromAPITenant(apiTenant Tenant) *synv1alpha1.Tenant {
	tenant := &synv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:        apiTenant.Id.String(),
			Annotations: map[string]string{},
		},
	}

	if apiTenant.GitRepo != nil {
		tenant.Spec.GitRepoTemplate = newGitRepoTemplate(&apiTenant.GitRepo.GitRepo, string(apiTenant.Id))
	}

	SyncCRDFromAPITenant(apiTenant.TenantProperties, tenant)

	return tenant
}

func SyncCRDFromAPITenant(source TenantProperties, target *synv1alpha1.Tenant) {
	if source.Annotations != nil {
		if target.Annotations == nil {
			target.Annotations = map[string]string{}
		}
		for key, val := range *source.Annotations {
			if str, ok := val.(string); ok {
				target.Annotations[key] = str
			}
		}
	}

	if source.DisplayName != nil {
		target.Spec.DisplayName = *source.DisplayName
	}

	if source.GitRepo != nil {
		if source.GitRepo.Url != nil {
			target.Spec.GitRepoURL = *source.GitRepo.Url
		}
		if source.GitRepo.Revision.Revision != nil {
			target.Spec.GitRepoRevision = *source.GitRepo.Revision.Revision
		}
	}

	if source.GlobalGitRepoURL != nil {
		target.Spec.GlobalGitRepoURL = *source.GlobalGitRepoURL
	}

	if source.GlobalGitRepoRevision != nil {
		target.Spec.GlobalGitRepoRevision = *source.GlobalGitRepoRevision
	}
}

// NewAPIClusterFromCRD transforms a CRD cluster into the API representation
func NewAPIClusterFromCRD(cluster synv1alpha1.Cluster) *Cluster {
	apiCluster := &Cluster{
		ClusterId: ClusterId{Id: Id(cluster.Name)},
		ClusterProperties: ClusterProperties{
			GitRepo: &GitRepo{},
		},
		ClusterTenant: ClusterTenant{Tenant: cluster.Spec.TenantRef.Name},
	}

	if len(cluster.Annotations) > 0 {
		apiCluster.Annotations = &Annotations{}
		for key, val := range cluster.Annotations {
			(*apiCluster.Annotations)[key] = val
		}
	}

	if len(cluster.Spec.DisplayName) > 0 {
		apiCluster.DisplayName = &cluster.Spec.DisplayName
	}

	if len(cluster.Spec.GitRepoURL) > 0 {
		apiCluster.GitRepo.Url = &cluster.Spec.GitRepoURL
	}

	if len(cluster.Spec.GitHostKeys) > 0 {
		apiCluster.GitRepo.HostKeys = &cluster.Spec.GitHostKeys
	}

	if len(cluster.Spec.TenantGitRepoRevision) > 0 {
		apiCluster.TenantGitRepoRevision = &cluster.Spec.TenantGitRepoRevision
	}

	if len(cluster.Spec.GlobalGitRepoRevision) > 0 {
		apiCluster.GlobalGitRepoRevision = &cluster.Spec.GlobalGitRepoRevision
	}

	if cluster.Spec.Facts != nil {
		facts := ClusterFacts{}
		for key, value := range cluster.Spec.Facts {
			facts[key] = value
		}
		apiCluster.Facts = &facts
	}

	if cluster.Status.Facts != nil {
		facts := DynamicClusterFacts{}
		for key, value := range cluster.Status.Facts {
			facts[key] = unmarshalFact(value)
		}
		apiCluster.DynamicFacts = &facts
	}

	if cluster.Spec.GitRepoTemplate != nil {
		if stewardKey, ok := cluster.Spec.GitRepoTemplate.DeployKeys["steward"]; ok {
			sshKey := fmt.Sprintf("%s %s", stewardKey.Type, stewardKey.Key)
			apiCluster.GitRepo.DeployKey = &sshKey
		}
		if len(cluster.Spec.GitRepoTemplate.RepoType) > 0 {
			repoType := string(cluster.Spec.GitRepoTemplate.RepoType)
			apiCluster.GitRepo.Type = &repoType
		}
	}

	return apiCluster
}

func unmarshalFact(fact string) interface{} {
	var intFact interface{}
	err := json.Unmarshal([]byte(fact), &intFact)
	if err != nil {
		// The given string is not a JSON value
		// Fall back to returning the raw string
		return fact
	}
	return intFact
}

// NewCRDFromAPICluster transforms an API cluster into the CRD representation
func NewCRDFromAPICluster(apiCluster Cluster) (*synv1alpha1.Cluster, error) {
	cluster := &synv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:        string(apiCluster.ClusterId.Id),
			Annotations: map[string]string{},
		},
		Spec: synv1alpha1.ClusterSpec{
			TenantRef: corev1.LocalObjectReference{
				Name: apiCluster.Tenant,
			},
		},
	}

	cluster.Spec.GitRepoTemplate = newGitRepoTemplate(apiCluster.GitRepo, string(apiCluster.Id))

	if apiCluster.GitRepo != nil && apiCluster.GitRepo.HostKeys != nil {
		cluster.Spec.GitHostKeys = *apiCluster.GitRepo.HostKeys
	}

	if err := SyncCRDFromAPICluster(apiCluster.ClusterProperties, cluster); err != nil {
		return cluster, err
	}

	return cluster, nil
}

func SyncCRDFromAPICluster(source ClusterProperties, target *synv1alpha1.Cluster) error {
	if source.Annotations != nil {
		if target.Annotations == nil {
			target.Annotations = map[string]string{}
		}
		for key, val := range *source.Annotations {
			if str, ok := val.(string); ok {
				target.Annotations[key] = str
			}
		}
	}

	if source.DisplayName != nil {
		target.Spec.DisplayName = *source.DisplayName
	}

	if source.GitRepo != nil {
		if source.GitRepo.Url != nil {
			target.Spec.GitRepoURL = *source.GitRepo.Url
		}
		if source.GitRepo.HostKeys != nil {
			target.Spec.GitHostKeys = *source.GitRepo.HostKeys
		}

		if source.GitRepo.DeployKey != nil {
			if target.Spec.GitRepoTemplate.DeployKeys == nil {
				target.Spec.GitRepoTemplate.DeployKeys = make(map[string]synv1alpha1.DeployKey)
			}

			k := strings.Split(*source.GitRepo.DeployKey, " ")
			target.Spec.GitRepoTemplate.DeployKeys["steward"] = synv1alpha1.DeployKey{
				Type:        k[0],
				Key:         k[1],
				WriteAccess: false,
			}
		}
	}

	if source.TenantGitRepoRevision != nil {
		target.Spec.TenantGitRepoRevision = *source.TenantGitRepoRevision
	}

	if source.GlobalGitRepoRevision != nil {
		target.Spec.GlobalGitRepoRevision = *source.GlobalGitRepoRevision
	}

	if source.Facts != nil {
		if target.Spec.Facts == nil {
			target.Spec.Facts = synv1alpha1.Facts{}
		}

		for key, value := range *source.Facts {
			if valueStr, ok := value.(string); ok {
				target.Spec.Facts[key] = valueStr
			}
		}
	}

	if source.DynamicFacts != nil {
		if target.Status.Facts == nil {
			target.Status.Facts = synv1alpha1.Facts{}
		}

		for key, value := range *source.DynamicFacts {
			encodedFact, err := json.Marshal(value)
			if err != nil {
				return err
			}
			target.Status.Facts[key] = string(encodedFact)
		}
	}
	return nil
}

func newGitRepoTemplate(repo *GitRepo, name string) *synv1alpha1.GitRepoTemplate {
	if repo == nil {
		// No git info was specified
		return nil
	}

	if repo.Type == nil || *repo.Type != string(synv1alpha1.UnmanagedRepoType) {
		if repo.Url != nil && len(*repo.Url) > 0 {
			// It's not unmanaged and the URL was specified, take it apart
			url, err := url.Parse(*repo.Url)
			if err != nil {
				return nil
			}
			pathParts := strings.Split(url.Path, "/")
			pathParts = pathParts[1:]
			if len(pathParts) < 2 {
				return nil
			}
			// remove .git extension
			repoName := strings.ReplaceAll(pathParts[len(pathParts)-1], ".git", "")
			repoPath := strings.Join(pathParts[:len(pathParts)-1], "/")
			return &synv1alpha1.GitRepoTemplate{
				RepoType: synv1alpha1.AutoRepoType,
				Path:     repoPath,
				RepoName: repoName,
			}
		}
	} else if repo.Type != nil {
		// Repo is unmanaged, remove name and path
		return &synv1alpha1.GitRepoTemplate{
			RepoType: synv1alpha1.UnmanagedRepoType,
		}
	}
	return nil
}
