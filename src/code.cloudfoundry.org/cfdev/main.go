package main

import (
	"os"
	"os/signal"
	"syscall"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/cmd"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin"
	"github.com/spf13/cobra"
)

type Command interface {
	Run(args []string) error
}

type Plugin struct {
	Exit    chan struct{}
	UI      terminal.UI
	Config  config.Config
	Root    *cobra.Command
	Version plugin.VersionType
}

func main() {
	exitChan := make(chan struct{})
	sigChan := make(chan os.Signal, 1)
	signal.Notify(make(chan os.Signal), syscall.SIGHUP)
	signal.Notify(sigChan, syscall.SIGINT)
	signal.Notify(sigChan, syscall.SIGTERM)

	go func() {
		<-sigChan
		close(exitChan)
	}()

	ui := terminal.NewUI(
		os.Stdin,
		os.Stdout,
		terminal.NewTeePrinter(os.Stdout),
		trace.NewLogger(os.Stdout, false, "", ""),
	)

	conf, err := config.NewConfig()
	if err != nil {
		ui.Failed(err.Error())
		os.Exit(1)
	}
	defer conf.Close()

	v := conf.CliVersion
	cfdev := &Plugin{
		Exit:    exitChan,
		UI:      ui,
		Config:  conf,
		Root:    cmd.NewRoot(exitChan, ui, conf),
		Version: plugin.VersionType{Major: v.Major, Minor: v.Minor, Build: v.Build},
	}

	plugin.Start(cfdev)
}

func (p *Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name:    "cfdev",
		Version: p.Version,
		Commands: []plugin.Command{
			{
				Name:     "dev",
				HelpText: "Start and stop a single vm CF deployment running on your workstation",
				UsageDetails: plugin.Usage{
					Usage: p.Root.UsageString(),
				},
			},
		},
	}
}

func (p *Plugin) Run(connection plugin.CliConnection, args []string) {
	if args[0] == "CLI-MESSAGE-UNINSTALL" {
		p.Config.Analytics.Event(cfanalytics.UNINSTALL, nil)
		stop := cmd.NewStop(p.Config)
		if err := stop.RunE(nil, []string{}); err != nil {
			p.UI.Say("Error stopping cfdev: %s", err)
			p.Config.Analytics.Event(cfanalytics.ERROR, map[string]interface{}{"error": err})
		}
		return
	}

	if err := p.Config.Analytics.PromptOptIn(p.Exit, p.UI); err != nil {
		p.UI.Failed(err.Error())
		p.Config.Close()
		os.Exit(1)
	}

	p.Root.SetArgs(args)
	if err := p.Root.Execute(); err != nil {
		p.UI.Failed(err.Error())
		p.Config.Analytics.Event(cfanalytics.ERROR, map[string]interface{}{"error": err.Error()})
		p.Config.Close()
		os.Exit(1)
	}

	// TODO why is the below here?????
	// select {
	// case <-p.Exit:
	// 	os.Exit(128)
	// default:
	// 	if err != nil {
	// 		fmt.Printf("Error encountered running '%s' : %s", args[0], err)
	// 		os.Exit(2)
	// 	}
	// }
}
