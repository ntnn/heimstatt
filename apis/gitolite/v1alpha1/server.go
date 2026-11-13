// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ServerKind is the kind of Server.
const ServerKind = "Server"

// ServerName is the name of the singleton Server an instance reads.
const ServerName = "gitolite"

// Server is the configuration of the gitolite instance. It lives in the
// workspace of the operator and is not part of the exported API.
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster
// +kubebuilder:subresource:status
// +kubebuilder:validation:XValidation:rule="self.metadata.name == 'gitolite'",message="the only server is named gitolite",reason=FieldValueInvalid
type Server struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ServerSpec `json:"spec"`

	// Status is what the operator observed of the instance.
	Status ServerStatus `json:"status,omitempty"`
}

// ServerList is the collection kind for Server.
//
// +kubebuilder:object:root=true
type ServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Server `json:"items"`
}

// ServerSpec is the gitolite rc of the instance.
type ServerSpec struct {
	// Umask is gitolite's UMASK, in octal.
	//
	// +kubebuilder:validation:Pattern=`^0[0-7]{3}$`
	Umask string `json:"umask,omitempty"`

	// GitConfigKeys are the regexes of GIT_CONFIG_KEYS a repository's git
	// config names have to match. Empty turns per-repository git config off.
	//
	// +listType=atomic
	GitConfigKeys []string `json:"gitConfigKeys,omitempty"`

	// Roles are the role names of ROLES.
	//
	// +listType=set
	Roles []string `json:"roles,omitempty"`

	// OwnerRolename is gitolite's OWNER_ROLENAME.
	OwnerRolename string `json:"ownerRolename,omitempty"`

	// Enable are the commands and features of ENABLE.
	//
	// +listType=set
	Enable []string `json:"enable,omitempty"`

	// LocalCode is gitolite's LOCAL_CODE.
	LocalCode string `json:"localCode,omitempty"`

	// HTTPAnonUser is gitolite's HTTP_ANON_USER.
	HTTPAnonUser string `json:"httpAnonUser,omitempty"`

	// AdminUser is the user gitolite is set up with.
	AdminUser string `json:"adminUser,omitempty"`

	// RCExtra is perl appended to the rc as written.
	RCExtra string `json:"rcExtra,omitempty"`
}

// ServerStatus is the observed state of the instance.
type ServerStatus struct {
	// ObservedGeneration is the generation the conditions were computed from.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// ConfigHash is the hash of the configuration gitolite last compiled.
	ConfigHash string `json:"configHash,omitempty"`

	// Conditions carry why the configuration is not compiled, in detail.
	//
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
