// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package server

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	gitolitev1alpha1 "heimstatt.dev/apis/gitolite/v1alpha1"
)

// updateGoldenEnv rewrites the golden files instead of comparing with them.
const updateGoldenEnv = "UPDATE_GOLDEN"

// okConf is what okRepository renders as.
const okConf = "repo ok\n    RW+ = alice\n"

// okRepository is the definition every case expects to survive its neighbours.
func okRepository() Repository {
	return Repository{
		Name: "ok",
		Rules: []gitolitev1alpha1.RepositoryRule{
			{Permission: "RW+", Subjects: []string{"alice"}},
		},
	}
}

func repositoryWithRule(rule gitolitev1alpha1.RepositoryRule) Repository {
	return Repository{
		Name:  "bad",
		Rules: []gitolitev1alpha1.RepositoryRule{rule},
	}
}

func TestRender(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		config               Config
		conf                 string
		rejectedGroups       map[string]error
		rejectedRepositories map[string]error
	}{
		"empty config": {
			config: Config{},
			conf:   "",
		},
		"repositories keep the order given": {
			config: Config{
				Repositories: []Repository{
					{Name: "zeta"},
					{Name: "@all"},
					{Name: "alpha"},
				},
			},
			conf: "repo zeta\n\nrepo @all\n\nrepo alpha\n",
		},
		"rule with refs and a deny": {
			config: Config{
				Repositories: []Repository{
					{
						Name: "foo",
						Rules: []gitolitev1alpha1.RepositoryRule{
							{Permission: "RW", Refs: []string{"main", "refs/tags/v[0-9]"}, Subjects: []string{"alice", "@devs"}},
							{Permission: "RW+", Denied: true, Subjects: []string{"eve"}},
						},
					},
				},
			},
			conf: "repo foo\n    RW main refs/tags/v[0-9] = alice @devs\n    - = eve\n",
		},
		"options, description and git config as git config sorted by key": {
			config: Config{
				Settings: gitolitev1alpha1.ServerSpec{GitConfigKeys: []string{`hooks\..*`, `gitweb\.description`}},
				Repositories: []Repository{
					{
						Name:        "foo",
						Description: "what foo is for",
						Options:     map[string]string{"deny-rules": "1", "a": "b"},
						GitConfig:   map[string]string{"hooks.z": "1", "Hooks.a": ""},
					},
				},
			},
			conf: "repo foo\n" +
				"    config Hooks.a =\n" +
				"    config gitolite-options.a = b\n" +
				"    config gitolite-options.deny-rules = 1\n" +
				"    config gitweb.description = what foo is for\n" +
				"    config hooks.z = 1\n",
		},
		"group without members left out": {
			config: Config{
				Groups: []Group{
					{Name: "nobody"},
					{Name: "ok", Members: []string{"alice", "foo/[a-z]+"}},
				},
			},
			conf: "@ok = alice foo/[a-z]+\n",
		},
		"repository name with a space": {
			config: Config{
				Repositories: []Repository{{Name: "a b"}, okRepository()},
			},
			conf:                 okConf,
			rejectedRepositories: map[string]error{"a b": ErrInvalidName},
		},
		"repository name with a newline": {
			config: Config{
				Repositories: []Repository{{Name: "a\nrepo @all"}, okRepository()},
			},
			conf:                 okConf,
			rejectedRepositories: map[string]error{"a\nrepo @all": ErrInvalidName},
		},
		"repository name with the git suffix": {
			config: Config{
				Repositories: []Repository{{Name: "a.git"}, okRepository()},
			},
			conf:                 okConf,
			rejectedRepositories: map[string]error{"a.git": ErrInvalidName},
		},
		"unknown permission": {
			config: Config{
				Repositories: []Repository{
					repositoryWithRule(gitolitev1alpha1.RepositoryRule{Permission: "RWX", Subjects: []string{"alice"}}),
					okRepository(),
				},
			},
			conf:                 okConf,
			rejectedRepositories: map[string]error{"bad": ErrInvalidPermission},
		},
		"rule without subjects": {
			config: Config{
				Repositories: []Repository{
					repositoryWithRule(gitolitev1alpha1.RepositoryRule{Permission: "R"}),
					okRepository(),
				},
			},
			conf:                 okConf,
			rejectedRepositories: map[string]error{"bad": ErrInvalidName},
		},
		"subject with a slash": {
			config: Config{
				Repositories: []Repository{
					repositoryWithRule(gitolitev1alpha1.RepositoryRule{Permission: "R", Subjects: []string{"a/b"}}),
					okRepository(),
				},
			},
			conf:                 okConf,
			rejectedRepositories: map[string]error{"bad": ErrInvalidName},
		},
		"ref with a space": {
			config: Config{
				Repositories: []Repository{
					repositoryWithRule(gitolitev1alpha1.RepositoryRule{Permission: "RW", Refs: []string{"a b"}, Subjects: []string{"alice"}}),
					okRepository(),
				},
			},
			conf:                 okConf,
			rejectedRepositories: map[string]error{"bad": ErrInvalidRef},
		},
		"ref naming a group": {
			config: Config{
				Repositories: []Repository{
					repositoryWithRule(gitolitev1alpha1.RepositoryRule{Permission: "RW", Refs: []string{"@refs"}, Subjects: []string{"alice"}}),
					okRepository(),
				},
			},
			conf:                 okConf,
			rejectedRepositories: map[string]error{"bad": ErrInvalidRef},
		},
		"git config key not allowed": {
			config: Config{
				Settings: gitolitev1alpha1.ServerSpec{GitConfigKeys: []string{`hooks\..*`}},
				Repositories: []Repository{
					{Name: "bad", GitConfig: map[string]string{"core.sshCommand": "x"}},
					okRepository(),
				},
			},
			conf:                 okConf,
			rejectedRepositories: map[string]error{"bad": ErrDisallowedConfigKey},
		},
		"git config key matching only a prefix": {
			config: Config{
				Settings: gitolitev1alpha1.ServerSpec{GitConfigKeys: []string{`hooks`}},
				Repositories: []Repository{
					{Name: "bad", GitConfig: map[string]string{"hooks.x": "1"}},
					okRepository(),
				},
			},
			conf:                 okConf,
			rejectedRepositories: map[string]error{"bad": ErrDisallowedConfigKey},
		},
		"git config without any allowed key": {
			config: Config{
				Repositories: []Repository{
					{Name: "bad", GitConfig: map[string]string{"hooks.x": "1"}},
					okRepository(),
				},
			},
			conf:                 okConf,
			rejectedRepositories: map[string]error{"bad": ErrDisallowedConfigKey},
		},
		"description without gitweb.description allowed": {
			config: Config{
				Repositories: []Repository{
					{Name: "bad", Description: "text"},
					okRepository(),
				},
			},
			conf:                 okConf,
			rejectedRepositories: map[string]error{"bad": ErrDisallowedConfigKey},
		},
		"git config key with a space": {
			config: Config{
				Settings: gitolitev1alpha1.ServerSpec{GitConfigKeys: []string{`.*`}},
				Repositories: []Repository{
					{Name: "bad", GitConfig: map[string]string{"a b": "1"}},
					okRepository(),
				},
			},
			conf:                 okConf,
			rejectedRepositories: map[string]error{"bad": ErrInvalidName},
		},
		"git config value with a newline": {
			config: Config{
				Settings: gitolitev1alpha1.ServerSpec{GitConfigKeys: []string{`.*`}},
				Repositories: []Repository{
					{Name: "bad", GitConfig: map[string]string{"hooks.x": "1\nrepo @all"}},
					okRepository(),
				},
			},
			conf:                 okConf,
			rejectedRepositories: map[string]error{"bad": ErrInvalidValue},
		},
		"option value with a shell character": {
			config: Config{
				Repositories: []Repository{
					{Name: "bad", Options: map[string]string{"x": "a;b"}},
					okRepository(),
				},
			},
			conf:                 okConf,
			rejectedRepositories: map[string]error{"bad": ErrInvalidValue},
		},
		"description with a comment character": {
			config: Config{
				Settings: gitolitev1alpha1.ServerSpec{GitConfigKeys: []string{`gitweb\..*`}},
				Repositories: []Repository{
					{Name: "bad", Description: "a # b"},
					okRepository(),
				},
			},
			conf:                 okConf,
			rejectedRepositories: map[string]error{"bad": ErrInvalidValue},
		},
		"group named all": {
			config: Config{
				Groups:       []Group{{Name: "all", Members: []string{"alice"}}},
				Repositories: []Repository{okRepository()},
			},
			conf:           okConf,
			rejectedGroups: map[string]error{"all": ErrInvalidName},
		},
		"group name with a slash": {
			config: Config{
				Groups: []Group{
					{Name: "a/b", Members: []string{"alice"}},
					{Name: "ok", Members: []string{"alice"}},
				},
			},
			conf:           "@ok = alice\n",
			rejectedGroups: map[string]error{"a/b": ErrInvalidName},
		},
		"group member with a space": {
			config: Config{
				Groups: []Group{
					{Name: "bad", Members: []string{"a b"}},
					{Name: "ok", Members: []string{"alice"}},
				},
			},
			conf:           "@ok = alice\n",
			rejectedGroups: map[string]error{"bad": ErrInvalidName},
		},
		"group naming a group": {
			config: Config{
				Groups: []Group{
					{Name: "bad", Members: []string{"@ok"}},
					{Name: "ok", Members: []string{"alice"}},
				},
			},
			conf:           "@ok = alice\n",
			rejectedGroups: map[string]error{"bad": ErrInvalidName},
		},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()

			rendered, err := Render(cas.config)
			require.NoError(t, err)

			assert.Equal(t, cas.conf, rendered.Conf)
			assert.Equal(t, hash(rendered.Conf, rendered.RC), rendered.Hash)

			require.Len(t, rendered.RejectedGroups, len(cas.rejectedGroups))
			for name, expected := range cas.rejectedGroups {
				require.ErrorIs(t, rendered.RejectedGroups[name], expected, "group %q", name)
			}
			require.Len(t, rendered.RejectedRepositories, len(cas.rejectedRepositories))
			for name, expected := range cas.rejectedRepositories {
				require.ErrorIs(t, rendered.RejectedRepositories[name], expected, "repository %q", name)
			}
		})
	}
}

