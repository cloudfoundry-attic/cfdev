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
	"gopkg.in/segmentio/analytics-go.v3"
)

type Command interface {
	Run(args []string) error
}

type Plugin struct {
	Exit            chan struct{}
	UI              terminal.UI
	Config          config.Config
	AnalyticsClient analytics.Client
	Root            *cobra.Command
	Version         plugin.VersionType
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

	analyticsClient := analytics.New(conf.AnalyticsKey)
	cfdev := &Plugin{
		Exit:            exitChan,
		UI:              ui,
		Config:          conf,
		AnalyticsClient: analyticsClient,
		Root:            cmd.NewRoot(exitChan, ui, conf, analyticsClient),
		Version:         plugin.VersionType{Major: 0, Minor: 0, Build: 2},
	}
	defer cfdev.AnalyticsClient.Close()

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
		cfanalytics.TrackEvent(cfanalytics.UNINSTALL, nil, p.AnalyticsClient)
		stop := cmd.NewStop(&p.Config, p.AnalyticsClient)
		if err := stop.RunE(nil, []string{}); err != nil {
			p.UI.Say("Error stopping cfdev: %s", err)
			cfanalytics.TrackEvent(cfanalytics.ERROR, map[string]interface{}{"error": err}, p.AnalyticsClient)
		}
		return
	}

	cfanalytics.PromptOptIn(p.Config, p.UI)

	p.Root.SetArgs(args)
	if err := p.Root.Execute(); err != nil {
		cfanalytics.TrackEvent(cfanalytics.ERROR, map[string]interface{}{"error": err}, p.AnalyticsClient)
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
