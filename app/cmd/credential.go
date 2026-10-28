package cmd

import (
	"github.com/rody-vx/jenkins-cli/app/cmd/common"
	"github.com/rody-vx/jenkins-cli/app/i18n"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(credentialCmd)
}

var credentialCmd = &cobra.Command{
	Use:     "credential",
	Aliases: []string{"secret", "cred"},
	Short:   i18n.T("Manage the credentials of your Jenkins"),
	Long:    i18n.T(`Manage the credentials of your Jenkins`),
	Annotations: map[string]string{
		common.Since: common.VersionSince0024,
	},
}
