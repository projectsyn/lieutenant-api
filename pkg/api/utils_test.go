package api

import (
	"testing"

	"github.com/AlekSi/pointer"
	"github.com/stretchr/testify/assert"
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
