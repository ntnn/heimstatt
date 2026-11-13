// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package v1alpha1

// Condition types of the gitolite API, all in positive polarity.
const (
	// ConditionProjected is true while the object is part of the compiled
	// gitolite configuration.
	ConditionProjected = "Projected"

	// ConditionCompiled is true on the Server while gitolite compiled the
	// current configuration.
	ConditionCompiled = "Compiled"
)

// Finalizer keeps a Repository or Group until the configuration compiled
// without it and, for a Repository, its data is removed.
const Finalizer = "gitolite.heimstatt.dev/released"

// AnnotationRetainOnDelete set to "true" on a Repository keeps its data on
// disk when the object is deleted.
const AnnotationRetainOnDelete = "gitolite.heimstatt.dev/retain-on-delete"

// Reasons the control loop writes on the conditions above.
const (
	// ReasonPathUnknown is used when the path of the object's workspace is not
	// known. An unknown path must not degrade into the empty string, which
	// would put the object in every tenant's namespace.
	ReasonPathUnknown = "PathUnknown"

	// ReasonReservedName is used for a name the compiler generates itself.
	ReasonReservedName = "ReservedName"

	// ReasonDuplicateName is used when the qualified name of an object is
	// already taken.
	ReasonDuplicateName = "DuplicateName"

	// ReasonInvalid is used for a definition gitolite would not accept.
	ReasonInvalid = "Invalid"

	// ReasonDisallowedConfigKey is used for a git config name the instance's
	// GIT_CONFIG_KEYS does not allow.
	ReasonDisallowedConfigKey = "DisallowedConfigKey"

	// ReasonCompiled is used on an object the compiler rendered.
	ReasonCompiled = "Compiled"

	// ReasonCompileFailed is used while gitolite's compile of the whole
	// configuration fails, which says nothing about the object itself.
	ReasonCompileFailed = "CompileFailed"
)
