// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package heimkcp

import (
	"testing"

	operatorv1alpha1 "github.com/kcp-dev/kcp-operator/sdk/apis/operator/v1alpha1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"heimstatt.dev/apis/deploy/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

func newInstallation() *v1alpha1.Installation {
	return &v1alpha1.Installation{
		ObjectMeta: metav1.ObjectMeta{Name: "heimstatt", Namespace: "heimstatt"},
		Spec: v1alpha1.InstallationSpec{
			DomainBase: "heimstatt.127-0-0-1.sslip.io",
			OIDC: v1alpha1.OIDC{
				IssuerURL: "https://dex.heimstatt.127-0-0-1.sslip.io:30443/dex",
				ClientID:  "heimstatt-kcp",
				CAFileRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "heimstatt-serving"},
					Key:                  "ca.crt",
				},
			},
			Etcd: v1alpha1.EtcdConfig{
				Endpoints: []string{"http://etcd.heimstatt.svc.cluster.local:2379"},
			},
		},
	}
}

// etcdPrefixes maps every component with storage to the prefix it was given.
func etcdPrefixes(manifests *Manifests) map[string]string {
	prefixes := map[string]string{
		"cache":                  manifests.CacheServer.Spec.Etcd.Prefix,
		manifests.RootShard.Name: manifests.RootShard.Spec.Etcd.Prefix,
	}
	for _, shard := range manifests.Shards {
		prefixes[shard.Name] = shard.Spec.Etcd.Prefix
	}
	return prefixes
}

func TestRender(t *testing.T) {
	t.Parallel()

	tlsSecret := newInstallation()
	tlsSecret.Spec.Etcd.TLSSecretRef = &corev1.LocalObjectReference{Name: "etcd-client"}

	noNamespace := newInstallation()
	noNamespace.Namespace = ""

	noEndpoints := newInstallation()
	noEndpoints.Spec.Etcd.Endpoints = nil

	cases := map[string]struct {
		opts          Options
		wantErr       bool
		wantShards    []string
		wantPrefixes  map[string]string
		wantTLSSecret string
	}{
		"empty Shards renders the default topology": {
			opts:       Options{Installation: newInstallation()},
			wantShards: []string{"secondary"},
			wantPrefixes: map[string]string{
				"cache":     "/registry/cache",
				"root":      "/registry/kcp/root",
				"secondary": "/registry/kcp/secondary",
			},
		},
		"more shards render as Shard objects": {
			opts:       Options{Installation: newInstallation(), Shards: []string{"east", "root", "west"}},
			wantShards: []string{"east", "west"},
			wantPrefixes: map[string]string{
				"cache": "/registry/cache",
				"root":  "/registry/kcp/root",
				"east":  "/registry/kcp/east",
				"west":  "/registry/kcp/west",
			},
		},
		"the etcd TLS Secret reaches every component that stores data": {
			opts:          Options{Installation: tlsSecret},
			wantShards:    []string{"secondary"},
			wantTLSSecret: "etcd-client",
		},
		"an Installation without a namespace cannot be namespaced": {
			opts:    Options{Installation: noNamespace},
			wantErr: true,
		},
		"endpoints are required": {
			opts:    Options{Installation: noEndpoints},
			wantErr: true,
		},
		"one shard must be the root": {
			opts:    Options{Installation: newInstallation(), Shards: []string{"east", "west"}},
			wantErr: true,
		},
		"a shard name may not repeat": {
			opts:    Options{Installation: newInstallation(), Shards: []string{"root", "east", "east"}},
			wantErr: true,
		},
		"a shard name must be a label": {
			opts:    Options{Installation: newInstallation(), Shards: []string{"root", "East One"}},
			wantErr: true,
		},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()

			manifests, err := Render(cas.opts)
			if cas.wantErr {
				require.Error(t, err, "the render must report what is wrong with the Installation")
				return
			}
			require.NoError(t, err)
			require.NotNil(t, manifests.RootShard, "exactly one RootShard is rendered")

			var shardNames []string
			for _, shard := range manifests.Shards {
				shardNames = append(shardNames, shard.Name)
			}
			assert.Equal(t, cas.wantShards, shardNames)

			if len(cas.wantPrefixes) > 0 {
				assert.Equal(t, cas.wantPrefixes, etcdPrefixes(manifests), "components share one etcd and need distinct prefixes")
			}

			if cas.wantTLSSecret != "" {
				for _, config := range []operatorv1alpha1.EtcdConfig{
					*manifests.CacheServer.Spec.Etcd,
					manifests.RootShard.Spec.Etcd,
					manifests.Shards[0].Spec.Etcd,
				} {
					require.NotNil(t, config.TLSConfig)
					assert.Equal(t, cas.wantTLSSecret, config.TLSConfig.SecretRef.Name)
				}
			}
		})
	}
}

