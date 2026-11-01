// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package main

import (
	"os"

	"heimstatt.dev/cmd/heimctl/app"
)

func main() {
	if err := app.NewCommand().Execute(); err != nil {
		os.Exit(1)
	}
}
