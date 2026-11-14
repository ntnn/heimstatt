// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package server

import (
	"errors"
	"fmt"
	"strings"
)

// The rules below are those of gitolite.
const (
	groupSigil = "@"

	// allGroup is gitolite's built-in group and cannot be defined.
	allGroup = "all"

	pathSeparator = "/"

	gitSuffix = ".git"

	// nameCharacters is the punctuation gitolite takes in a name,
	// next to letters and digits.
	nameCharacters = "-._@+"

	// patternCharacters is the punctuation gitolite takes in a pattern,
	// next to what it takes in a name.
	patternCharacters = `/\^$|()[]*?{},`

	// patternStart is what a pattern may start with,
	// next to a letter or a digit.
	patternStart = "["

	// keyCharacters is the punctuation taken in an option or git config name,
	// next to letters and digits.
	keyCharacters = "-_."

	// unsafeCharacters fail gitolite's compile when a value contains one.
	unsafeCharacters = "`~#$&()|;<>"

	branchPrefix = "refs/heads/"
)

// ValidateRepositoryName checks that name is one literal repository,
// not a pattern and not an "@group".
func ValidateRepositoryName(name string) error {
	if strings.HasPrefix(name, groupSigil) {
		return fmt.Errorf("%w: %q is a group", ErrInvalidName, name)
	}

	if strings.HasSuffix(name, gitSuffix) {
		return fmt.Errorf("%w: %q ends in %q", ErrInvalidName, name, gitSuffix)
	}

	if err := checkCharacters(name, "", nameCharacters+pathSeparator); err != nil {
		return fmt.Errorf("%w: %q %w", ErrInvalidName, name, err)
	}

	for segment := range strings.SplitSeq(name, pathSeparator) {
		switch segment {
		case "":
			return fmt.Errorf("%w: %q has an empty path segment", ErrInvalidName, name)
		case ".", "..":
			return fmt.Errorf("%w: %q has the path segment %q", ErrInvalidName, name, segment)
		}
	}
	return nil
}

// validateRepository checks that name is what a "repo" line takes:
// a repository, a pattern or an "@group".
func validateRepository(name string) error {
	// gitolite strips the suffix, which would render a name nobody asked for.
	if strings.HasSuffix(name, gitSuffix) {
		return fmt.Errorf("%w: %q ends in %q", ErrInvalidName, name, gitSuffix)
	}
	if err := checkPattern(name); err != nil {
		return fmt.Errorf("%w: %q %w", ErrInvalidName, name, err)
	}
	return nil
}

// validateRef checks a refex the way gitolite does, after qualifying it.
func validateRef(ref string) error {
	// gitolite expands a leading "@" as a group and dies on an undefined one.
	if strings.HasPrefix(ref, groupSigil) {
		return fmt.Errorf("%w: %q names a group", ErrInvalidRef, ref)
	}
	if ref == "" {
		return fmt.Errorf("%w: %q is empty", ErrInvalidRef, ref)
	}
	qualified := ref
	if !strings.HasPrefix(ref, "refs/") && !strings.HasPrefix(ref, "VREF/") {
		qualified = branchPrefix + ref
	}
	if err := checkPattern(qualified); err != nil {
		return fmt.Errorf("%w: %q %w", ErrInvalidRef, ref, err)
	}
	return nil
}

// validatePermission checks a permission: "-", "C", "R",
// or "RW" followed by an optional "+", "C" and "D" in either order, and "M".
func validatePermission(permission string) error {
	switch permission {
	case "-", "C", "R":
		return nil
	}
	rest, isWrite := strings.CutPrefix(permission, "RW")
	if !isWrite {
		return fmt.Errorf("%w: %q", ErrInvalidPermission, permission)
	}
	rest = strings.TrimPrefix(rest, "+")
	rest = strings.TrimSuffix(rest, "M")
	switch rest {
	case "", "C", "D", "CD", "DC":
		return nil
	}
	return fmt.Errorf("%w: %q", ErrInvalidPermission, permission)
}

// validateValue checks the right side of an option, config or desc line.
func validateValue(value string) error {
	if i := strings.IndexAny(value, unsafeCharacters); i >= 0 {
		return fmt.Errorf("%w: contains %q", ErrInvalidValue, value[i])
	}
	for _, r := range value {
		if r < ' ' || r == 0x7f {
			return fmt.Errorf("%w: contains the control character %q", ErrInvalidValue, r)
		}
	}
	return nil
}

// checkPattern checks a repository pattern, which may be an "@group".
func checkPattern(pattern string) error {
	name := strings.TrimPrefix(pattern, groupSigil)
	return checkCharacters(name, patternStart, nameCharacters+patternCharacters)
}

// checkCharacters checks that s starts with a letter, a digit or one of start
// and continues with letters, digits and punctuation only.
func checkCharacters(s, start, punctuation string) error {
	if s == "" {
		return errors.New("is empty")
	}
	for i, r := range s {
		if isAlphanumeric(r) {
			continue
		}
		if i == 0 && !strings.ContainsRune(start, r) {
			return fmt.Errorf("starts with %q", r)
		}
		if i > 0 && !strings.ContainsRune(punctuation, r) {
			return fmt.Errorf("contains %q", r)
		}
	}
	return nil
}

func isAlphanumeric(r rune) bool {
	return r >= '0' && r <= '9' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z'
}
