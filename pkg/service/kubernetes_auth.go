package service

import (
	"context"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// AuthScheme to be used in the Authorization header
const (
	AuthScheme = "Bearer"
)

var (
	// The following endpoints don't need auth
	noAuth = map[string]bool{
		"/healthz":      true,
		"/openapi.json": true,
		"/docs":         true,
	}
)

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

		if _, ok := noAuth[c.Path()]; ok {
			return next(c)
		}

		if strings.HasSuffix(c.Path(), "/install/steward.json") {
			// Special case for installing Steward:
			// The bootstrap token will be used and the lieutenants kubeconfig.
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
			context: context.TODO(),
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
		Scheme: scheme,
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
