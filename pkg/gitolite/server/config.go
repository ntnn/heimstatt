// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package server

import (
	gitolitev1alpha1 "heimstatt.dev/apis/gitolite/v1alpha1"
)

// Config is the whole configuration of a gitolite instance.
// Names within Groups and within Repositories are unique.
type Config struct {
	// Settings are the instance-wide knobs of the gitolite rc.
	Settings gitolitev1alpha1.ServerSpec

	// Groups cannot be members of groups.
	Groups []Group

	// Repositories are rendered in the order given, which is the order
	// gitolite matches their rules in.
	Repositories []Repository
}

// Group is a gitolite group of users or of repositories.
type Group struct {
	// Name is the group name without the leading "@".
	Name string

	// Members are users, repositories, or other groups with a leading "@".
	Members []string
}

// Repository is one "repo" block. Name may be a pattern or an "@group".
type Repository struct {
	Name string

	// Description is gitolite's "desc", which is the git config
	// gitweb.description and has to be allowed like any other.
	Description string

	// Rules are rendered as "-" when Denied, whatever Permission says.
	Rules []gitolitev1alpha1.RepositoryRule

	// Options are gitolite "option" lines keyed by option name.
	Options map[string]string

	// GitConfig are gitolite "config" lines keyed by git config name.
	GitConfig map[string]string
}
