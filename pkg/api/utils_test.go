package api

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/projectsyn/lieutenant-operator/api/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/version"

	"github.com/AlekSi/pointer"
)

func TestRepoConversionDefaultAuto(t *testing.T) {
	apiRepo := &GitRepo{
		Type: nil,
		Url:  nil,
	}
	repoName := "c-dshfjuhrtu"
	repoTemplate := newGitRepoTemplate(apiRepo, repoName)
	assert.Nil(t, repoTemplate)
	assert.Nil(t, newGitRepoTemplate(nil, repoName))
}

func TestRepoConversionUnmanagedo(t *testing.T) {
	apiRepo := &GitRepo{
		Type: pointer.ToString("unmanaged"),
		Url:  pointer.ToString("ssh://git@some.host/path/to/repo.git"),
	}
	repoTemplate := newGitRepoTemplate(apiRepo, "some-name")
	assert.Empty(t, repoTemplate.RepoName)
	assert.Empty(t, repoTemplate.Path)
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
	assert.Empty(t, repoTemplate.APISecretRef.Name)
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
	assert.Empty(t, repoTemplate.APISecretRef.Name)
}

func TestRepoConversionFail(t *testing.T) {
	apiRepo := &GitRepo{
		Url: pointer.ToString("://git@some.host/group/example.git"),
	}
	repoTemplate := newGitRepoTemplate(apiRepo, "some-name")
	assert.Nil(t, repoTemplate)

	repoTemplate = newGitRepoTemplate(&GitRepo{
		Url: pointer.ToString("ssh://git@some.host/example.git"),
	}, "test")
	assert.Nil(t, repoTemplate)
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
		return id.Id.String()
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

var tenantTests = map[string]struct {
	properties TenantProperties
	spec       v1alpha1.TenantSpec
}{
	"empty": {
		TenantProperties{
			GitRepo: &RevisionedGitRepo{
				GitRepo: GitRepo{
					Url: pointer.ToString("ssh://git@example.com/foo/t-buzz.git"),
				},
			},
		},
		v1alpha1.TenantSpec{
			GitRepoURL: "ssh://git@example.com/foo/t-buzz.git",
		},
	},
	"global git URL": {
		TenantProperties{
			GlobalGitRepoURL: pointer.ToString("ssh://git@example.com/foo/bar.git"),
			GitRepo: &RevisionedGitRepo{
				GitRepo: GitRepo{
					Url: pointer.ToString("ssh://git@example.com/foo/t-buzz.git"),
				},
			},
		},
		v1alpha1.TenantSpec{
			GlobalGitRepoURL: "ssh://git@example.com/foo/bar.git",
			GitRepoURL:       "ssh://git@example.com/foo/t-buzz.git",
		},
	},
	"global git revision": {
		TenantProperties{
			GlobalGitRepoRevision: pointer.ToString("v1.2.3"),
			GitRepo: &RevisionedGitRepo{
				GitRepo: GitRepo{
					Url: pointer.ToString("ssh://git@example.com/foo/t-buzz.git"),
				},
			},
		},
		v1alpha1.TenantSpec{
			GlobalGitRepoRevision: "v1.2.3",
			GitRepoURL:            "ssh://git@example.com/foo/t-buzz.git",
		},
	},
	"git revision": {
		TenantProperties{
			GitRepo: &RevisionedGitRepo{
				Revision: Revision{pointer.ToString("v1.2.3")},
				GitRepo: GitRepo{
					Url: pointer.ToString("ssh://git@example.com/foo/t-buzz.git"),
				},
			},
		},
		v1alpha1.TenantSpec{
			GitRepoRevision: "v1.2.3",
			GitRepoURL:      "ssh://git@example.com/foo/t-buzz.git",
		},
	},
}

func TestNewCRDFromAPITenant(t *testing.T) {
	for name, test := range tenantTests {
		t.Run(name, func(t *testing.T) {
			apiTenant := Tenant{
				TenantId{
					Id: Id(fmt.Sprintf("t-%s", t.Name())),
				},
				test.properties,
			}
			tenant, err := NewCRDFromAPITenant(apiTenant)
			require.NoError(t, err)
			assert.Equal(t, test.spec.GitRepoURL, tenant.Spec.GitRepoURL)
			assert.Equal(t, test.spec.GitRepoRevision, tenant.Spec.GitRepoRevision)
			assert.Equal(t, test.spec.GlobalGitRepoURL, tenant.Spec.GlobalGitRepoURL)
			assert.Equal(t, test.spec.GlobalGitRepoRevision, tenant.Spec.GlobalGitRepoRevision)
		})
	}
}

func TestNewAPITenantFromCRD(t *testing.T) {
	for name, test := range tenantTests {
		t.Run(name, func(t *testing.T) {
			tenant := v1alpha1.Tenant{
				Spec: test.spec,
			}
			apiTenant := NewAPITenantFromCRD(tenant)
			if test.properties.GitRepo == nil {
				test.properties.GitRepo = &RevisionedGitRepo{}
			}
			assert.Equal(t, test.properties, apiTenant.TenantProperties)
		})
	}
}

var clusterTests = map[string]struct {
	properties ClusterProperties
	cluster    v1alpha1.Cluster
}{
	"empty": {
		ClusterProperties{},
		v1alpha1.Cluster{},
	},
	"global git revision": {
		ClusterProperties{
			GlobalGitRepoRevision: pointer.ToString("v1.2.3"),
		},
		v1alpha1.Cluster{
			Spec: v1alpha1.ClusterSpec{
				GlobalGitRepoRevision: "v1.2.3",
			},
		},
	},
	"tenant git revision": {
		ClusterProperties{
			TenantGitRepoRevision: pointer.ToString("v1.2.3"),
		},
		v1alpha1.Cluster{
			Spec: v1alpha1.ClusterSpec{
				TenantGitRepoRevision: "v1.2.3",
			},
		},
	},
}

func TestNewCRDFromAPICluster(t *testing.T) {
	for name, test := range clusterTests {
		t.Run(name, func(t *testing.T) {
			apiCluster := Cluster{
				ClusterId{
					Id: Id(fmt.Sprintf("c-%s", t.Name())),
				},
				ClusterTenant{fmt.Sprintf("t-%s", t.Name())},
				test.properties,
			}
			cluster, err := NewCRDFromAPICluster(apiCluster)
			assert.NoError(t, err)
			if len(test.cluster.Spec.TenantRef.Name) == 0 {
				test.cluster.Spec.TenantRef.Name = fmt.Sprintf("t-%s", t.Name())
			}
			assert.Equal(t, test.cluster.Spec, cluster.Spec)
			assert.Equal(t, test.cluster.Status.Facts, cluster.Status.Facts)
		})
	}
}

func TestNewAPIClusterFromCRD(t *testing.T) {
	for name, test := range clusterTests {
		t.Run(name, func(t *testing.T) {
			cluster := test.cluster
			apiCluster := NewAPIClusterFromCRD(cluster)
			if test.properties.GitRepo == nil {
				test.properties.GitRepo = &GitRepo{}
			}
			assert.Equal(t, test.properties, apiCluster.ClusterProperties)
		})
	}
}

func TestFactEncoding(t *testing.T) {
	facts := &DynamicClusterFacts{
		"kubernetesVersion": version.Info{
			Major:      "1",
			Minor:      "22",
			GitVersion: "1.22.14rc1",
		},
		"foo":  "bar",
		"buzz": `"bar"`,
	}

	apiCluster := Cluster{
		ClusterProperties: ClusterProperties{
			DynamicFacts: facts,
		},
	}
	cluster, err := NewCRDFromAPICluster(apiCluster)
	assert.NoError(t, err)
	apiCluster = *NewAPIClusterFromCRD(*cluster)

	act, err := json.Marshal(apiCluster.DynamicFacts)
	assert.NoError(t, err)
	exp, err := json.Marshal(facts)
	assert.NoError(t, err)
	assert.JSONEq(t, string(exp), string(act))
}

func TestDecodeFact(t *testing.T) {
	facts := []string{`"foo"`, "foo", `{"name": "bar"}`, "[1,2,3]"}
	decoded := []interface{}{}

	for _, f := range facts {
		require.NotPanics(t, func() {
			d := unmarshalFact(f)
			decoded = append(decoded, d)
		})
	}
	_, err := json.Marshal(decoded)
	assert.NoError(t, err)
}
