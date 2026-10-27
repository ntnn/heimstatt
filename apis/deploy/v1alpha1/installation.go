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

	// OIDC is the OpenID Connect issuer the heimstatt components trust.
	//
	// +kubebuilder:validation:Required
	OIDC OIDC `json:"oidc"`

	// Kcp is the kcp control plane to deploy.
	//
	// +kubebuilder:validation:Required
	Kcp Kcp `json:"kcp"`
}

// Kcp is the kcp control plane of one heimstatt deployment.
type Kcp struct {
	// Shards are the kcp shards making up the control plane. Exactly one must be named root.
	//
	// +listType=map
	// +listMapKey=name
	// +kubebuilder:validation:MinItems=1
	// +kubebuilder:validation:XValidation:rule="self.filter(s, s.name == 'root').size() == 1",message="exactly one shard must be named root"
	Shards []Shard `json:"shards"`
}

// Shard is one kcp shard: one apiserver backed by one etcd.
type Shard struct {
	// Name is the shard's name in kcp. It also picks the etcd prefix the shard stores its data under,
	// so it outlives the object: renaming a shard or dropping it from the list leaves its data in etcd,
	// and a new shard under the same name reads it back.
	// The name root is reserved for the shard that becomes the kcp root shard.
	//
	// +kubebuilder:validation:Pattern=`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`
	Name string `json:"name"`

	// Etcd is the storage this shard writes to.
	//
	// +kubebuilder:validation:Required
	Etcd EtcdConfig `json:"etcd"`
}

// EtcdConfig points at one etcd cluster.
type EtcdConfig struct {
	// Endpoints are etcd client URLs, e.g. https://etcd.heimstatt.svc.cluster.local:2379.
	// Several shards may share one etcd, they are kept apart by their prefixes.
	//
	// +kubebuilder:validation:MinItems=1
	Endpoints []string `json:"endpoints"`

	// TLSSecretRef is the Secret with the material to reach etcd over TLS: a CA bundle under ca.crt
	// and, if the etcd requires it, tls.crt and tls.key.
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

// OIDC configures the OpenID Connect issuer the control plane authenticates against.
type OIDC struct {
	// IssuerURL is the issuer to authenticate against. Only https is accepted.
	IssuerURL string `json:"issuerURL,omitempty"`

	// ClientID is the OIDC client id..
	ClientID string `json:"clientID,omitempty"`

	// ClientSecretRef is the OIDC client secret. Not handed to kcp: kube OIDC is secretless, and
	// kcp-operator deprecated the field. It stays for the provider-side client registration.
	ClientSecretRef *corev1.SecretKeySelector `json:"clientSecret,omitempty"`

	// CAFileRef is the CA bundle for the OIDC provider.
	CAFileRef *corev1.SecretKeySelector `json:"caFileRef,omitempty"`
}
