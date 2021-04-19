package service

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/getkin/kin-openapi/openapi2"
	"github.com/labstack/echo/v4"
	synv1alpha1 "github.com/projectsyn/lieutenant-operator/pkg/apis/syn/v1alpha1"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"

	"github.com/projectsyn/lieutenant-api/pkg/api"
)

const bearerToken = AuthScheme + " eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIiwibmFtZSI6IkpvaG4gRG9lIiwiaWF0IjoxNTE2MjM5MDIyfQ.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"

var (
	tenantA = &synv1alpha1.Tenant{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "tenant-a",
			Namespace: "default",
			Annotations: map[string]string{
				"some":                     "annotations",
				"monitoring.syn.tools/sla": "247",
			},
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
				RepoType: synv1alpha1.AutoRepoType,
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
			Annotations: map[string]string{
				"some":                     "value",
				"monitoring.syn.tools/sla": "247",
			},
		},
		Spec: synv1alpha1.ClusterSpec{
			DisplayName: "Sample Cluster A",
			GitHostKeys: "some keys",
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
				Token:      "haevechee2ethot",
				TokenValid: true,
				ValidUntil: metav1.NewTime(time.Now().Add(30 * time.Minute)),
			},
		},
	}
	clusterB = &synv1alpha1.Cluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "sample-cluster-b",
			Namespace: "default",
			Annotations: map[string]string{
				"existing": "annotation",
			},
		},
		Spec: synv1alpha1.ClusterSpec{
			DisplayName: "Sample Cluster B",
			TenantRef: corev1.LocalObjectReference{
				Name: tenantB.Name,
			},
			GitRepoTemplate: &synv1alpha1.GitRepoTemplate{
				Path:         tenantB.Spec.GitRepoTemplate.Path,
				APISecretRef: tenantB.Spec.GitRepoTemplate.APISecretRef,
				RepoName:     "cluster-b",
				RepoType:     synv1alpha1.AutoRepoType,
			},
			Facts: &synv1alpha1.Facts{
				"cloud": "cloudscale",
			},
		},
		Status: synv1alpha1.ClusterStatus{
			BootstrapToken: &synv1alpha1.BootstrapToken{
				Token:      "shuaCh1k",
				TokenValid: false,
				ValidUntil: metav1.NewTime(time.Now().Add(-1 * time.Hour)),
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
	testObjects = []client.Object{
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

	server, _ := setupTest(t)
	for _, route := range server.Routes() {
		if strings.HasSuffix(route.Path, "*") {
			continue
		}
		p := route.Path
		if strings.ContainsRune(p, ':') {
			p = strings.Replace(p, ":", "{", 1) + "}"
		}
		path := swagger.Paths.Find(p)
		assert.NotNil(t, path, p)
	}
}

func setupTest(t *testing.T, objs ...[]runtime.Object) (*echo.Echo, client.Client) {
	os.Setenv("NAMESPACE", "default")
	f := fake.NewClientBuilder().WithScheme(scheme).WithObjects(testObjects...).Build()
	testMiddleWare := KubernetesAuth{
		CreateClientFunc: func(token string) (client.Client, error) {
			return f, nil
		},
	}
	e, err := NewAPIServer(testMiddleWare)
	assert.NoError(t, err)
	return e, f
}

func TestHealthz(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().Get("/healthz").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.Equal(t, "ok", string(result.Recorder.Body.String()))
}

func TestOpenAPI(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().Get("/openapi.json").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	swaggerSpec := &openapi2.Swagger{}
	err := json.Unmarshal(result.Recorder.Body.Bytes(), swaggerSpec)
	assert.NoError(t, err)
	assert.NotNil(t, swaggerSpec)
	assert.Equal(t, "Lieutenant API", swaggerSpec.Info.Title)
}

func TestSwaggerUI(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().Get("/docs").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.NotEmpty(t, result.Recorder.Body.Bytes)
}
