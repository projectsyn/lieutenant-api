package api

import (
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AlekSi/pointer"
)

func TestRepoConversionDefaultAuto(t *testing.T) {
	apiRepo := &GitRepo{
		Type: nil,
		Url:  nil,
	}
	repoName := "c-dshfjuhrtu"
	repoTemplate := newGitRepoTemplate(apiRepo, repoName)
	assert.Equal(t, repoName, repoTemplate.RepoName)
	assert.Equal(t, "vshn-gitlab", repoTemplate.APISecretRef.Name)
}

func TestRepoConversionUnmanagedo(t *testing.T) {
	apiRepo := &GitRepo{
		Type: pointer.ToString("unmanaged"),
		Url:  pointer.ToString("ssh://git@some.host/path/to/repo.git"),
	}
	repoTemplate := newGitRepoTemplate(apiRepo, "some-name")
	assert.Empty(t, repoTemplate.RepoName)
	assert.Empty(t, repoTemplate.Path)
	assert.Equal(t, "vshn-gitlab", repoTemplate.APISecretRef.Name)
}

func TestRepoConversionSpecSubGroupPath(t *testing.T) {
	repoName := "myName"
	repoPath := "path/to"
	apiRepo := &GitRepo{
		Type: pointer.ToString("auto"),
		Url:  pointer.ToString("ssh://git@some.host/" + repoPath + "/" + repoName + ".git"),
	}
	repoTemplate := newGitRepoTemplate(apiRepo, "some-name")
	assert.Equal(t, repoName, repoTemplate.RepoName)
	assert.Equal(t, repoPath, repoTemplate.Path)
	assert.Equal(t, "vshn-gitlab", repoTemplate.APISecretRef.Name)
}

func TestRepoConversionSpecPath(t *testing.T) {
	repoName := "myName"
	repoPath := "path"
	apiRepo := &GitRepo{
		Type: pointer.ToString("auto"),
		Url:  pointer.ToString("ssh://git@some.host/" + repoPath + "/" + repoName + ".git"),
	}
	repoTemplate := newGitRepoTemplate(apiRepo, "some-name")
	assert.Equal(t, repoName, repoTemplate.RepoName)
	assert.Equal(t, repoPath, repoTemplate.Path)
	assert.Equal(t, "vshn-gitlab", repoTemplate.APISecretRef.Name)
}

func TestGenerateClusterID(t *testing.T) {
	assertGeneratedID(t, ClusterIDPrefix, func() (s string) {
		id, err := GenerateClusterID()
		require.NoError(t, err)
		return string(id.Id)
	})
}

func TestGenerateTenantID(t *testing.T) {
	assertGeneratedID(t, TenantIDPrefix, func() (s string) {
		id, err := GenerateTenantID()
		require.NoError(t, err)
		return string(id.Id)
	})
}

func assertGeneratedID(t *testing.T, prefix string, supplier func() string) {
	// Verify generated ID so that it conforms to https: //kubernetes.io/docs/concepts/overview/working-with-objects/names/#names
	// Regex pattern tested on regexr.com
	r := regexp.MustCompile("^[a-z]-[a-z0-9]{3,}(-|_)[a-z0-9]{3,}(-|_)[0-9]+$")
	// Run the randomizer a few times
	for i := 1; i <= 1000; i++ {
		id := supplier()
		require.LessOrEqualf(t, len(id), 63, "Iteration %d: too long for a DNS-compatible name: %s", i, id)
		require.Regexpf(t, r, id, "Iteration %d: not in the form of 'adjective-noun-number' %s", i, id)
		require.True(t, strings.HasPrefix(id, prefix))
	}
}
