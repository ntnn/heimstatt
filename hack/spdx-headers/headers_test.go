// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestClassify(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		rel     string
		prefix  string
		handled bool
	}{
		"makefile by name":       {rel: "Makefile", prefix: "# ", handled: true},
		"makefile with suffix":   {rel: "hack/Makefile.local", prefix: "# ", handled: true},
		"dockerfile by name":     {rel: "Dockerfile", prefix: "# ", handled: true},
		"tiltfile by name":       {rel: "Tiltfile", prefix: "# ", handled: true},
		"yaml":                   {rel: ".github/workflows/dco.yaml", prefix: "# ", handled: true},
		"shell":                  {rel: "hack/run.sh", prefix: "# ", handled: true},
		"go belongs to goheader": {rel: "pkg/render/render.go", handled: false},
		"markdown unknown":       {rel: "README.md", handled: false},
		"json unknown":           {rel: "sbom.json", handled: false},
		"no extension unknown":   {rel: "LICENSE", handled: false},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()

			prefix, handled := classify(cas.rel)
			assert.Equal(t, cas.handled, handled, "handled")
			assert.Equal(t, cas.prefix, prefix, "comment prefix")
		})
	}
}

func TestSkipped(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		rel     string
		skipped bool
	}{
		"pinned binaries":      {rel: "hack/tools/golangci-lint-2.15.0", skipped: true},
		"generated crds":       {rel: "config/deploy/crds/deploy.heimstatt.dev_installations.yaml", skipped: true},
		"golden render output": {rel: "config/deploy/golden/ca.yaml", skipped: true},
		"built binaries":       {rel: "bin/heimctl", skipped: true},
		"git internals":        {rel: ".git/config", skipped: true},
		"license file":         {rel: "LICENSE", skipped: true},
		"mindl sums":           {rel: "mindl.sum", skipped: true},
		"go module files":      {rel: "go.sum", skipped: true},
		"example config":       {rel: "config/examples/installation.yaml"},
		"group config source":  {rel: "config/deploy/kustomization.yaml"},
		"the tool itself":      {rel: "hack/spdx-headers/run.go"},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, cas.skipped, skipped(cas.rel))
		})
	}
}
