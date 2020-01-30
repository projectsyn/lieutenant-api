package service

import (
	"github.com/labstack/echo/v4"
	"github.com/projectsyn/lieutenant-api/pkg/api"
	"net/http"
)

// InstallSteward returns the JSON to install Steward on a cluster
func (s *APIImpl) InstallSteward(ctx echo.Context, params api.InstallStewardParams) error {
	if params.Token == nil || len(*params.Token) == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "Missing or malformed token")
	}
	install := map[string]interface{}{
		"apiVersion": "apps/v1",
		"kind":       "Deployment",
		"metadata": map[string]interface{}{
			"name":      "steward",
			"namespace": "syn",
			"labels": map[string]string{
				"app.kubernetes.io/name": "steward",
			},
		},
	}
	return ctx.JSON(http.StatusOK, install)
}
