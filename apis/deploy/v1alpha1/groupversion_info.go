// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var (
	// GroupVersion is deploy.heimstatt.dev/v1alpha1.
	GroupVersion = schema.GroupVersion{Group: "deploy.heimstatt.dev", Version: "v1alpha1"}

	// SchemeBuilder registers this group's types.
	SchemeBuilder = runtime.NewSchemeBuilder(addKnownTypes)

	// AddToScheme registers this group's types into a scheme.
	AddToScheme = SchemeBuilder.AddToScheme
)

func addKnownTypes(scheme *runtime.Scheme) error {
	scheme.AddKnownTypes(GroupVersion,
		&Installation{},
		&InstallationList{},
	)
	metav1.AddToGroupVersion(scheme, GroupVersion)
	return nil
}
