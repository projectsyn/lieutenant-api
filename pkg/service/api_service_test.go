package service

import (
	"net/http"
	"strings"
	"testing"

	"github.com/deepmap/oapi-codegen/pkg/testutil"
	"github.com/projectsyn/lieutenant/pkg/api"
	"github.com/stretchr/testify/assert"
)

func TestNewServer(t *testing.T) {
	swagger, err := api.GetSwagger()
	assert.NoError(t, err)

	server, err := NewAPIServer()
	assert.NoError(t, err)
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

func TestHealthz(t *testing.T) {
	// Setup
	e, err := NewAPIServer()
	assert.NoError(t, err)

	result := testutil.NewRequest().Get(APIBasePath+"/healthz").Go(t, e)
	assert.Equal(t, http.StatusOK, result.Code())
	assert.Equal(t, "ok", string(result.Recorder.Body.String()))
}
