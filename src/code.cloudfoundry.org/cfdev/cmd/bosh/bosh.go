package bosh

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

type UI interface {
	Say(message string, args ...interface{})
}

type Bosh struct {
	Exit   chan struct{}
	UI     UI
	Config config.Config
}

func (b *Bosh) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "bosh",
		Run: func(cmd *cobra.Command, args []string) {
			b.UI.Say(`Usage: eval $(cf dev bosh env)`)
		},
	}
	envCmd := &cobra.Command{
		Use:  "env",
		RunE: b.RunE,
	}
	cmd.AddCommand(envCmd)
	return cmd
}

func (b *Bosh) RunE(cmd *cobra.Command, args []string) error {
	go func() {
		<-b.Exit
		os.Exit(128)
	}()

	gClient := client.New(connection.New("tcp", "localhost:8888"))
	config, err := gdn.FetchBOSHConfig(gClient)
	if err != nil {
		return errors.SafeWrap(err, "failed to fetch bosh configuration")
	}

	env := shell.Environment{StateDir: b.Config.StateDir}
	shellScript, err := env.Prepare(config)
	if err != nil {
		return errors.SafeWrap(err, "failed to prepare bosh configuration")
	}

	b.UI.Say(shellScript)
	return nil
}
