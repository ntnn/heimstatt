// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package app

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"heimstatt.dev/apis/deploy/v1alpha1"
	"heimstatt.dev/pkg/heimkcp"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/yaml"
)

// newTemplateCommand renders one Installation to stdout.
func newTemplateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "template <installation.yaml>",
		Short: "Render the kcp objects of one Installation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// The arguments are fine from here on, so failures need no usage.
			cmd.SilenceUsage = true

			file, err := os.Open(args[0])
			if err != nil {
				return err
			}
			defer file.Close()

			if err := Template(cmd.Context(), TemplateOptions{In: file, Out: cmd.OutOrStdout()}); err != nil {
				return fmt.Errorf("%s: %w", args[0], err)
			}
			return nil
		},
	}
}

// TemplateOptions configures one template run.
type TemplateOptions struct {
	// In is the Installation YAML to read.
	In io.Reader
	// Out receives one YAML document per rendered object.
	Out io.Writer
}

// Template renders the kcp objects of one Installation from In to Out.
func Template(ctx context.Context, opts TemplateOptions) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	data, err := io.ReadAll(opts.In)
	if err != nil {
		return fmt.Errorf("read: %w", err)
	}

	installation, err := decodeInstallation(data)
	if err != nil {
		return fmt.Errorf("decode: %w", err)
	}

	manifests, err := heimkcp.Render(heimkcp.Options{Installation: installation})
	if err != nil {
		return fmt.Errorf("render: %w", err)
	}

	for i, object := range manifests.Objects() {
		if i > 0 {
			if _, err := io.WriteString(opts.Out, "---\n"); err != nil {
				return fmt.Errorf("write: %w", err)
			}
		}
		document, err := marshal(object)
		if err != nil {
			return fmt.Errorf("marshal %T: %w", object, err)
		}
		if _, err := opts.Out.Write(document); err != nil {
			return fmt.Errorf("write: %w", err)
		}
	}
	return nil
}

// render to yaml omitting null fields.
// kcp-operator SDK doesn't have omitempty on
// CertificateSpecTemplate.IssuerRef, which marshalls to `null`, which
// errors on when applying
func marshal(object runtime.Object) ([]byte, error) {
	fields, err := runtime.DefaultUnstructuredConverter.ToUnstructured(object)
	if err != nil {
		return nil, err
	}
	dropNull(fields)
	return yaml.Marshal(fields)
}

// dropNulls deletes keys whose value is nil, recursively and in place.
func dropNull(fields map[string]any) {
	for key, value := range fields {
		switch typed := value.(type) {
		case nil:
			delete(fields, key)
		case map[string]any:
			dropNull(typed)
		case []any:
			for _, item := range typed {
				if element, ok := item.(map[string]any); ok {
					dropNull(element)
				}
			}
		}
	}
}

// decodeInstallation reads a single Installation, rejecting anything that is not one.
func decodeInstallation(data []byte) (*v1alpha1.Installation, error) {
	// validate manifest is an Installation before strict unmarshalling
	var meta metav1.TypeMeta
	if err := yaml.Unmarshal(data, &meta); err != nil {
		return nil, err
	}
	if meta.APIVersion != v1alpha1.GroupVersion.String() || meta.Kind != v1alpha1.InstallationKind {
		return nil, fmt.Errorf("document is %q/%q, want %s/%s",
			meta.APIVersion, meta.Kind, v1alpha1.GroupVersion, v1alpha1.InstallationKind)
	}

	installation := &v1alpha1.Installation{}
	if err := yaml.UnmarshalStrict(data, installation); err != nil {
		return nil, err
	}
	return installation, nil
}
