// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package heimkcp

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	gwapiv1 "sigs.k8s.io/gateway-api/apis/v1"

	deployv1alpha1 "heimstatt.dev/apis/deploy/v1alpha1"
)

// gatewayName is the Gateway the kcp routes attach to. It lives in the
// Installation's namespace and is created by config/tilt/components/gateway.yaml.
const gatewayName = "heimstatt"

// routes exposes the front-proxy and every shard on its own hostname.
// kcp serves TLS with a certificate of its own CA, so the gateway can only pass the
// connection through and matches the hostnames on SNI.
func routes(installation *deployv1alpha1.Installation, shards []string) []*gwapiv1.TLSRoute {
	rendered := make([]*gwapiv1.TLSRoute, 0, 1+len(shards))
	rendered = append(rendered, route(installation, frontProxyServiceName(installation), kcpHost(installation), externalPort))
	for _, name := range shards {
		service := shardServiceName(name)
		if name == rootShardName {
			service = rootShardServiceName(name)
		}
		rendered = append(rendered, route(installation, service, shardHost(installation, name), shardPort))
	}
	return rendered
}

func route(installation *deployv1alpha1.Installation, service, host string, port int) *gwapiv1.TLSRoute {
	return &gwapiv1.TLSRoute{
		TypeMeta: metav1.TypeMeta{
			APIVersion: gwapiv1.GroupVersion.String(),
			Kind:       "TLSRoute",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      service,
			Namespace: installation.Namespace,
		},
		Spec: gwapiv1.TLSRouteSpec{
			CommonRouteSpec: gwapiv1.CommonRouteSpec{
				ParentRefs: []gwapiv1.ParentReference{
					{
						Group: ptr.To(gwapiv1.Group(gwapiv1.GroupName)),
						Kind:  ptr.To(gwapiv1.Kind("Gateway")),
						Name:  gwapiv1.ObjectName(gatewayName),
					},
				},
			},
			Hostnames: []gwapiv1.Hostname{
				gwapiv1.Hostname(host),
			},
			Rules: []gwapiv1.TLSRouteRule{
				{
					BackendRefs: []gwapiv1.BackendRef{
						{
							BackendObjectReference: gwapiv1.BackendObjectReference{
								Group: ptr.To(gwapiv1.Group("")),
								Kind:  ptr.To(gwapiv1.Kind("Service")),
								Name:  gwapiv1.ObjectName(service),
								Port:  ptr.To(gwapiv1.PortNumber(port)),
							},
						},
					},
				},
			},
		},
	}
}
