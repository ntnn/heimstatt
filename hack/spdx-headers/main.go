// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

// Command spdx-headers checks that non-go files have the license header.
package main

import (
	"flag"
	"fmt"
	"os"
)

// Options configures a run.
type Options struct {
	Root string
	Fix  bool
	Log  func(format string, args ...any)
}

func main() {
	opts := Options{}
	verbose := false
	flag.StringVar(&opts.Root, "root", ".", "path to walk")
	flag.BoolVar(&opts.Fix, "fix", false, "add the header where it is missing")
	flag.BoolVar(&verbose, "v", false, "print every handled file")
	flag.Parse()
	if verbose {
		opts.Log = func(format string, args ...any) { fmt.Printf(format, args...) }
	}

	if err := run(opts); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
