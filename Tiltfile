# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: The heimstatt Authors.

# Upstream operators are one-shot installs whose pods we do not build. They are resources rather than
# local() calls at Tiltfile parse time so the dependency graph sees them: nothing that consumes their
# CRDs is deployed before they are up, and a failure shows in its own log pane instead of aborting startup.
local_resource(
    name = 'cert-manager',
    cmd = 'kubectl apply --server-side -k config/tilt/components/cert-manager' +
          ' && kubectl -n cert-manager wait --for=condition=Available --timeout=180s deployment/cert-manager-webhook',
    deps = ['config/tilt/components/cert-manager'],
    allow_parallel = True,
)

local_resource(
    name = 'envoy-gateway',
    cmd = 'kubectl apply --server-side -k config/tilt/components/envoy-gateway' +
          ' && kubectl -n envoy-gateway-system wait --for=condition=Available --timeout=180s deployment/envoy-gateway',
    deps = ['config/tilt/components/envoy-gateway'],
    allow_parallel = True,
)

local_resource(
    name = 'kcp-operator',
    cmd = 'kubectl apply --server-side -k config/tilt/components/kcp-operator' +
          ' && kubectl -n kcp-operator-system wait --for=condition=Available --timeout=180s deployment/kcp-operator-controller-manager',
    deps = ['config/tilt/components/kcp-operator'],
    allow_parallel = True,
)

k8s_yaml('config/tilt/namespace.yaml')
k8s_yaml('config/deploy/crds/deploy.heimstatt.dev_installations.yaml')
k8s_yaml(kustomize('config/tilt/components/etcd'))
k8s_yaml(kustomize('config/tilt/components'))
k8s_yaml(kustomize('config/tilt/manifests'))

# Objects that create no pods are only "ready" once they are applied.
k8s_resource(
    new_name = 'heimstatt-namespace',
    objects = ['heimstatt:namespace'],
    pod_readiness = 'ignore',
)

k8s_resource(
    new_name = 'heimstatt-crds',
    objects = ['installations.deploy.heimstatt.dev:customresourcedefinition'],
    pod_readiness = 'ignore',
)

k8s_resource(
    new_name = 'heimstatt-etcd',
    workload = 'etcd',
    resource_deps = ['heimstatt-namespace'],
)

# The serving certificate the gateway listens with.
k8s_resource(
    new_name = 'heimstatt-tls',
    objects = ['selfsigned:issuer', 'heimstatt-serving:certificate'],
    pod_readiness = 'ignore',
    resource_deps = ['heimstatt-namespace', 'cert-manager'],
)

k8s_resource(
    new_name = 'heimstatt-gateway',
    objects = ['heimstatt:gatewayclass', 'heimstatt:gateway', 'heimstatt:envoyproxy'],
    pod_readiness = 'ignore',
    resource_deps = ['heimstatt-namespace', 'envoy-gateway', 'heimstatt-tls'],
)

k8s_resource(
    workload = 'dex',
    objects = ['dex:namespace', 'dex:configmap', 'dex-heimstatt-kcp:secret', 'dex:httproute'],
    port_forwards = 5556,
    resource_deps = ['heimstatt-gateway', 'cert-manager'],
)

k8s_resource(
    new_name = 'installation',
    objects = ['heimstatt:installation'],
    pod_readiness = 'ignore',
    resource_deps = ['heimstatt-crds', 'heimstatt-namespace', 'heimstatt-etcd', 'kcp-operator'],
)
