// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

// Package heimkcp renders the kcp-operator objects of one deploy.heimstatt.dev
// Installation: a cache-server, a root shard, its shards, a front-proxy and the
// gateway routes that expose them.
//
// It holds the topology a single-cluster installation has, because the
// Installation CRD does not express one yet.
//
// The package is heimkcp rather than kcp so it cannot be read as the upstream
// kcp API surface whose objects it builds.
package heimkcp
