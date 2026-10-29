// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Installation is one heimstatt deployment.
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=inst
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`
type Installation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InstallationSpec `json:"spec"`

	// Status is what heimstatt observed of the deployed components.
	Status InstallationStatus `json:"status,omitempty"`
}

// InstallationList is the collection kind for Installation.
//
// +kubebuilder:object:root=true
type InstallationList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`

	Items []Installation `json:"items"`
}

// InstallationSpec is the whole configuration of one heimstatt deployment.
type InstallationSpec struct {
	// DomainBase is the DNS suffix for heimstatt components.
	//
	// +kubebuilder:validation:MinLength=1
	DomainBase string `json:"domainBase"`

	// OIDC configures OpenID Connect Authentication.
	//
	// +kubebuilder:validation:Required
	OIDC OIDC `json:"oidc"`

	// Etcd is the storage this shard writes to.
	// The prefix is automatically the name.
	//
	// +kubebuilder:validation:Required
	Etcd EtcdConfig `json:"etcd"`
}

// OIDC configures OpenID Connect Authentication.
type OIDC struct {
	// IssuerURL is the issuer to authenticate against. Only https is accepted.
	IssuerURL string `json:"issuerURL,omitempty"`

	// ClientID is the OIDC client id.
	ClientID string `json:"clientID,omitempty"`

	// CAFileRef is the CA bundle for the OIDC provider.
	CAFileRef *corev1.SecretKeySelector `json:"caFileRef,omitempty"`
}

// EtcdConfig points at one etcd cluster.
type EtcdConfig struct {
	// Endpoints are etcd client URLs, e.g. https://etcd.heimstatt.svc.cluster.local:2379.
	//
	// +kubebuilder:validation:MinItems=1
	Endpoints []string `json:"endpoints"`

	// TLSSecretRef is the Secret with the material to reach etcd over TLS.
	// A CA bundle under ca.crt and, if the etcd requires it, tls.crt and tls.key.
	TLSSecretRef *corev1.LocalObjectReference `json:"tlsSecretRef,omitempty"`
}

// InstallationStatus is the observed state of one heimstatt deployment.
type InstallationStatus struct {
	// Phase summarizes whether the deployment is usable.
	//
	// +kubebuilder:validation:MinLength=1
	Phase string `json:"phase,omitempty"`

	// ObservedGeneration is the generation Status was computed for.
	ObservedGeneration int64 `json:"observedGeneration,omitempty"`

	// Conditions carry the detail behind Phase.
	//
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}
