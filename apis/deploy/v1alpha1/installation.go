// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Installation is one heimstatt deployment on one cluster.
//
// +kubebuilder:object:root=true
// +kubebuilder:resource:shortName=inst
type Installation struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec InstallationSpec `json:"spec"`
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

	// OIDC is the OpenID Connect issuer the heimstatt components trust.
	//
	// +kubebuilder:validation:Required
	OIDC OIDC `json:"oidc"`
}

// OIDC configures the OpenID Connect issuer the control plane authenticates against.
type OIDC struct {
	// IssuerURL is the issuer to authenticate against. Only https is accepted.
	IssuerURL string `json:"issuerURL,omitempty"`

	// ClientID is the OIDC client id..
	ClientID string `json:"clientID,omitempty"`

	// ClientSecretRef is the OIDC client secret.
	ClientSecretRef *corev1.SecretKeySelector `json:"clientSecret,omitempty"`

	// CAFileRef is the CA bundle for the OIDC provider.
	CAFileRef *corev1.SecretKeySelector `json:"caFileRef,omitempty"`
}
