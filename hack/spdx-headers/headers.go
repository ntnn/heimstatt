// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package main

import (
	"path"
	"path/filepath"
	"strings"
)

// header is the header every file should carry.
var header = []string{
	"SPDX-License-Identifier: Apache-2.0",
	"SPDX-FileCopyrightText: The heimstatt Authors.",
}

// commentPrefixes map file types to the comment prefix.
var commentPrefixes = map[string]string{
	"dockerfile": "# ",
	"makefile":   "# ",
	"tiltfile":   "# ",
	".bash":      "# ",
	".hcl":       "# ",
	".sh":        "# ",
	".tf":        "# ",
	".toml":      "# ",
	".yaml":      "# ",
	".yml":       "# ",
	".zsh":       "# ",
}

// skipDirs are directories that will be skipped, e.g. because they only
// cache artifacts or contain generated output.
// A "*" matches one path segment, so one entry covers one directory per API group.
var skipDirs = []string{
	".git/*",
	"bin/*",
	"config/**/crds/*",
	"config/**/golden/*",
	"config/tilt/generated",
	"hack/tools/*",
}

// skipNames is skipDirs but for files.
var skipNames = []string{
	"LICENSE",
	"go.mod",
	"go.sum",
	"mindl.sum",
	"*.kubeconfig.yaml",
}

// classify takes a filepath rel and returns a string and a bool.
// The string is the comment prefix.
// The bool is whether the file's headers should be managed.
func classify(rel string) (string, bool) {
	base := strings.ToLower(filepath.Base(rel))
	ext := filepath.Ext(base)
	stem := strings.TrimSuffix(base, ext)

	if ext == ".go" {
		return "", false
	}
	if prefix, ok := commentPrefixes[ext]; ok {
		return prefix, true
	}
	prefix, ok := commentPrefixes[stem]
	return prefix, ok
}

// skipped reports if rel matches either skipDirs or skipNames.
func skipped(rel string) bool {
	slash := filepath.ToSlash(rel)
	if slash == "" || slash == "." {
		return false
	}
	for _, pattern := range skipDirs {
		match, err := path.Match(pattern, rel)
		if err != nil {
			panic(err)
		}
		if match {
			return true
		}
	}
	filename := filepath.Base(rel)
	for _, pattern := range skipNames {
		match, err := path.Match(pattern, filename)
		if err != nil {
			panic(err)
		}
		if match {
			return true
		}
	}
	return false
}
