// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package server

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// requireValidity asserts that err is nil when valid and wraps sentinel otherwise.
func requireValidity(t *testing.T, valid bool, sentinel, err error) {
	t.Helper()

	if valid {
		require.NoError(t, err)
		return
	}
	require.ErrorIs(t, err, sentinel)
}

func TestValidateRepositoryName(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		name  string
		valid bool
	}{
		"plain":                 {"foo", true},
		"nested":                {"root.tenants.acme/a/b", true},
		"plus and at":           {"a+b@c", true},
		"dots inside a segment": {"a..b", true},
		"hidden segment":        {"a/.b", true},
		"group":                 {"@all", false},
		"pattern":               {"foo/[a-z]+", false},
		"wildcard":              {"foo/.*", false},
		"parent segment":        {"a/../b", false},
		"current segment":       {"a/./b", false},
		"empty segment":         {"a//b", false},
		"trailing slash":        {"a/", false},
		"leading slash":         {"/a", false},
		"leading dot":           {"../b", false},
		"leading dash":          {"-a", false},
		"git suffix":            {"a.git", false},
		"space":                 {"a b", false},
		"newline at the end":    {"a\n", false},
		"letter outside ASCII":  {"ä", false},
		"empty":                 {"", false},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()

			requireValidity(t, cas.valid, ErrInvalidName, ValidateRepositoryName(cas.name))
		})
	}
}

func TestValidateRepository(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		name  string
		valid bool
	}{
		"plain":                  {"foo/bar", true},
		"group":                  {"@all", true},
		"pattern":                {"foo/[a-z]+", true},
		"pattern from the start": {"[a-z]+/foo", true},
		"alternation":            {"foo/(a|b){1,2}$", true},
		"escaped dot":            {`foo\.bar`, true},
		"git suffix":             {"a.git", false},
		"two sigils":             {"@@all", false},
		"sigil only":             {"@", false},
		"leading caret":          {"^foo", false},
		"space":                  {"a b", false},
		"equals":                 {"a=b", false},
		"comment":                {"a#b", false},
		"newline":                {"a\nrepo @all", false},
		"empty":                  {"", false},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()

			requireValidity(t, cas.valid, ErrInvalidName, validateRepository(cas.name))
		})
	}
}

func TestValidateRef(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		ref   string
		valid bool
	}{
		"branch":          {"main", true},
		"qualified":       {"refs/tags/v[0-9]", true},
		"anchored branch": {"main$", true},
		"virtual ref":     {"VREF/NAME/docs/", true},
		"leading caret":   {"^main", true},
		"group":           {"@refs", false},
		"space":           {"a b", false},
		"equals":          {"a=b", false},
		"newline":         {"a\nb", false},
		"empty":           {"", false},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()

			requireValidity(t, cas.valid, ErrInvalidRef, validateRef(cas.ref))
		})
	}
}

func TestValidatePermission(t *testing.T) {
	t.Parallel()

	valid := []string{
		"-", "C", "R",
		"RW", "RWC", "RWD", "RWCD", "RWDC",
		"RWM", "RWCM", "RWDM", "RWCDM", "RWDCM",
		"RW+", "RW+C", "RW+D", "RW+CD", "RW+DC",
		"RW+M", "RW+CM", "RW+DM", "RW+CDM", "RW+DCM",
	}
	for _, permission := range valid {
		t.Run("valid "+permission, func(t *testing.T) {
			t.Parallel()

			assert.NoError(t, validatePermission(permission))
		})
	}

	invalid := []string{
		"", "W", "RWX", "rw", "R+", "RC", "-R", "RW++", "RWMM", "RWMC", "RWCC", "RWCDC", "RW +", "RW\n",
	}
	for _, permission := range invalid {
		t.Run("invalid "+permission, func(t *testing.T) {
			t.Parallel()

			assert.ErrorIs(t, validatePermission(permission), ErrInvalidPermission)
		})
	}
}

func TestValidateValue(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		value string
		valid bool
	}{
		"words":           {"what foo is for", true},
		"mail address":    {"acme@example.com", true},
		"equals":          {"a=b", true},
		"outside ASCII":   {"für alle", true},
		"empty":           {"", true},
		"comment":         {"a # b", false},
		"semicolon":       {"a;b", false},
		"backtick":        {"a`b", false},
		"parenthesis":     {"a (b)", false},
		"newline":         {"a\nrepo @all", false},
		"carriage return": {"a\rb", false},
		"tab":             {"a\tb", false},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()

			requireValidity(t, cas.valid, ErrInvalidValue, validateValue(cas.value))
		})
	}
}
