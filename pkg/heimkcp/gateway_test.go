// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package heimkcp

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRoutes(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		shards []string
		want   []string
	}{
		"default topology": {
			shards: defaultShards(),
			want: []string{
				"kcp.heimstatt.127-0-0-1.sslip.io heimstatt-front-proxy:30444",
				// The root shard's Service is named after the RootShard object, not with
				// the -shard-kcp suffix a Shard gets.
				"root.kcp.heimstatt.127-0-0-1.sslip.io root-kcp:6443",
				"secondary.kcp.heimstatt.127-0-0-1.sslip.io secondary-shard-kcp:6443",
			},
		},
		"three shards": {
			shards: []string{"root", "east", "west"},
			want: []string{
				"kcp.heimstatt.127-0-0-1.sslip.io heimstatt-front-proxy:30444",
				"root.kcp.heimstatt.127-0-0-1.sslip.io root-kcp:6443",
				"east.kcp.heimstatt.127-0-0-1.sslip.io east-shard-kcp:6443",
				"west.kcp.heimstatt.127-0-0-1.sslip.io west-shard-kcp:6443",
			},
		},
	}

	for title, cas := range cases {
		t.Run(title, func(t *testing.T) {
			t.Parallel()

			rendered := routes(newInstallation(), cas.shards)

			got := make([]string, 0, len(rendered))
			for _, route := range rendered {
				require.Len(t, route.Spec.Hostnames, 1)
				require.Len(t, route.Spec.Rules, 1)
				require.Len(t, route.Spec.Rules[0].BackendRefs, 1)

				backend := route.Spec.Rules[0].BackendRefs[0].BackendObjectReference
				got = append(got, fmt.Sprintf("%s %s:%d", route.Spec.Hostnames[0], backend.Name, *backend.Port))

				assert.Equal(t, route.Name, string(backend.Name), "the route is named after its backend")
				assert.Equal(t, "heimstatt", string(route.Spec.ParentRefs[0].Name))
				assert.Empty(t, route.Spec.ParentRefs[0].Namespace, "the Gateway sits in the Installation's namespace")
			}
			assert.Equal(t, cas.want, got)
		})
	}
}