func TestRenderConfigErrors(t *testing.T) {
	t.Parallel()

	t.Run("git config key pattern that does not compile", func(t *testing.T) {
		t.Parallel()

		_, err := Render(Config{
			Settings: gitolitev1alpha1.ServerSpec{GitConfigKeys: []string{"("}},
		})
		require.Error(t, err)
	})
}

func TestRenderHash(t *testing.T) {
	t.Parallel()

	first, err := Render(Config{Repositories: []Repository{okRepository()}})
	require.NoError(t, err)
	again, err := Render(Config{Repositories: []Repository{okRepository()}})
	require.NoError(t, err)
	other, err := Render(Config{Repositories: []Repository{{Name: "other"}}})
	require.NoError(t, err)

	assert.Len(t, first.Hash, hashLength)
	assert.Equal(t, first.Hash, again.Hash, "the same config hashes the same")
	assert.NotEqual(t, first.Hash, other.Hash, "a different conf hashes differently")
}

func TestRenderGolden(t *testing.T) {
	t.Parallel()

	config := Config{
		Settings: gitolitev1alpha1.ServerSpec{
			GitConfigKeys: []string{`hooks\..*`, `gitweb\.(owner|description|category)`},
		},
		Groups: []Group{
			{Name: "root.tenants.acme+all", Members: []string{"root.tenants.acme/app", "root.tenants.acme/lib/core"}},
			{Name: "root.tenants.acme+maintainers", Members: []string{"alice", "carol"}},
			{Name: "root.tenants.acme+devs", Members: []string{"alice", "bob"}},
			{Name: "staff", Members: []string{"dave"}},
		},
		Repositories: []Repository{
			{
				Name: "@all",
				Rules: []gitolitev1alpha1.RepositoryRule{
					{Permission: "R", Subjects: []string{"@staff"}},
				},
			},
			{
				Name: "sandbox/[a-z]+",
				Rules: []gitolitev1alpha1.RepositoryRule{
					{Permission: "C", Subjects: []string{"@staff"}},
					{Permission: "RW+", Subjects: []string{"CREATOR"}},
				},
			},
			{
				Name:        "root.tenants.acme/app",
				Description: "the acme application",
				Rules: []gitolitev1alpha1.RepositoryRule{
					{Permission: "RW+", Subjects: []string{"@root.tenants.acme+maintainers"}},
					{Permission: "RW", Refs: []string{"main"}, Denied: true, Subjects: []string{"@root.tenants.acme+devs"}},
					{Permission: "RWCD", Refs: []string{"refs/heads/feature/"}, Subjects: []string{"@root.tenants.acme+devs"}},
					{Permission: "R", Subjects: []string{"@all"}},
				},
				Options:   map[string]string{"deny-rules": "1"},
				GitConfig: map[string]string{"hooks.mailinglist": "acme@example.com"},
			},
			{
				Name: "root.tenants.acme/lib/core",
				Rules: []gitolitev1alpha1.RepositoryRule{
					{Permission: "RW", Subjects: []string{"@root.tenants.acme+devs"}},
				},
			},
		},
	}

	rendered, err := Render(config)
	require.NoError(t, err)
	require.Empty(t, rendered.RejectedGroups)
	require.Empty(t, rendered.RejectedRepositories)

	path := filepath.Join("testdata", "golden", "full.conf")
	if os.Getenv(updateGoldenEnv) != "" {
		require.NoError(t, os.WriteFile(path, []byte(rendered.Conf), 0o600))
	}

	expected, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, string(expected), rendered.Conf)
}
