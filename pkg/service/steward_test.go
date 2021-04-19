package service

import (
	"net/http"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"

	"github.com/projectsyn/lieutenant-api/pkg/api"
)

func TestInstallSteward(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		WithHeader("X-Forwarded-Proto", "https").
		Get("/install/steward.json?token="+clusterA.Status.BootstrapToken.Token).
		Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	manifests := &corev1.List{}
	err := result.UnmarshalJsonToObject(&manifests)
	assert.NoError(t, err)
	assert.Len(t, manifests.Items, 6)
	decoder := json.NewSerializer(json.DefaultMetaFactory, scheme, scheme, true)
	foundSecret := false
	foundDeployment := false
	for i, item := range manifests.Items {
		obj, err := runtime.Decode(decoder, item.Raw)
		assert.NoError(t, err)
		if i == 0 {
			_, ok := obj.(*corev1.Namespace)
			assert.True(t, ok, "First object needs to be a namespace")
		}
		if secret, ok := obj.(*corev1.Secret); ok {
			foundSecret = true
			assert.Equal(t, secret.StringData["token"], string(clusterASecret.Data["token"]))
		}
		if deployment, ok := obj.(*appsv1.Deployment); ok {
			foundDeployment = true
			assert.Equal(t, "https://example.com", deployment.Spec.Template.Spec.Containers[0].Env[0].Value)
			assert.Equal(t, clusterA.Name, deployment.Spec.Template.Spec.Containers[0].Env[1].Value)
		}
	}
	assert.True(t, foundSecret, "Could not find secret with steward token")
	assert.True(t, foundDeployment, "Could not find deployment for steward")
}

func TestInstallStewardNoToken(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Get("/install/steward.json").
		Go(t, e)
	assert.Equal(t, http.StatusBadRequest, result.Code())
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.Contains(t, reason.Reason, "Missing or malformed token")
}

func TestInstallStewardInvalidToken(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Get("/install/steward.json?token=NonExistentToken").
		Go(t, e)
	assert.Equal(t, http.StatusUnauthorized, result.Code())
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.Contains(t, reason.Reason, "Invalid token")
}

func TestInstallStewardUsedToken(t *testing.T) {
	e, _ := setupTest(t)

	result := testutil.NewRequest().
		Get("/install/steward.json?token="+clusterB.Status.BootstrapToken.Token).
		Go(t, e)
	assert.Equal(t, http.StatusUnauthorized, result.Code())
	reason := &api.Reason{}
	err := result.UnmarshalJsonToObject(reason)
	assert.NoError(t, err)
	assert.Contains(t, reason.Reason, "Token already used or expired")
}