func TestManifests_Objects(t *testing.T) {
	t.Parallel()

	manifests, err := Render(Options{Installation: newInstallation()})
	require.NoError(t, err)

	var kinds []string
	for _, object := range manifests.Objects() {
		switch object.(type) {
		case *operatorv1alpha1.CacheServer:
			kinds = append(kinds, "CacheServer")
		case *operatorv1alpha1.RootShard:
			kinds = append(kinds, "RootShard")
		case *operatorv1alpha1.Shard:
			kinds = append(kinds, "Shard")
		case *operatorv1alpha1.FrontProxy:
			kinds = append(kinds, "FrontProxy")
		case *gwapiv1.TLSRoute:
			kinds = append(kinds, "TLSRoute")
		default:
			t.Errorf("unexpected object type %T", object)
		}
	}

	assert.Equal(t, []string{
		"CacheServer", "RootShard", "Shard", "FrontProxy",
		"TLSRoute", "TLSRoute", "TLSRoute",
	}, kinds, "objects are applied in dependency order")
}

func TestRootShardExternal(t *testing.T) {
	t.Parallel()

	manifests, err := Render(Options{Installation: newInstallation()})
	require.NoError(t, err)

	external := manifests.RootShard.Spec.External
	assert.Equal(t, "kcp.heimstatt.127-0-0-1.sslip.io", external.Hostname)
	assert.Equal(t, uint32(externalPort), external.Port, "clients reach kcp on the published port")
	assert.Equal(t, "heimstatt-front-proxy.heimstatt.svc.cluster.local", external.PrivateHostname,
		"pods must reach the front-proxy in-cluster, sslip.io answers 127.0.0.1 to them")
	assert.Equal(t, ptr.To(uint32(externalPort)), external.PrivatePort)

	assert.Equal(t, "root", manifests.RootShard.Name)
	assert.Equal(t, "heimstatt", manifests.CacheServer.Name)
	assert.Equal(t, "heimstatt", manifests.FrontProxy.Name)
	assert.Equal(t, []corev1.LocalObjectReference{{Name: "heimstatt"}},
		[]corev1.LocalObjectReference{*manifests.RootShard.Spec.Cache.Reference},
		"the root shard coordinates through the cache-server")
	assert.Equal(t, "root", manifests.Shards[0].Spec.RootShard.Reference.Name)

	assert.Equal(t,
		[]string{"root.kcp.heimstatt.127-0-0-1.sslip.io"},
		manifests.RootShard.Spec.CertificateTemplates[string(operatorv1alpha1.ServerCertificate)].Spec.DNSNames,
		"the shard hostname is added to the serving certificate, ShardBaseURL would break in-cluster URLs")
	assert.Equal(t,
		[]string{"secondary.kcp.heimstatt.127-0-0-1.sslip.io"},
		manifests.Shards[0].Spec.CertificateTemplates[string(operatorv1alpha1.ServerCertificate)].Spec.DNSNames)
}

func TestAuthSpec(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		oidc    v1alpha1.OIDC
		wantNil bool
		wantCA  *operatorv1alpha1.OIDCCAFileRef
	}{
		"full OIDC": {
			oidc: v1alpha1.OIDC{
				IssuerURL: "https://dex.example.com/dex",
				ClientID:  "heimstatt-kcp",
				CAFileRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "heimstatt-serving"},
					Key:                  "tls.crt",
				},
			},
			wantCA: &operatorv1alpha1.OIDCCAFileRef{Name: "heimstatt-serving", Key: "tls.crt"},
		},
		"a CA reference without a key defaults to ca.crt": {
			oidc: v1alpha1.OIDC{
				IssuerURL: "https://dex.example.com/dex",
				ClientID:  "heimstatt-kcp",
				CAFileRef: &corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "heimstatt-serving"},
				},
			},
			wantCA: &operatorv1alpha1.OIDCCAFileRef{Name: "heimstatt-serving", Key: "ca.crt"},
		},
		"no OIDC leaves authentication to the client certificates": {
			wantNil: true,
		},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()

			installation := newInstallation()
			installation.Spec.OIDC = cas.oidc

			auth := authSpec(installation)

			assert.True(t, auth.ServiceAccount.Enabled, "the front-proxy has to validate the shards' service account tokens")
			if cas.wantNil {
				assert.Nil(t, auth.OIDC)
				return
			}
			require.NotNil(t, auth.OIDC)
			assert.Equal(t, cas.oidc.IssuerURL, auth.OIDC.IssuerURL)
			assert.Equal(t, cas.oidc.ClientID, auth.OIDC.ClientID)
			assert.Equal(t, cas.wantCA, auth.OIDC.CAFileRef)
		})
	}
}
