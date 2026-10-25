// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStateOf(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		content string
		prefix  string
		want    state
	}{
		"header first": {
			content: "# SPDX-License-Identifier: Apache-2.0\n# SPDX-FileCopyrightText: The heimstatt Authors.\n\nname: x\n",
			prefix:  "# ",
			want:    stateMatch,
		},
		"header after another comment": {
			content: "# Build targets\n#\n# SPDX-License-Identifier: Apache-2.0\n# SPDX-FileCopyrightText: The heimstatt Authors.\n\nGO ?= go\n",
			prefix:  "# ",
			want:    stateMatch,
		},
		"header after shebang": {
			content: "#!/usr/bin/env sh\n# SPDX-License-Identifier: Apache-2.0\n# SPDX-FileCopyrightText: The heimstatt Authors.\n\necho hi\n",
			prefix:  "# ",
			want:    stateMatch,
		},
		"different spdx identifier": {
			content: "# SPDX-License-Identifier: MIT\n\nname: x\n",
			prefix:  "# ",
			want:    stateOther,
		},
		"header with a year": {
			content: "# SPDX-License-Identifier: Apache-2.0\n# SPDX-FileCopyrightText: 2026 The heimstatt Authors.\n\nname: x\n",
			prefix:  "# ",
			want:    stateOther,
		},
		"no comment at all": {
			content: "name: x\n",
			prefix:  "# ",
			want:    stateNone,
		},
		"spdx below the content": {
			content: "name: x\n# SPDX-License-Identifier: Apache-2.0\n",
			prefix:  "# ",
			want:    stateNone,
		},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()

			got := stateOf(strings.Split(cas.content, "\n"), cas.prefix)
			assert.Equal(t, cas.want, got)
		})
	}
}

func TestRun(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	write(t, root, "Makefile", "GO ?= go\nall:\n\t@echo hi\n", 0o644)
	write(t, root, "hack/run.sh", "#!/usr/bin/env sh\necho hi\n", 0o755)
	write(t, root, ".github/workflows/x.yaml", "# SPDX-License-Identifier: Apache-2.0\n# SPDX-FileCopyrightText: The heimstatt Authors.\n\nname: x\n", 0o644)
	write(t, root, "pkg/x.go", "package pkg\n", 0o644)
	write(t, root, "config/deploy/golden/render.yaml", "kind: ConfigMap\n", 0o644)
	write(t, root, "LICENSE", "Apache License\nVersion 2.0\n", 0o644)

	t.Run("Check reports the files this tool owns and nothing else", func(t *testing.T) {
		t.Parallel()

		err := run(Options{Root: root})
		require.Error(t, err, "two files are missing the header")
		assert.Contains(t, err.Error(), "Makefile")
		assert.Contains(t, err.Error(), "hack/run.sh")
		assert.NotContains(t, err.Error(), "pkg/x.go", "goheader owns .go files")
		assert.NotContains(t, err.Error(), "config/deploy/golden/render.yaml", "generated output is skipped")
		assert.NotContains(t, err.Error(), "LICENSE", "the license text carries the licensing itself")
	})

	t.Run("Fix writes the header and leaves foreign content alone", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, run(Options{Root: root, Fix: true}))

		assert.Equal(t,
			"# SPDX-License-Identifier: Apache-2.0\n# SPDX-FileCopyrightText: The heimstatt Authors.\n\nGO ?= go\nall:\n\t@echo hi\n",
			read(t, root, "Makefile"))
		assert.Equal(t,
			"#!/usr/bin/env sh\n# SPDX-License-Identifier: Apache-2.0\n# SPDX-FileCopyrightText: The heimstatt Authors.\n\necho hi\n",
			read(t, root, "hack/run.sh"), "the shebang stays the first line")
		assert.EqualValues(t, 0o755, readMode(t, root, "hack/run.sh").Perm(), "the script keeps its mode")
		assert.Equal(t, "package pkg\n", read(t, root, "pkg/x.go"))
		assert.Equal(t, "kind: ConfigMap\n", read(t, root, "config/deploy/golden/render.yaml"))

		require.NoError(t, run(Options{Root: root}), "a second check pass is clean")
	})
}

func write(t *testing.T, root, rel, content string, mode os.FileMode) {
	t.Helper()

	path := filepath.Join(root, filepath.FromSlash(rel))
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0o755))
	require.NoError(t, os.WriteFile(path, []byte(content), mode))
}

func read(t *testing.T, root, rel string) string {
	t.Helper()

	content, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
	require.NoError(t, err)
	return string(content)
}

func readMode(t *testing.T, root, rel string) os.FileMode {
	t.Helper()

	info, err := os.Stat(filepath.Join(root, filepath.FromSlash(rel)))
	require.NoError(t, err)
	return info.Mode()
}
