package api

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/url"
	"strings"

	synv1alpha1 "github.com/projectsyn/lieutenant-operator/pkg/apis/syn/v1alpha1"
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

	idCharset = "abcdefghijklmnopqrstuvwxyz" + "0123456789"
)

// GenerateClusterID creates a new cluster id
func GenerateClusterID() (ClusterId, error) {
	id, err := generateID()
	if err != nil {
		return ClusterId{}, err
	}
	return ClusterId{
		Id: Id(ClusterIDPrefix + id),
	}, nil
}

// GenerateTenantID creates a new tenant id
func GenerateTenantID() (TenantId, error) {
	id, err := generateID()
	if err != nil {
		return TenantId{}, err
	}
	return TenantId{
		Id: Id(TenantIDPrefix + id),
	}, nil
}

// GenerateID generates a new id from random alphanumeric characters
func generateID() (Id, error) {
	id := strings.Builder{}
	for i := 0; i < 6; i++ {
		max := big.NewInt(int64(len(idCharset)))
		r, err := rand.Int(rand.Reader, max)
		if err != nil {
			return "", err
		}
		_, err = id.WriteString(string(idCharset[r.Int64()]))
		if err != nil {
			return "", err
		}
	}
	return Id(id.String()), nil
}

// NewAPITenantFromCRD transforms a CRD tenant into the API representation
func NewAPITenantFromCRD(tenant synv1alpha1.Tenant) *Tenant {
	apiTenant := &Tenant{
		TenantId: TenantId{Id: Id(tenant.Name)},
		TenantProperties: TenantProperties{
			GitRepo: &GitRepo{},
		},
	}

	if len(tenant.Spec.DisplayName) > 0 {
		apiTenant.DisplayName = &tenant.Spec.DisplayName
	}

	if len(tenant.Spec.GitRepoURL) > 0 {
		apiTenant.GitRepo.Url = &tenant.Spec.GitRepoURL
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
			Name: string(apiTenant.TenantId.Id),
		},
	}
	if apiTenant.DisplayName != nil {
		tenant.Spec.DisplayName = *apiTenant.DisplayName
	}

	tenant.Spec.GitRepoTemplate = newGitRepoTemplate(apiTenant.GitRepo, string(apiTenant.Id))
	if apiTenant.GitRepo != nil {
		if apiTenant.GitRepo.Url != nil {
			tenant.Spec.GitRepoURL = *apiTenant.GitRepo.Url
		}
	}
	return tenant
}

// NewAPIClusterFromCRD transforms a CRD cluster into the API representation
func NewAPIClusterFromCRD(cluster synv1alpha1.Cluster) *Cluster {
	apiCluster := &Cluster{
		ClusterId: ClusterId{Id: Id(cluster.Name)},
		ClusterProperties: ClusterProperties{
			Tenant:  cluster.Spec.TenantRef.Name,
			GitRepo: &GitRepo{},
		},
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

	if cluster.Spec.Facts != nil {
		facts := ClusterFacts{}
		for key, value := range *cluster.Spec.Facts {
			facts[string(key)] = value
		}
		apiCluster.Facts = &facts
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

// NewCRDFromAPICluster transforms an API cluster into the CRD representation
func NewCRDFromAPICluster(apiCluster Cluster) *synv1alpha1.Cluster {
	cluster := &synv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: string(apiCluster.ClusterId.Id),
		},
		Spec: synv1alpha1.ClusterSpec{
			TenantRef: corev1.LocalObjectReference{
				Name: apiCluster.Tenant,
			},
		},
	}
	if apiCluster.DisplayName != nil {
		cluster.Spec.DisplayName = *apiCluster.DisplayName
	}
	cluster.Spec.GitRepoTemplate = newGitRepoTemplate(apiCluster.GitRepo, string(apiCluster.Id))
	if apiCluster.GitRepo != nil {
		if apiCluster.GitRepo.HostKeys != nil {
			cluster.Spec.GitHostKeys = *apiCluster.GitRepo.HostKeys
		}
		if apiCluster.GitRepo.Url != nil {
			cluster.Spec.GitRepoURL = *apiCluster.GitRepo.Url
		}
	}
	if apiCluster.Facts != nil {
		facts := synv1alpha1.Facts{}
		for key, value := range *apiCluster.Facts {
			if valueStr, ok := value.(string); ok {
				facts[key] = valueStr
			}
		}
	}
	return cluster
}

func newGitRepoTemplate(repo *GitRepo, name string) *synv1alpha1.GitRepoTemplate {
	//TODO: this default repoTemplate should be configurable (ideally per tenant)
	repoTemplate := &synv1alpha1.GitRepoTemplate{
		Path:     "syn/cluster-catalogs",
		RepoName: name,
		RepoType: synv1alpha1.AutoRepoType,
		APISecretRef: corev1.SecretReference{
			Name: "vshn-gitlab",
		},
	}
	if repo == nil {
		// No git info was specified, just return the default
		return repoTemplate
	}

	if repo.Type == nil || *repo.Type != string(synv1alpha1.UnmanagedRepoType) {
		if repo.Url != nil && len(*repo.Url) > 0 {
			// It's not unmanaged and the URL was specified, take it apart
			url, err := url.Parse(*repo.Url)
			if err != nil {
				return &synv1alpha1.GitRepoTemplate{}
			}
			pathParts := strings.Split(url.Path, "/")
			pathParts = pathParts[1:len(pathParts)]
			if len(pathParts) < 2 {
				return &synv1alpha1.GitRepoTemplate{}
			}
			// remove .git extension
			repoName := strings.ReplaceAll(pathParts[len(pathParts)-1], ".git", "")
			repoPath := strings.Join(pathParts[:len(pathParts)-1], "/")
			repoTemplate.Path = repoPath
			repoTemplate.RepoName = repoName
		}
	} else if repo.Type != nil {
		repoTemplate.RepoType = synv1alpha1.UnmanagedRepoType
		// Repo is unmanaged, remove name and path
		repoTemplate.RepoName = ""
		repoTemplate.Path = ""
	}
	return repoTemplate
}
