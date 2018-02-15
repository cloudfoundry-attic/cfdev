package cmd

import (
	"os"
	"fmt"
	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
	"code.cloudfoundry.org/cfdev/shell"
)

type Bosh struct{
	Exit chan struct{}
	UI UI
}

func(b *Bosh) Run(args []string) error {
	go func() {
		<-b.Exit
		os.Exit(128)
	}()
	_, stateDir, _, err := setupHomeDir()
	if err != nil {
		return err
	}

	if len(args) == 0 || args[0] != "env" {
		b.UI.Say(`Usage: eval $(cf dev bosh env)`)
		return nil
	}

	gClient := client.New(connection.New("tcp", "localhost:8888"))
	config, err := gdn.FetchBOSHConfig(gClient)
	if err != nil {
		return fmt.Errorf( "failed to fetch bosh configuration: %v\n", err)
	}

	env := shell.Environment{StateDir: stateDir}
	shellScript, err := env.Prepare(config)
	if err != nil {
		return fmt.Errorf( "failed to prepare bosh configuration: %v\n", err)
	}

	b.UI.Say(shellScript)
	return nil
}