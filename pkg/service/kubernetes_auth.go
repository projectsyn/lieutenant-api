package service

import (
	"context"
	"os"
	"strconv"
	"strings"

	lruCache "github.com/hashicorp/golang-lru"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"k8s.io/apimachinery/pkg/util/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
)

// AuthScheme to be used in the Authorization header
const (
	AuthScheme         = "Bearer"
	K8sCacheSizeEnvKey = "K8S_AUTH_CLIENT_CACHE_SIZE"
)

var (
	// The following endpoints don't need auth
	noAuth = map[string]bool{
		"/healthz":      true,
		"/openapi.json": true,
		"/docs":         true,
	}
)

func getCacheSizeOrDefault(def int) int {
	rawSize := os.Getenv(K8sCacheSizeEnvKey)
	if rawSize == "" {
		return def
	}
	parsed, err := strconv.Atoi(rawSize)
	if err != nil || parsed <= 0 {
		return def
	}
	return parsed
}

func createCache() *lruCache.Cache {
	cache, err := lruCache.NewWithEvict(getCacheSizeOrDefault(128), nil)
	runtime.Must(err)
	return cache
}

// KubernetesAuth provides middleware to authenticate with Kubernetes JWT tokens
type KubernetesAuth struct {
	CreateClientFunc func(string) (client.Client, error)
	cache            *lruCache.Cache
}

// DefaultKubernetesAuth uses the JWT bearer token to authenticate
var DefaultKubernetesAuth = &KubernetesAuth{
	CreateClientFunc: getClientFromToken,
	cache:            createCache(),
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

		cachedClient, exists := k.cache.Get(token)
		if !exists {
			var err error
			cachedClient, err = k.CreateClientFunc(token)
			if err != nil {
				return err
			}
			k.cache.Add(token, cachedClient)
		}

		apiContext := &APIContext{
			Context: c,
			context: context.TODO(),
			client:  cachedClient.(client.Client),
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
