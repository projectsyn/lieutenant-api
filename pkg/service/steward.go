package service

import (
	"net/http"
	"os"
	"time"

	"github.com/AlekSi/pointer"
	"github.com/labstack/echo/v4"
	synv1alpha1 "github.com/projectsyn/lieutenant-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/projectsyn/lieutenant-api/pkg/api"
)

const (
	namespace           = "syn"
	appName             = "steward"
	stewardImageDefault = "docker.io/projectsyn/steward:v0.2.2"
)

var (
	appLabels = map[string]string{
		"app.kubernetes.io/name":       appName,
		"app.kubernetes.io/managed-by": "syn",
	}
)

// InstallSteward returns the JSON to install Steward on a cluster
func (s *APIImpl) InstallSteward(c echo.Context, params api.InstallStewardParams) error {
	ctx := c.(*APIContext)

	if params.Token == nil || len(*params.Token) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing or malformed token")
	}

	clusterList := &synv1alpha1.ClusterList{}
	if err := ctx.client.List(ctx.Request().Context(), clusterList, client.InNamespace(s.namespace)); err != nil {
		return err
	}
	var token string
	cluster := synv1alpha1.Cluster{}
	for _, c := range clusterList.Items {
		if bToken := c.Status.BootstrapToken; bToken != nil {
			if len(bToken.Token) > 0 && bToken.Token == *params.Token {
				if bToken.TokenValid && time.Now().Before(bToken.ValidUntil.Time) {
					t, err := s.getServiceAccountToken(ctx, c.Name)
					if err != nil {
						return err
					}
					token = t
					cluster = c
				} else {
					return echo.NewHTTPError(http.StatusUnauthorized, "Token already used or expired")
				}
			}
		}
	}
	if len(token) == 0 {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
	}

	installList := &corev1.List{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "List",
		},
	}
	apiHost := ctx.Scheme() + "://" + ctx.Request().Host
	stewardDeployment := createStewardDeployment(apiHost, cluster.Name)
	installList.Items = append(installList.Items, runtime.RawExtension{Object: &corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:   namespace,
			Labels: appLabels,
		},
	}})
	installList.Items = append(installList.Items, createRBAC()...)
	installList.Items = append(installList.Items, runtime.RawExtension{Object: stewardDeployment})
	installList.Items = append(installList.Items, runtime.RawExtension{Object: createSecret(token)})
	if err := ctx.JSON(http.StatusOK, installList); err != nil {
		return err
	}
	cluster.Status.BootstrapToken.TokenValid = false
	return ctx.client.Status().Update(ctx.Request().Context(), &cluster)
}

func (s *APIImpl) getServiceAccountToken(ctx *APIContext, saName string) (string, error) {

	secrets := &corev1.SecretList{}
	if err := ctx.client.List(
		ctx.Request().Context(),
		secrets,
		client.InNamespace(s.namespace),
		client.MatchingFields{"type": string(corev1.SecretTypeServiceAccountToken)},
	); err != nil {
		return "", err
	}

	token := findOldestSAToken(secrets.Items, saName)
	if token == "" {
		return "", echo.NewHTTPError(http.StatusServiceUnavailable, "Unable to find token for Cluster. This error might be transient, please try again.")
	}
	return token, nil
}

func findOldestSAToken(secrets []corev1.Secret, saName string) string {
	token := ""
	var created *metav1.Time

	for i, secret := range secrets {
		if secret.Type == corev1.SecretTypeServiceAccountToken && // Not strictly necessary but our testing framework can't handle field selectors
			secret.Annotations[corev1.ServiceAccountNameKey] == saName &&
			len(secret.Data[corev1.ServiceAccountTokenKey]) > 0 &&
			!created.Before(&secret.CreationTimestamp) {

			token = string(secret.Data[corev1.ServiceAccountTokenKey])
			created = &secrets[i].CreationTimestamp
		}
	}
	return token
}

