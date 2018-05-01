package cmd

import (
	"os"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/errors"
	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/cfdev/shell"
	"code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
	"github.com/spf13/cobra"
)

func NewBosh(Exit chan struct{}, UI UI, Config config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use: "bosh",
		Run: func(cmd *cobra.Command, args []string) {
			UI.Say(`Usage: eval $(cf dev bosh env)`)
		},
	}
	envCmd := &cobra.Command{
		Use: "env",
		RunE: func(cmd *cobra.Command, args []string) error {
			go func() {
				<-Exit
				os.Exit(128)
			}()

			gClient := client.New(connection.New("tcp", "localhost:8888"))
			config, err := gdn.FetchBOSHConfig(gClient)
			if err != nil {
				return errors.SafeWrap(err, "failed to fetch bosh configuration")
			}

			env := shell.Environment{StateDir: Config.StateDir}
			shellScript, err := env.Prepare(config)
			if err != nil {
				return errors.SafeWrap(err, "failed to prepare bosh configuration")
			}

			UI.Say(shellScript)
			return nil
		},
	}
	cmd.AddCommand(envCmd)
	return cmd
}
