// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package heimkcp

import (
	"errors"
	"fmt"
	"slices"
	"strings"

	operatorv1alpha1 "github.com/kcp-dev/kcp-operator/sdk/apis/operator/v1alpha1"
	deployv1alpha1 "heimstatt.dev/apis/deploy/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/validation"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"
)

// externalPort is where kcp answers from outside the cluster. It is the port the
// front-proxy Service listens on and the NodePort the ingress publishes it under,
// mirrored in config/tilt/kind.yaml and config/tilt/components/gateway.yaml.
const externalPort = 30444

// shardPort is the port the kcp apiserver Services answer on in-cluster.
const shardPort = 6443

// hostLabel is the DNS label kcp owns under spec.domainBase. The front-proxy gets
// <hostLabel>.<domainBase>, every shard <shard>.<hostLabel>.<domainBase>.
const hostLabel = "kcp"

// issuerName is the cert-manager Issuer that signs kcp's own PKI. It is created in
// the Installation's namespace by config/tilt/components/gateway.yaml.
const issuerName = "selfsigned"

// clusterDomain is the in-cluster DNS suffix of the Services the operator creates.
// It matches the ClusterDomain the operator defaults to when the field is left empty.
const clusterDomain = "cluster.local"

// rootShardName names the shard kcp makes its root: only a RootShard owns the PKI
// and is referenced by the front-proxy and the other shards.
const rootShardName = "root"

// shardEtcdPrefix prefixes a shard's data in the shared etcd. Left empty, every
// shard would fall back to kcp's own default prefix and they would collide.
const shardEtcdPrefix = "/registry/kcp"

// cacheEtcdPrefix prefixes the cache-server's data in the shared etcd.
const cacheEtcdPrefix = "/registry/cache"

// The kcp-operator names the Service of each object it creates after that object.
// Routes and in-cluster URLs are built from those names.
const (
	rootShardServiceSuffix  = "-kcp"
	shardServiceSuffix      = "-shard-kcp"
	frontProxyServiceSuffix = "-front-proxy"
)

// defaultShards is the topology of a single-cluster installation.
// Shards exists as an Option so a topology can vary before the CRD can.
func defaultShards() []string {
	return []string{rootShardName, "secondary"}
}

// Options configures one render of an Installation.
type Options struct {
	Installation *deployv1alpha1.Installation

	// Shards names the kcp shards to deploy, exactly one of them rootShardName.
	// Empty means defaultShards.
	Shards []string
}

func (o *Options) validate() error {
	if o.Installation == nil {
		return errors.New("Installation is required")
	}
	// Every rendered object is namespaced, so the Installation has to be.
	if o.Installation.Namespace == "" {
		return errors.New("Installation must set metadata.namespace")
	}
	if o.Installation.Spec.DomainBase == "" {
		return errors.New("Installation spec.domainBase is required")
	}
	if len(o.Installation.Spec.Etcd.Endpoints) == 0 {
		return errors.New("Installation spec.etcd.endpoints is required")
	}
	if len(o.Shards) == 0 {
		o.Shards = defaultShards()
	}

	seen := map[string]struct{}{}
	roots := 0
	for _, name := range o.Shards {
		if errs := validation.IsDNS1123Label(name); len(errs) > 0 {
			return fmt.Errorf("shard name %q: %s", name, strings.Join(errs, ", "))
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("shard name %q is used twice", name)
		}
		seen[name] = struct{}{}
		if name == rootShardName {
			roots++
		}
	}
	if roots != 1 {
		return fmt.Errorf("shards %v: exactly one shard must be named %q", o.Shards, rootShardName)
	}
	return nil
}

// Manifests are the objects one Installation renders.
type Manifests struct {
	CacheServer *operatorv1alpha1.CacheServer
	RootShard   *operatorv1alpha1.RootShard
	Shards      []*operatorv1alpha1.Shard
	FrontProxy  *operatorv1alpha1.FrontProxy
	Routes      []*gwapiv1.TLSRoute
}

