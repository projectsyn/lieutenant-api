package service

import (
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestInexistentKubeConfig(t *testing.T) {
	err := os.Setenv("KUBECONFIG", "/non/existing")
	assert.NoError(t, err)
	_, err = getClientFromToken("sometoken")
	assert.Error(t, err)
}
