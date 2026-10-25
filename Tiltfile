# SPDX-License-Identifier: Apache-2.0
# SPDX-FileCopyrightText: The heimstatt Authors.

local("kubectl apply --server-side -f config/deploy/crds")

# apply and wait for cert-manager - if the webhook isn't ready some object deadlock
local("kubectl apply --server-side -k config/tilt/components/cert-manager")
local("kubectl -n cert-manager wait --for=condition=Available --timeout=180s deployment/cert-manager-webhook")

local("kubectl apply --server-side -k config/tilt/components/kcp-operator")

local("kubectl apply --server-side -k config/tilt/components/envoy-gateway")
local("kubectl -n envoy-gateway-system wait --for=condition=Available --timeout=180s deployment/envoy-gateway")

k8s_yaml(kustomize('config/tilt/components'))
k8s_yaml(kustomize("config/tilt/manifests"))

k8s_resource("dex", port_forwards = 5556)