// Objects lists the manifests in apply order: the cache-server the shards
// coordinate through, the root shard, its shards, the front-proxy, then the routes.
func (m *Manifests) Objects() []runtime.Object {
	objects := make([]runtime.Object, 0, 3+len(m.Shards)+len(m.Routes))
	objects = append(objects, m.CacheServer, m.RootShard)
	for _, shard := range m.Shards {
		objects = append(objects, shard)
	}
	objects = append(objects, m.FrontProxy)
	for _, route := range m.Routes {
		objects = append(objects, route)
	}
	return objects
}

// Render turns an Installation into the kcp-operator objects of a single-cluster kcp.
func Render(opts Options) (*Manifests, error) {
	if err := opts.validate(); err != nil {
		return nil, fmt.Errorf("invalid render options: %w", err)
	}
	installation := opts.Installation

	manifests := &Manifests{
		CacheServer: cacheServer(installation),
		FrontProxy:  frontProxy(installation),
	}
	for _, name := range opts.Shards {
		if name == rootShardName {
			manifests.RootShard = rootShard(installation)
			continue
		}
		manifests.Shards = append(manifests.Shards, shard(installation, name))
	}
	manifests.Routes = routes(installation, opts.Shards)

	return manifests, nil
}

func cacheServer(installation *deployv1alpha1.Installation) *operatorv1alpha1.CacheServer {
	return &operatorv1alpha1.CacheServer{
		TypeMeta:   operatorTypeMeta("CacheServer"),
		ObjectMeta: metav1.ObjectMeta{Name: installation.Name, Namespace: installation.Namespace},
		Spec: operatorv1alpha1.CacheServerSpec{
			Certificates: certificates(),
			// One replica: the cache writes to the installation's etcd, so its own
			// availability follows that of the etcd, which runs a single member.
			Replicas: ptr.To(int32(1)),
			Etcd:     ptr.To(etcdConfig(installation, cacheEtcdPrefix)),
		},
	}
}

func rootShard(installation *deployv1alpha1.Installation) *operatorv1alpha1.RootShard {
	return &operatorv1alpha1.RootShard{
		TypeMeta:   operatorTypeMeta("RootShard"),
		ObjectMeta: metav1.ObjectMeta{Name: rootShardName, Namespace: installation.Namespace},
		Spec: operatorv1alpha1.RootShardSpec{
			CommonShardSpec: commonShardSpec(installation, rootShardName),
			External: operatorv1alpha1.ExternalConfig{
				Hostname: kcpHost(installation),
				Port:     externalPort,
				// The operator puts this URL into the kubeconfigs it mounts into the kcp
				// pods. sslip.io answers 127.0.0.1 there, so in-cluster traffic has to
				// reach the front-proxy Service by name instead.
				PrivateHostname: serviceHost(frontProxyServiceName(installation), installation),
				PrivatePort:     ptr.To(uint32(externalPort)),
			},
			Certificates: certificates(),
			Cache: operatorv1alpha1.RootShardCacheConfig{
				Reference: &corev1.LocalObjectReference{Name: installation.Name},
			},
		},
	}
}

func shard(installation *deployv1alpha1.Installation, name string) *operatorv1alpha1.Shard {
	return &operatorv1alpha1.Shard{
		TypeMeta:   operatorTypeMeta("Shard"),
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: installation.Namespace},
		Spec: operatorv1alpha1.ShardSpec{
			CommonShardSpec: commonShardSpec(installation, name),
			RootShard:       rootShardRef(),
		},
	}
}

func frontProxy(installation *deployv1alpha1.Installation) *operatorv1alpha1.FrontProxy {
	return &operatorv1alpha1.FrontProxy{
		TypeMeta:   operatorTypeMeta("FrontProxy"),
		ObjectMeta: metav1.ObjectMeta{Name: installation.Name, Namespace: installation.Namespace},
		Spec: operatorv1alpha1.FrontProxySpec{
			RootShard: rootShardRef(),
			// The Service port follows the external port, which is what the route sends to.
			External: operatorv1alpha1.ExternalConfig{
				Hostname: kcpHost(installation),
				Port:     externalPort,
			},
			Replicas: ptr.To(int32(1)),
			Auth:     authSpec(installation),
		},
	}
}

