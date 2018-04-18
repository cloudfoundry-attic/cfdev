package cmd

import (
	"fmt"
	"os"

	"code.cloudfoundry.org/cfdev/config"
	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/cfdev/shell"
	"code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
)

type Bosh struct {
	Exit   chan struct{}
	UI     UI
	Config config.Config
}

func (b *Bosh) Run(args []string) error {
	go func() {
		<-b.Exit
		os.Exit(128)
	}()

	if len(args) == 0 || args[0] != "env" {
		b.UI.Say(`Usage: eval $(cf dev bosh env)`)
		return nil
	}

	gClient := client.New(connection.New("tcp", "localhost:8888"))
	config, err := gdn.FetchBOSHConfig(gClient)
	if err != nil {
		return fmt.Errorf("failed to fetch bosh configuration: %v\n", err)
	}

	env := shell.Environment{StateDir: b.Config.StateDir}
	shellScript, err := env.Prepare(config)
	if err != nil {
		return fmt.Errorf("failed to prepare bosh configuration: %v\n", err)
	}

	b.UI.Say(shellScript)
	return nil
}
