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
		Id: &id,
	}, err
}

// GenerateTenantID creates a new tenant id
func GenerateTenantID() (TenantId, error) {
	id, err := generateID(TenantIDPrefix)
	return TenantId{
		Id: &id,
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
	id := Id(tenant.Name)
	apiTenant := &Tenant{
		TenantId: TenantId{Id: &id},
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
func NewCRDFromAPITenant(apiTenant Tenant) (*synv1alpha1.Tenant, error) {
	if !strings.HasPrefix(apiTenant.Id.String(), TenantIDPrefix) {
		if apiTenant.Id.String() == "" {
			id, err := GenerateTenantID()
			if err != nil {
				return nil, err
			}
			apiTenant.TenantId = id
		} else {
			id := TenantIDPrefix + *apiTenant.Id
			apiTenant.Id = &id
		}
	}
	if apiTenant.GitRepo == nil ||
		apiTenant.GitRepo.Url == nil ||
		*apiTenant.GitRepo.Url == "" {
		return nil, fmt.Errorf("GitRepo URL is required")
	}

	tenant := &synv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:        apiTenant.Id.String(),
			Annotations: map[string]string{},
		},
	}

	if apiTenant.GitRepo != nil {
		tmpl, err := newGitRepoTemplate(&apiTenant.GitRepo.GitRepo, apiTenant.Id.String())
		if err != nil {
			return nil, fmt.Errorf("failed to create git repo template: %w", err)
		}
		tenant.Spec.GitRepoTemplate = tmpl
	}

	SyncCRDFromAPITenant(apiTenant.TenantProperties, tenant)

	return tenant, nil
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
func NewAPIClusterFromCRD(cluster synv1alpha1.Cluster) (*Cluster, error) {
	id := Id(cluster.Name)
	apiCluster := &Cluster{
		ClusterId: ClusterId{Id: &id},
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

	acm, err := crdCompileMetaToAPICompileMeta(cluster.Status.CompileMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to convert compile meta: %w", err)
	}
	apiCluster.ClusterProperties.CompileMeta = acm

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

	return apiCluster, nil
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
	if !strings.HasPrefix(apiCluster.Id.String(), ClusterIDPrefix) {
		if apiCluster.Id.String() == "" {
			id, err := GenerateClusterID()
			if err != nil {
				return nil, err
			}
			apiCluster.ClusterId = id
		} else {
			id := Id(ClusterIDPrefix + apiCluster.Id.String())
			apiCluster.Id = &id
		}
	}
	cluster := &synv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:        apiCluster.Id.String(),
			Annotations: map[string]string{},
		},
		Spec: synv1alpha1.ClusterSpec{
			TenantRef: corev1.LocalObjectReference{
				Name: apiCluster.Tenant,
			},
		},
	}

	tmpl, err := newGitRepoTemplate(apiCluster.GitRepo, apiCluster.Id.String())
	if err != nil {
		return nil, fmt.Errorf("failed to create git repo template: %w", err)
	}
	cluster.Spec.GitRepoTemplate = tmpl

	if apiCluster.GitRepo != nil && apiCluster.GitRepo.HostKeys != nil {
		cluster.Spec.GitHostKeys = *apiCluster.GitRepo.HostKeys
	}

	err = SyncCRDFromAPICluster(apiCluster.ClusterProperties, cluster)
	return cluster, err
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
			if target.Spec.GitRepoTemplate == nil {
				return fmt.Errorf("Cannot set deploy key for unmanaged git repo")
			}
			if target.Spec.GitRepoTemplate.DeployKeys == nil {
				target.Spec.GitRepoTemplate.DeployKeys = make(map[string]synv1alpha1.DeployKey)
			}

			k := strings.Split(*source.GitRepo.DeployKey, " ")
			if len(k) != 2 {
				return fmt.Errorf("Illegal deploy key format. Expected '<type> <public key>'")
			}
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

	clcm, err := apiCompileMetaToCRDCompileMeta(source.CompileMeta)
	if err != nil {
		return fmt.Errorf("failed to convert compile meta: %w", err)
	}
	target.Status.CompileMeta = clcm

	return nil
}

func newGitRepoTemplate(repo *GitRepo, name string) (*synv1alpha1.GitRepoTemplate, error) {
	if repo == nil {
		// No git info was specified
		return nil, nil
	}

	if repo.Type == nil || *repo.Type != string(synv1alpha1.UnmanagedRepoType) {
		if repo.Url != nil && len(*repo.Url) > 0 {
			// It's not unmanaged and the URL was specified, take it apart
			url, err := url.Parse(*repo.Url)
			if err != nil {
				return nil, fmt.Errorf("failed to parse git repo URL: %w", err)
			}
			pathParts := strings.Split(url.Path, "/")
			pathParts = pathParts[1:]
			if len(pathParts) < 2 {
				return nil, fmt.Errorf("failed to parse git repo URL, expected 2+ path elements in '%s'", url.Path)
			}
			// remove .git extension
			repoName := strings.ReplaceAll(pathParts[len(pathParts)-1], ".git", "")
			repoPath := strings.Join(pathParts[:len(pathParts)-1], "/")
			return &synv1alpha1.GitRepoTemplate{
				RepoType: synv1alpha1.AutoRepoType,
				Path:     repoPath,
				RepoName: repoName,
			}, nil
		}
	} else if repo.Type != nil {
		// Repo is unmanaged, remove name and path
		return &synv1alpha1.GitRepoTemplate{
			RepoType: synv1alpha1.UnmanagedRepoType,
		}, nil
	}
	return nil, nil
}

// crdCompileMetaToAPICompileMeta converts a CRD compile meta to an API compile meta.
// Uses json marshalling to convert the structs since their codegen representations are very different.
// Errors only if the marshalling fails.
func crdCompileMetaToAPICompileMeta(crdCompileMeta synv1alpha1.CompileMeta) (*ClusterCompileMeta, error) {
	j, err := json.Marshal(crdCompileMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal compile meta for conversion: %w", err)
	}
	var apiCompileMeta ClusterCompileMeta
	err = json.Unmarshal(j, &apiCompileMeta)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal compile meta for conversion: %w", err)
	}
	return &apiCompileMeta, nil
}

// apiCompileMetaToCRDCompileMeta converts an API compile meta to a CRD compile meta.
// Uses json marshalling to convert the structs since their codegen representations are very different.
// Errors only if the marshalling fails.
func apiCompileMetaToCRDCompileMeta(apiCompileMeta *ClusterCompileMeta) (synv1alpha1.CompileMeta, error) {
	if apiCompileMeta == nil {
		return synv1alpha1.CompileMeta{}, nil
	}

	j, err := json.Marshal(apiCompileMeta)
	if err != nil {
		return synv1alpha1.CompileMeta{}, fmt.Errorf("failed to marshal compile meta for conversion: %w", err)
	}
	var crdCompileMeta synv1alpha1.CompileMeta
	err = json.Unmarshal(j, &crdCompileMeta)
	if err != nil {
		return synv1alpha1.CompileMeta{}, fmt.Errorf("failed to unmarshal compile meta for conversion: %w", err)
	}
	return crdCompileMeta, nil
}
