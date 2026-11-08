// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package main

import (
	"os"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"
)

// TestMain re-executes this test binary as heimctl for the scripts.
func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"heimctl": main,
	})
}

func TestScript(t *testing.T) {
	testscript.Run(t, testscript.Params{
		Dir:                 "testdata/script",
		RequireExplicitExec: true,
		// UPDATE=1 go test ./cmd/heimctl -run TestScript rewrites the golden sections.
		UpdateScripts: os.Getenv("UPDATE") != "",
	})
}
