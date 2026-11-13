// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// GroupKind is the kind of Group.
const GroupKind = "Group"

// Group is one gitolite user group.
//
// Through the tenant export it is named after its workspace, so that
// "team" in "root:tenants:acme" is "@root.tenants.acme+team", and is
// usable by the rules of that workspace only.
//
// Through the admin export it is named "@<metadata.name>".
//
// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:validation:XValidation:rule="self.metadata.name != 'all'",message="all is the built-in group of gitolite",reason=FieldValueInvalid
type Group struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec GroupSpec `json:"spec"`

	// Status is what the operator observed of this group.
	Status GroupStatus `json:"status,omitempty"`
}

// GroupList is the collection kind for Group.
//
// +kubebuilder:object:root=true
type GroupList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Group `json:"items"`
}

// GroupSpec is the membership of one gitolite group.
type GroupSpec struct {
	// Members are usernames.
	// A group cannot be a member of a group.
	//
	// +listType=atomic
	// +kubebuilder:validation:items:Pattern=`^[0-9a-zA-Z][-0-9a-zA-Z._@+]*$`
	Members []string `json:"members,omitempty"`
}

// GroupStatus is the observed state of a Group.
type GroupStatus struct {
	// ConfName is the group name the compiler rendered.
	ConfName string `json:"confName,omitempty"`

	// ObservedConfigHash is the hash of the compiled configuration ConfName
	// came out of.
	ObservedConfigHash string `json:"observedConfigHash,omitempty"`

	// Conditions carry why the object is not projected, in detail.
	//
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
