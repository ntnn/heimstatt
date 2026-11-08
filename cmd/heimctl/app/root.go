// SPDX-License-Identifier: Apache-2.0
// SPDX-FileCopyrightText: The heimstatt Authors.

package app

import (
	"github.com/spf13/cobra"
)

// NewCommand returns the heimctl root command with every subcommand attached.
func NewCommand() *cobra.Command {
	root := &cobra.Command{
		Use:   "heimctl",
		Short: "heimstatt cli",
	}
	root.AddCommand(newTemplateCommand())
	return root
}
