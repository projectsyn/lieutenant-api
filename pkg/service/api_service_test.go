package service

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/labstack/echo/v4"
	"github.com/projectsyn/lieutenant-api/pkg/api"
	synv1alpha1 "github.com/projectsyn/lieutenant-operator/pkg/apis/syn/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

const bearerToken = AuthScheme + " eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

var (
	tenantA = &synv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-a",
			Namespace: "default",
		},
		Spec: synv1alpha1.TenantSpec{
			DisplayName: "Tenant A",
			GitRepoURL:  "ssh://git@github.com/tenant-a/defaults",
		},
	}
	tenantB = &synv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-b",
			Namespace: "default",
		},
		Spec: synv1alpha1.TenantSpec{
			DisplayName: "Tenant B",
			GitRepoTemplate: &synv1alpha1.GitRepoTemplate{
				RepoName: "defaults",
				Path:     "tenant-a",
				APISecretRef: corev1.SecretReference{
					Name:      "api-creds",
					Namespace: "default",
				},
			},
		},
	}
	clusterA = &synv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-cluster-a",
			Namespace: "default",
		},
		Spec: synv1alpha1.ClusterSpec{
			DisplayName: "Sample Cluster A",
			GitRepoURL:  "ssh://git@github.com/example/repo.git",
			TenantRef: corev1.LocalObjectReference{
				Name: tenantA.Name,
			},
			Facts: &synv1alpha1.Facts{
				"cloud": "cloudscale",
			},
		},
		Status: synv1alpha1.ClusterStatus{
			BootstrapToken: &synv1alpha1.BootstrapToken{
				Token:               "haevechee2ethot",
				BootstrapTokenValid: true,
				ValidUntil:          metav1.NewTime(time.Now().Add(30 * time.Minute)),
			},
		},
	}
	clusterB = &synv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-cluster-b",
			Namespace: "default",
		},
		Spec: synv1alpha1.ClusterSpec{
			DisplayName: "Sample Cluster B",
			GitRepoURL:  "ssh://git@github.com/example/repo.git",
			TenantRef: corev1.LocalObjectReference{
				Name: tenantB.Name,
			},
			GitRepoTemplate: &synv1alpha1.GitRepoTemplate{
				Path:         tenantB.Spec.GitRepoTemplate.Path,
				APISecretRef: tenantB.Spec.GitRepoTemplate.APISecretRef,
				RepoName:     "cluster-b",
				DeployKeys: map[string]synv1alpha1.DeployKey{
					"steward": synv1alpha1.DeployKey{
						Type: "ssh-ed25519",
						Key:  "AAAAC3NzaC1lZDI1NTE5AAAAIPEx4k5NQ46DA+m49Sb3aIyAAqqbz7TdHbArmnnYqwjf",
					},
				},
			},
			Facts: &synv1alpha1.Facts{
				"cloud": "cloudscale",
			},
		},
	}
	clusterASecret = &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      clusterA.Name,
			Namespace: clusterA.Namespace,
		},
		Data: map[string][]byte{"token": []byte("sometoken")},
	}
	testObjects = []runtime.Object{
		tenantA,
		tenantB,
		clusterA,
		clusterASecret,
		clusterB,
		&corev1.ServiceAccount{
			ObjectMeta: metav1.ObjectMeta{
				Name:      clusterA.Name,
				Namespace: clusterA.Namespace,
			},
			Secrets: []corev1.ObjectReference{{
				Name:      clusterASecret.Name,
				Namespace: clusterASecret.Namespace,
			}},
		},
	}
)

func TestNewServer(t *testing.T) {
	swagger, err := api.GetSwagger()
	assert.NoError(t, err)

	server := setupTest(t)
	for _, route := range server.Routes() {
		if route.Path == APIBasePath || strings.HasSuffix(route.Path, "*") {
			continue
		}
		p := strings.TrimPrefix(route.Path, APIBasePath)
		if strings.ContainsRune(p, ':') {
			p = strings.Replace(p, ":", "{", 1) + "}"
		}
		path := swagger.Paths.Find(p)
		assert.NotNil(t, path, p)
	}
}

func setupTest(t *testing.T, objs ...[]runtime.Object) *echo.Echo {
	testMiddleWare := KubernetesAuth{
		CreateClientFunc: func(token string) (client.Client, error) {
			return fake.NewFakeClientWithScheme(scheme.Scheme, testObjects...), nil
		},
	}
	e, err := NewAPIServer(testMiddleWare)
	assert.NoError(t, err)
	return e
}

func TestHealthz(t *testing.T) {
	e := setupTest(t)

	result := testutil.NewRequest().Get(APIBasePath+"/healthz").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.Equal(t, "ok", string(result.Recorder.Body.String()))
}