func createRBAC() []runtime.RawExtension {
	return []runtime.RawExtension{{
		Object: &rbacv1.ClusterRole{
			TypeMeta: metav1.TypeMeta{
				APIVersion: rbacv1.SchemeGroupVersion.String(),
				Kind:       "ClusterRole",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:   "syn-admin",
				Labels: appLabels,
			},
			Rules: []rbacv1.PolicyRule{{
				APIGroups: []string{"*"},
				Resources: []string{"*"},
				Verbs:     []string{"*"},
			}, {
				NonResourceURLs: []string{"*"},
				Verbs:           []string{"*"},
			}},
		},
	}, {
		Object: &rbacv1.ClusterRoleBinding{
			TypeMeta: metav1.TypeMeta{
				APIVersion: rbacv1.SchemeGroupVersion.String(),
				Kind:       "ClusterRoleBinding",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:   "syn-steward",
				Labels: appLabels,
			},
			RoleRef: rbacv1.RoleRef{
				APIGroup: rbacv1.SchemeGroupVersion.Group,
				Kind:     "ClusterRole",
				Name:     "syn-admin",
			},
			Subjects: []rbacv1.Subject{{
				Kind:      "ServiceAccount",
				Name:      appName,
				Namespace: namespace,
			}},
		},
	}, {
		Object: &corev1.ServiceAccount{
			TypeMeta: metav1.TypeMeta{
				APIVersion: corev1.SchemeGroupVersion.String(),
				Kind:       "ServiceAccount",
			},
			ObjectMeta: metav1.ObjectMeta{
				Name:      appName,
				Namespace: namespace,
				Labels:    appLabels,
			},
		},
	}}
}

func createSecret(token string) *corev1.Secret {
	return &corev1.Secret{
		TypeMeta: metav1.TypeMeta{
			APIVersion: corev1.SchemeGroupVersion.String(),
			Kind:       "Secret",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: namespace,
			Labels:    appLabels,
		},
		StringData: map[string]string{
			"token": token,
		},
	}
}

func createStewardDeployment(apiHost, clusterID string) *appsv1.Deployment {
	image := os.Getenv("STEWARD_IMAGE")
	if len(image) == 0 {
		image = stewardImageDefault
	}
	apiHostEnv := os.Getenv("API_HOST")
	if len(apiHostEnv) > 0 {
		apiHost = apiHostEnv
	}
	stewardDeployment := &appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      appName,
			Namespace: namespace,
			Labels:    appLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: appLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: appLabels,
				},
				Spec: corev1.PodSpec{
					SecurityContext: &corev1.PodSecurityContext{
						RunAsNonRoot: pointer.ToBool(true),
					},
					ServiceAccountName: appName,
					Containers: []corev1.Container{{
						Name:            appName,
						Image:           image,
						ImagePullPolicy: corev1.PullAlways,
						Env: []corev1.EnvVar{
							{
								Name:  "STEWARD_API",
								Value: apiHost,
							},
							{
								Name:  "STEWARD_CLUSTER_ID",
								Value: clusterID,
							},
							{
								Name: "STEWARD_TOKEN",
								ValueFrom: &corev1.EnvVarSource{
									SecretKeyRef: &corev1.SecretKeySelector{
										LocalObjectReference: corev1.LocalObjectReference{
											Name: appName,
										},
										Key: "token",
									},
								},
							},
							{
								Name: "STEWARD_NAMESPACE",
								ValueFrom: &corev1.EnvVarSource{
									FieldRef: &corev1.ObjectFieldSelector{
										FieldPath: "metadata.namespace",
									},
								},
							},
						},
						Resources: corev1.ResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("100m"),
								corev1.ResourceMemory: resource.MustParse("32Mi"),
							},
							Limits: corev1.ResourceList{
								corev1.ResourceCPU:    resource.MustParse("200m"),
								corev1.ResourceMemory: resource.MustParse("64Mi"),
							},
						},
					}},
				},
			},
		},
	}
	return stewardDeployment
}
