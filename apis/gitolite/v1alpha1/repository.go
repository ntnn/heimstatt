// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// RepositoryKind is the kind of Repository.
const RepositoryKind = "Repository"

// Repository is one gitolite repository.
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
type Repository struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec RepositorySpec `json:"spec"`

	// Status is what the operator observed of this repository.
	Status RepositoryStatus `json:"status,omitempty"`
}

// RepositoryList is the collection kind for Repository.
//
// +kubebuilder:object:root=true
type RepositoryList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Repository `json:"items"`
}

// RepositorySpec is the gitolite configuration of one repository.
type RepositorySpec struct {
	// Name is the repository name and may contain slashes.
	//
	// Through the tenant export it renders below the workspace path, so
	// that "a/b" in "root:tenants:acme" is "root.tenants.acme/a/b".
	//
	// Through the admin export it renders as written.
	//
	// +kubebuilder:validation:MinLength=1
	// +kubebuilder:validation:Pattern=`^@?[\[0-9a-zA-Z][-0-9a-zA-Z._@/+\\^$|()\[\]*?{},]*$`
	// +kubebuilder:validation:XValidation:rule="!self.endsWith('.git')",message="the name is given without the .git suffix"
	Name string `json:"name"`

	// Description is free text about what the repository is for.
	//
	// +kubebuilder:validation:Pattern="^[^`~#$&()|;<>[:cntrl:]]*$"
	Description string `json:"description,omitempty"`

	// Rules are the access rules in gitolite's order, first match wins.
	//
	// +kubebuilder:validation:MinItems=1
	Rules []RepositoryRule `json:"rules"`

	// Options are gitolite "option" lines keyed by option name.
	//
	// +kubebuilder:validation:XValidation:rule="self.all(key, key.matches('^[0-9a-zA-Z][-0-9a-zA-Z_.]*$'))",message="option names are letters, digits, dashes, underscores and dots"
	Options map[string]RepositoryConfigValue `json:"options,omitempty"`

	// GitConfig are gitolite "config" lines keyed by git config name.
	//
	// +kubebuilder:validation:XValidation:rule="self.all(key, key.matches('^[0-9a-zA-Z][-0-9a-zA-Z_.]*$'))",message="git config names are letters, digits, dashes, underscores and dots"
	GitConfig map[string]RepositoryConfigValue `json:"gitConfig,omitempty"`
}

// RepositoryConfigValue is the value of an option or a git config.
//
// +kubebuilder:validation:Pattern="^[^`~#$&()|;<>[:cntrl:]]*$"
type RepositoryConfigValue string

// RepositoryRule is one gitolite access rule.
type RepositoryRule struct {
	// Permission is the gitolite permission, one of "-", "R", "RW", "RW+",
	// and their "C", "D", "M" additions.
	//
	// +kubebuilder:validation:Enum="-";"C";"R";"RW";"RWC";"RWD";"RWCD";"RWDC";"RWM";"RWCM";"RWDM";"RWCDM";"RWDCM";"RW+";"RW+C";"RW+D";"RW+CD";"RW+DC";"RW+M";"RW+CM";"RW+DM";"RW+CDM";"RW+DCM"
	Permission string `json:"permission"`

	// Refs restricts the rule to refname regexes; empty means every ref.
	//
	// +kubebuilder:validation:items:Pattern=`^[-0-9a-zA-Z._/+\\^$|()\[\]*?{},]+$`
	Refs []string `json:"refs,omitempty"`

	// Subjects are gitolite users, groups and UMLs.
	//
	// Through the tenant export a group named here is always the
	// caller's: "@team" is rewritten to "@root.tenants.acme+team" and
	// "@all" to the group of the caller's own repositories.
	//
	// Through the admin export subjects render as written.
	//
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:items:Pattern=`^@?[0-9a-zA-Z][-0-9a-zA-Z._@+]*$`
	Subjects []string `json:"subjects"`

	// Denied makes the rule a deny rather than a grant.
	Denied bool `json:"denied,omitempty"`
}

// RepositoryStatus is the observed state of a Repository.
type RepositoryStatus struct {
	// Path is the qualified name the compiler rendered.
	Path string `json:"path,omitempty"`

	// CloneURL is the address to clone the repository from.
	CloneURL string `json:"cloneURL,omitempty"`

	// ObservedConfigHash is the hash of the compiled configuration Path came
	// out of.
	ObservedConfigHash string `json:"observedConfigHash,omitempty"`

	// Conditions carry why the object is not projected, in detail.
	//
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
