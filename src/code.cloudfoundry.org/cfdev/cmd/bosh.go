package cmd

import (
	"os"
	"fmt"
	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
	"code.cloudfoundry.org/cfdev/shell"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
)

type Bosh struct{
	Exit chan struct{}
}

func(b *Bosh) Run(args []string) {
	go func() {
		<-b.Exit
		os.Exit(128)
	}()

	cfui := terminal.NewUI(
		os.Stdin,
		os.Stdout,
		terminal.NewTeePrinter(os.Stdout),
		trace.NewLogger(os.Stdout, false, "", ""),
	)


	_, stateDir, _ := setupHomeDir()

	if len(args) == 0 || args[0] != "env" {
		fmt.Fprintf(os.Stderr, `Usage: eval $(cf dev bosh env)`)
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}

	gClient := client.New(connection.New("tcp", "localhost:8888"))
	config, err := gdn.FetchBOSHConfig(gClient)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to fetch bosh configuration: %v\n", err)
		os.Exit(1)
	}

	env := shell.Environment{StateDir: stateDir}
	shellScript, err := env.Prepare(config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to prepare bosh configuration: %v\n", err)
		os.Exit(1)
	}

	cfui.Say(shellScript)
}