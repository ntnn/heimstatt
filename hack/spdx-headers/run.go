// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// scanLimit is how far down a file the header may start.
const scanLimit = 20

// state is what a file says about the configured header.
type state string

const (
	stateMatch state = "match"

	// stateOther means an SPDX identifier is present but not the configured header.
	// other is not rewritten when fixing, that is likely a deliberate choice.
	stateOther state = "other"

	stateNone state = "none"
)

func run(opts Options) error {
	var files []string
	err := filepath.WalkDir(opts.Root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(opts.Root, path)
		if err != nil {
			return fmt.Errorf("relative path of %s: %w", path, err)
		}
		if skipped(rel) {
			if entry.IsDir() {
				return fs.SkipDir
			}
			return nil
		}
		if entry.IsDir() || !handled(rel) {
			return nil
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return fmt.Errorf("walking %s: %w", opts.Root, err)
	}

	var problems []string
	for _, path := range files {
		rel := display(opts.Root, path)
		prefix := prefixOf(rel)
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("reading %s: %w", rel, err)
		}
		switch stateOf(strings.Split(string(content), "\n"), prefix) {
		case stateMatch:
			logf(opts, "ok     %s\n", rel)
		case stateOther:
			problems = append(problems, rel+": SPDX identifier is not the configured header")
		case stateNone:
			if !opts.Fix {
				problems = append(problems, rel+": missing SPDX header")
				continue
			}
			if err := rewrite(path, content, prefix); err != nil {
				return fmt.Errorf("adding header to %s: %w", rel, err)
			}
			logf(opts, "fixed  %s\n", rel)
		}
	}

	if len(problems) > 0 {
		return fmt.Errorf("SPDX header problems:\n%s", strings.Join(problems, "\n"))
	}
	return nil
}

// handled reports whether this tool writes headers for rel.
func handled(rel string) bool {
	_, ok := classify(rel)
	return ok
}

func prefixOf(rel string) string {
	prefix, _ := classify(rel)
	return prefix
}

func display(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

// stateOf classifies the comment block at the top of a file against the configured header.
func stateOf(lines []string, prefix string) state {
	region := headerRegion(lines, prefix)
	if containsRun(region, header) {
		return stateMatch
	}
	for _, line := range region {
		if strings.Contains(line, "SPDX-License-Identifier") {
			return stateOther
		}
	}
	return stateNone
}

// headerRegion collects the leading comment lines with their marker stripped,
// stopping at the first line that is neither blank nor a comment.
func headerRegion(lines []string, prefix string) []string {
	marker := strings.TrimRight(prefix, " ")
	var region []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
			continue
		case strings.HasPrefix(trimmed, "#!"):
			continue
		case !strings.HasPrefix(trimmed, marker):
			return region
		default:
			region = append(region, strings.TrimSpace(strings.TrimPrefix(trimmed, marker)))
			if len(region) >= scanLimit {
				return region
			}
		}
	}
	return region
}

func containsRun(lines, want []string) bool {
	if len(lines) < len(want) {
		return false
	}
	for start := 0; start+len(want) <= len(lines); start++ {
		matched := true
		for i, want := range want {
			if lines[start+i] != want {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

// rewrite adds the header after a shebang line, keeping the file's own mode.
func rewrite(path string, content []byte, prefix string) error {
	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("reading file mode: %w", err)
	}
	lines := strings.Split(string(content), "\n")
	var shebang string
	if len(lines) > 0 && strings.HasPrefix(strings.TrimSpace(lines[0]), "#!") {
		shebang, lines = lines[0], lines[1:]
	}
	out := make([]string, 0, len(header)+len(lines)+2)
	if shebang != "" {
		out = append(out, shebang)
	}
	for _, line := range header {
		out = append(out, prefix+line)
	}
	out = append(out, "")
	out = append(out, lines...)
	if err := os.WriteFile(path, []byte(strings.Join(out, "\n")), info.Mode()); err != nil {
		return fmt.Errorf("writing file: %w", err)
	}
	return nil
}

func logf(opts Options, format string, args ...any) {
	if opts.Log == nil {
		return
	}
	opts.Log(format, args...)
}
