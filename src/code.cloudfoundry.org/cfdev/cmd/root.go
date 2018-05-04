package cmd

import (
	"io"
	"strings"

	b2 "code.cloudfoundry.org/cfdev/cmd/bosh"
	b3 "code.cloudfoundry.org/cfdev/cmd/catalog"
	b4 "code.cloudfoundry.org/cfdev/cmd/download"
	b5 "code.cloudfoundry.org/cfdev/cmd/start"
	b6 "code.cloudfoundry.org/cfdev/cmd/stop"
	b7 "code.cloudfoundry.org/cfdev/cmd/telemetry"
	b1 "code.cloudfoundry.org/cfdev/cmd/version"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/process"
	cfdevdClient "code.cloudfoundry.org/cfdevd/client"
	launchdModels "code.cloudfoundry.org/cfdevd/launchd/models"
	"github.com/spf13/cobra"
)

type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
}
type Launchd interface {
	AddDaemon(launchdModels.DaemonSpec) error
	RemoveDaemon(label string) error
	Start(label string) error
	Stop(label string) error
	IsRunning(label string) (bool, error)
}
type cmdBuilder interface {
	Cmd() *cobra.Command
}
type AnalyticsClient interface {
	Event(event string, data ...map[string]interface{}) error
	PromptOptIn() error
}
type Toggle interface {
	Get() bool
	Set(value bool) error
	SetProp(k, v string) error
}

func NewRoot(exit chan struct{}, ui UI, config config.Config, launchd Launchd, analyticsClient AnalyticsClient, analyticsToggle Toggle) *cobra.Command {
	root := &cobra.Command{Use: "cf", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().Bool("help", false, "")
	root.PersistentFlags().Lookup("help").Hidden = true

	usageTemplate := strings.Replace(root.UsageTemplate(), "\n"+`Use "{{.CommandPath}} [command] --help" for more information about a command.`, "", -1)
	root.SetUsageTemplate(usageTemplate)

	dev := &cobra.Command{
		Use:           "dev",
		Short:         "Start and stop a single vm CF deployment running on your workstation",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(dev)

	for _, cmd := range []cmdBuilder{
		&b1.Version{
			UI:     ui,
			Config: config,
		},
		&b2.Bosh{
			Exit:   exit,
			UI:     ui,
			Config: config,
		},
		&b3.Catalog{
			UI:     ui,
			Config: config,
		},
		&b4.Download{
			Exit:   exit,
			UI:     ui,
			Config: config,
		},
		&b5.Start{
			Exit:            exit,
			LocalExit:       make(chan struct{}, 3),
			UI:              ui,
			Config:          config,
			Launchd:         launchd,
			ProcManager:     &process.Manager{},
			Analytics:       analyticsClient,
			AnalyticsToggle: analyticsToggle,
		},
		&b6.Stop{
			Config:       config,
			Analytics:    analyticsClient,
			Launchd:      launchd,
			ProcManager:  &process.Manager{},
			CfdevdClient: cfdevdClient.New("CFD3V", config.CFDevDSocketPath),
		},
		&b7.Telemetry{
			UI:              ui,
			AnalyticsToggle: analyticsToggle,
		},
	} {
		dev.AddCommand(cmd.Cmd())
	}

	dev.AddCommand(&cobra.Command{
		Use:   "help [command]",
		Short: "Help about any command",
		Run: func(c *cobra.Command, args []string) {
			cmd, _, _ := dev.Find(args)
			cmd.Help()
		},
	})

	return root
}
