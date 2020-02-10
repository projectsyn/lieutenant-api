package api

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/AlekSi/pointer"
	synv1alpha1 "github.com/projectsyn/lieutenant-operator/pkg/apis/syn/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// ClusterIDPrefix is prefixed to all cluster IDs
	ClusterIDPrefix = "c-"
	// TenantIDPrefix is prefixed to all tenant IDs
	TenantIDPrefix = "t-"

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
func NewAPITenantFromCRD(tenant *synv1alpha1.Tenant) *Tenant {
	apiTenant := &Tenant{
		TenantId: TenantId{Id: Id(tenant.Name)},
	}

	if len(tenant.Spec.DisplayName) > 0 {
		apiTenant.DisplayName = &tenant.Spec.DisplayName
	}

	if len(tenant.Spec.GitRepoURL) > 0 {
		apiTenant.GitRepo = &tenant.Spec.GitRepoURL
	}

	return apiTenant
}

// NewCRDFromAPITenant transforms an API tenant into the CRD representation
func NewCRDFromAPITenant(apiTenant *Tenant) *synv1alpha1.Tenant {
	tenant := &synv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name: string(apiTenant.TenantId.Id),
		},
	}
	if apiTenant.DisplayName != nil {
		tenant.Spec.DisplayName = *apiTenant.DisplayName
	}

	if apiTenant.GitRepo != nil {
		tenant.Spec.GitRepoURL = *apiTenant.GitRepo
	} else {
		// TODO: properly generate GitRepoTemplate
		tenant.Spec.GitRepoTemplate = &synv1alpha1.GitRepoTemplate{
			Path:     "syn/cluster-catalogs",
			RepoName: string(apiTenant.Id),
			APISecretRef: corev1.SecretReference{
				Name: "vshn-gitlab",
			},
		}
	}

	return tenant
}

// NewAPIClusterFromCRD transforms a CRD cluster into the API representation
func NewAPIClusterFromCRD(cluster *synv1alpha1.Cluster) *Cluster {
	apiCluster := &Cluster{
		ClusterId: ClusterId{Id: Id(cluster.Name)},
		ClusterProperties: ClusterProperties{
			Tenant: cluster.Spec.TenantRef.Name,
		},
	}

	if len(cluster.Spec.DisplayName) > 0 {
		apiCluster.DisplayName = pointer.ToString(cluster.Spec.DisplayName)
	}

	if len(cluster.Spec.GitRepoURL) > 0 {
		apiCluster.GitRepo = pointer.ToString(cluster.Spec.GitRepoURL)
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
			apiCluster.SshDeployKey = &sshKey
		}
	}

	return apiCluster
}

// NewCRDFromAPICluster transforms an API cluster into the CRD representation
func NewCRDFromAPICluster(apiCluster *Cluster) *synv1alpha1.Cluster {
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
	if apiCluster.GitRepo != nil {
		cluster.Spec.GitRepoURL = *apiCluster.GitRepo
	} else {
		// TODO: properly generate GitRepoTemplate
		cluster.Spec.GitRepoTemplate = &synv1alpha1.GitRepoTemplate{
			Path:     "syn/cluster-catalogs",
			RepoName: fmt.Sprintf("%s-%s", apiCluster.Tenant, apiCluster.Id),
			APISecretRef: corev1.SecretReference{
				Name: "vshn-gitlab",
			},
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
