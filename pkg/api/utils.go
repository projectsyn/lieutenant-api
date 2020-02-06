package api

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"strings"

	"github.com/AlekSi/pointer"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	synv1alpha1 "github.com/projectsyn/lieutenant-operator/pkg/apis/syn/v1alpha1"
)

const idCharset = "abcdefghijklmnopqrstuvwxyz" + "0123456789"

// NewClusterID creates a new id from a string
func NewClusterID(id string) ClusterId {
	return ClusterId{
		Id: Id(id),
	}
}

// NewTenantID creates a new id from a string
func NewTenantID(id string) TenantId {
	return TenantId{
		Id: Id(id),
	}
}

func GenerateID() (Id, error) {
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

func NewAPITenantFromCRD(tenant *synv1alpha1.Tenant) *Tenant {
	apiTenant := &Tenant{
		TenantId: NewTenantID(tenant.Name),
	}

	if len(tenant.Spec.DisplayName) > 0 {
		apiTenant.DisplayName = &tenant.Spec.DisplayName
	}

	if len(tenant.Spec.GitRepoURL) > 0 {
		apiTenant.GitRepo = &tenant.Spec.GitRepoURL
	}

	return apiTenant
}

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
	}

	return tenant
}

func NewAPIClusterFromCRD(cluster *synv1alpha1.Cluster) *Cluster {
	apiCluster := &Cluster{
		ClusterId: NewClusterID(cluster.Name),
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
		deployKeys := cluster.Spec.GitRepoTemplate.Spec.DeployKeys
		if len(deployKeys) > 0 {
			key := deployKeys[0]
			sshKey := fmt.Sprintf("%s %s", key.Type, key.Key)
			apiCluster.SshDeployKey = &sshKey
		}
	}

	return apiCluster
}

func NewCRDFromAPICluster(apiCluster *Cluster) *synv1alpha1.Cluster {
	cluster := &synv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name: string(apiCluster.ClusterId.Id),
		},
		Spec: synv1alpha1.ClusterSpec{
			TenantRef: synv1alpha1.TenantRef{
				Name: apiCluster.Tenant,
			},
		},
	}
	if apiCluster.DisplayName != nil {
		cluster.Spec.DisplayName = *apiCluster.DisplayName
	}
	if apiCluster.GitRepo != nil {
		cluster.Spec.GitRepoURL = *apiCluster.GitRepo
	}
	if apiCluster.Facts != nil {
		facts := synv1alpha1.Facts{}
		for key, value := range *apiCluster.Facts {
			if valueStr, ok := value.(string); ok {
				facts[synv1alpha1.FactKey(key)] = synv1alpha1.FactValue(valueStr)
			}
		}
	}
	return cluster
}
