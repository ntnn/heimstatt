// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package app

import (
	"bytes"
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const installationYAML = `apiVersion: deploy.heimstatt.dev/v1alpha1
kind: Installation
metadata:
  name: heimstatt
  namespace: heimstatt
spec:
  domainBase: heimstatt.127-0-0-1.sslip.io
  etcd:
    endpoints:
    - http://etcd.heimstatt.svc.cluster.local:2379
`

// The CRDs reject a null where an object is expected, so nothing may be rendered as null.
func TestTemplate_RendersNoNulls(t *testing.T) {
	t.Parallel()

	var out bytes.Buffer
	err := Template(context.Background(), TemplateOptions{In: strings.NewReader(installationYAML), Out: &out})
	require.NoError(t, err)

	rendered := out.String()
	assert.NotContains(t, rendered, "null", "the kcp-operator SDK marshals unset pointers as null")
	assert.Contains(t, rendered, "issuerRef:\n      group: cert-manager.io\n      kind: Issuer\n      name: selfsigned",
		"the certificates issuerRef is set and survives")
}
