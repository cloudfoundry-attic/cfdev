package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/cmd"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cli/cf/terminal"
	"code.cloudfoundry.org/cli/cf/trace"
	"code.cloudfoundry.org/cli/plugin"
	"gopkg.in/segmentio/analytics-go.v3"
)

var analyticsClient analytics.Client

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

	cfdev := &Plugin{
		Exit:   exitChan,
		UI:     ui,
		Config: conf,
	}

	plugin.Start(cfdev)
}

type Command interface {
	Run(args []string) error
}

type Plugin struct {
	Exit   chan struct{}
	UI     terminal.UI
	Config config.Config
}

func (p *Plugin) Run(connection plugin.CliConnection, args []string) {
	if args[0] == "CLI-MESSAGE-UNINSTALL" {
		stop := &cmd.Stop{
			Config:          p.Config,
			AnalyticsClient: analyticsClient,
		}
		if err := stop.Run([]string{}); err != nil {
			p.UI.Say("Error stopping cfdev: %s", err)
		}
		return
	}
	p.execute(args[1:])
}

func (p *Plugin) GetMetadata() plugin.PluginMetadata {
	return plugin.PluginMetadata{
		Name: "cfdev",
		Version: plugin.VersionType{
			Major: 0,
			Minor: 0,
			Build: 2,
		},
		Commands: []plugin.Command{
			{
				Name:     "dev",
				HelpText: "Start and stop a single vm CF deployment running on your workstation",
				UsageDetails: plugin.Usage{
					Usage: "cf dev [start|stop|bosh]",
				},
			},
		},
	}
}

func (p *Plugin) usage() {
	p.UI.Say("cf dev [start|stop|bosh]")
	os.Exit(1)
}

func (p *Plugin) execute(args []string) {
	if len(args) == 0 {
		p.usage()
	}

	analyticsClient = analytics.New(p.Config.AnalyticsKey)
	cfanalytics.PromptOptIn(p.Config, p.UI)

	var command Command
	switch args[0] {
	case "start":
		command = &cmd.Start{
			Exit:            p.Exit,
			UI:              p.UI,
			Config:          p.Config,
			AnalyticsClient: analyticsClient,
		}
	case "stop":
		command = &cmd.Stop{
			Config:          p.Config,
			AnalyticsClient: analyticsClient,
		}
	case "download":
		command = &cmd.Download{
			Exit:   p.Exit,
			UI:     p.UI,
			Config: p.Config,
		}
	case "bosh":
		command = &cmd.Bosh{
			Exit:   p.Exit,
			UI:     p.UI,
			Config: p.Config,
		}
	case "catalog":
		command = &cmd.Catalog{
			UI:     p.UI,
			Config: p.Config,
		}
	case "telemetry":
		command = &cmd.Telemetry{
			UI:     p.UI,
			Config: p.Config,
		}
	default:
		p.usage()
	}

	err := command.Run(args[1:])
	if err != nil {
		cfanalytics.TrackEvent(cfanalytics.ERROR, map[string]interface{}{"error": err}, analyticsClient)
	}
	analyticsClient.Close()

	select {
	case <-p.Exit:
		os.Exit(128)
	default:
		if err != nil {
			fmt.Printf("Error encountered running '%s' : %s", args[0], err)
			os.Exit(2)
		}
	}
}