func commonShardSpec(installation *deployv1alpha1.Installation, name string) operatorv1alpha1.CommonShardSpec {
	return operatorv1alpha1.CommonShardSpec{
		// One replica per shard: the development cluster is a single kind node and
		// the operator would otherwise default to two.
		Replicas: ptr.To(int32(1)),
		Etcd:     etcdConfig(installation, shardEtcdPrefix+"/"+name),
		Auth:     authSpec(installation),
		// CertificateTemplates and not ShardBaseURL: the latter also becomes the
		// server of the kubeconfig the operator mounts into the shard's own Pod,
		// where sslip.io answers 127.0.0.1.
		CertificateTemplates: serverCertificate(shardHost(installation, name)),
	}
}

func etcdConfig(installation *deployv1alpha1.Installation, prefix string) operatorv1alpha1.EtcdConfig {
	config := operatorv1alpha1.EtcdConfig{
		Endpoints: slices.Clone(installation.Spec.Etcd.Endpoints),
		Prefix:    prefix,
	}
	if ref := installation.Spec.Etcd.TLSSecretRef; ref != nil {
		config.TLSConfig = &operatorv1alpha1.EtcdTLSConfig{SecretRef: *ref}
	}
	return config
}

func authSpec(installation *deployv1alpha1.Installation) *operatorv1alpha1.AuthSpec {
	auth := &operatorv1alpha1.AuthSpec{
		// kcp mints service account tokens per shard, and the front-proxy has to be
		// told which ones to trust. The FrontProxy CRD also rejects OIDC without it.
		ServiceAccount: &operatorv1alpha1.ServiceAccountAuthentication{Enabled: true},
	}

	oidc := installation.Spec.OIDC
	if oidc.IssuerURL == "" && oidc.ClientID == "" {
		return auth
	}
	auth.OIDC = &operatorv1alpha1.OIDCConfiguration{
		IssuerURL: oidc.IssuerURL,
		ClientID:  oidc.ClientID,
	}
	if ref := oidc.CAFileRef; ref != nil && ref.Name != "" {
		auth.OIDC.CAFileRef = &operatorv1alpha1.OIDCCAFileRef{
			Name: ref.Name,
			Key:  ref.Key,
		}
		if auth.OIDC.CAFileRef.Key == "" {
			auth.OIDC.CAFileRef.Key = "ca.crt"
		}
	}
	return auth
}

func certificates() operatorv1alpha1.Certificates {
	return operatorv1alpha1.Certificates{
		IssuerRef: &operatorv1alpha1.ObjectReference{
			Name:  issuerName,
			Kind:  "Issuer",
			Group: "cert-manager.io",
		},
	}
}

// serverCertificate adds a public name to a component's serving certificate.
// No IssuerRef in the template on purpose: with one, the operator takes DNSNames
// as given instead of merging them into the names it derives itself.
func serverCertificate(hosts ...string) operatorv1alpha1.CertificateTemplateMap {
	template := operatorv1alpha1.CertificateTemplate{
		Spec: &operatorv1alpha1.CertificateSpecTemplate{
			DNSNames: hosts,
		},
	}
	return operatorv1alpha1.CertificateTemplateMap{
		string(operatorv1alpha1.ServerCertificate): template,
	}
}

func rootShardRef() operatorv1alpha1.RootShardConfig {
	return operatorv1alpha1.RootShardConfig{
		Reference: &corev1.LocalObjectReference{Name: rootShardName},
	}
}

func operatorTypeMeta(kind string) metav1.TypeMeta {
	return metav1.TypeMeta{
		APIVersion: operatorv1alpha1.SchemeGroupVersion.String(),
		Kind:       kind,
	}
}

// kcpHost is the external host of the front-proxy.
func kcpHost(installation *deployv1alpha1.Installation) string {
	return hostLabel + "." + installation.Spec.DomainBase
}

// shardHost is the external host of one shard.
func shardHost(installation *deployv1alpha1.Installation, name string) string {
	return name + "." + kcpHost(installation)
}

func rootShardServiceName(name string) string { return name + rootShardServiceSuffix }

func shardServiceName(name string) string { return name + shardServiceSuffix }

func frontProxyServiceName(installation *deployv1alpha1.Installation) string {
	return installation.Name + frontProxyServiceSuffix
}

// serviceHost is the in-cluster address of one of those Services.
func serviceHost(service string, installation *deployv1alpha1.Installation) string {
	return fmt.Sprintf("%s.%s.svc.%s", service, installation.Namespace, clusterDomain)
}
