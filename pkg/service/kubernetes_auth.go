package service

import (
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	synv1alpha1 "github.com/projectsyn/lieutenant-operator/pkg/apis/syn/v1alpha1"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// AuthScheme to be used in the Authorization header
const AuthScheme = "Bearer"

func init() {
	synv1alpha1.SchemeBuilder.AddToScheme(scheme.Scheme)
}

// KubernetesAuth provides middleware to authenticate with Kubernetes JWT tokens
type KubernetesAuth struct {
	CreateClientFunc func(string) (client.Client, error)
}

// DefaultKubernetesAuth uses the JWT bearer token to authenticate
var DefaultKubernetesAuth = &KubernetesAuth{
	CreateClientFunc: getClientFromToken,
}

// JWTAuth makes sure a JWT bearer token is provided and creates a Kubernetes client
func (k *KubernetesAuth) JWTAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		var token string
		// No auth needed for the health endpoint
		if strings.HasSuffix(c.Path(), "/healthz") {
			return next(c)
			// Special case for installing Steward:
			// The bootstrap token will be used and the lieutenants kubeconfig.
		} else if strings.HasSuffix(c.Path(), "/install/steward.json") {
			token = ""
		} else {
			t, err := extractToken(c)
			if err != nil {
				return err
			}
			token = t
		}

		client, err := k.CreateClientFunc(token)
		if err != nil {
			return err
		}

		apiContext := &APIContext{
			Context: c,
			client:  client,
		}
		return next(apiContext)
	}
}

func getClientFromToken(token string) (client.Client, error) {
	cfg, err := config.GetConfig()
	if err != nil {
		return nil, err
	}
	if len(token) > 0 {
		cfg.Username = ""
		cfg.Password = ""
		cfg.BearerToken = token
		cfg.BearerTokenFile = ""
		cfg.TLSClientConfig.KeyFile = ""
		cfg.TLSClientConfig.KeyData = []byte{}
		cfg.TLSClientConfig.CertFile = ""
		cfg.TLSClientConfig.CertData = []byte{}
	}
	return client.New(cfg, client.Options{
		Scheme: scheme.Scheme,
	})
}

func extractToken(c echo.Context) (string, error) {
	auth := c.Request().Header.Get(echo.HeaderAuthorization)
	l := len(AuthScheme)
	if len(auth) > l+1 && auth[:l] == AuthScheme {
		token := auth[l+1:]
		return strings.TrimSpace(token), nil
	}
	return "", middleware.ErrJWTMissing
}
