package cmd

import (
	"code.cloudfoundry.org/cfdev/config"
	"github.com/spf13/cobra"
)

func NewVersion(UI UI, Config config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use: "version",
		RunE: func(cmd *cobra.Command, args []string) error {
			UI.Say("Version: %s", Config.CliVersion.Original)
			return nil
		},
	}
	return cmd
}
