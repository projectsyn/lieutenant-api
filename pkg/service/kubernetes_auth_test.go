package service

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInexistentKubeConfig(t *testing.T) {
	err := os.Setenv("KUBECONFIG", "/non/existing")
	assert.NoError(t, err)
	_, err = getClientFromToken("sometoken")
	assert.Error(t, err)
}

func Test_getCacheSizeOrDefault(t *testing.T) {
	tests := map[string]struct {
		envValue     string
		defaultValue int
		expected     int
	}{
		"GivenNoEnvVar_WhenGet_ThenReturnDefault": {
			defaultValue: 128,
			envValue:     "",
			expected:     128,
		},
		"GivenInvalidEnvVar_WhenGet_ThenReturnDefault": {
			defaultValue: 128,
			envValue:     "infinite",
			expected:     128,
		},
		"GivenEnvVar_WhenGet_ThenReturnParsedValue": {
			defaultValue: 128,
			envValue:     "64",
			expected:     64,
		},
		"GivenNegativeEnvVar_WhenGet_ThenReturnDefault": {
			defaultValue: 128,
			envValue:     "-64",
			expected:     128,
		},
	}
	for name, tt := range tests {
		t.Run(name, func(t *testing.T) {
			assert.NoError(t, os.Setenv(K8sCacheSizeEnvKey, tt.envValue))
			result := getCacheSizeOrDefault(tt.defaultValue)
			assert.Equal(t, tt.expected, result)
		})
	}
}
