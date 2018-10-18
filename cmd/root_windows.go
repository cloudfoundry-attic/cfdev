package cmd

import (
	"code.cloudfoundry.org/cfdev/profiler"
	"code.cloudfoundry.org/cfdev/runner"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"path/filepath"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	b2 "code.cloudfoundry.org/cfdev/cmd/bosh"
	b3 "code.cloudfoundry.org/cfdev/cmd/catalog"
	b4 "code.cloudfoundry.org/cfdev/cmd/download"
	b8 "code.cloudfoundry.org/cfdev/cmd/logs"
	b5 "code.cloudfoundry.org/cfdev/cmd/start"
	b6 "code.cloudfoundry.org/cfdev/cmd/stop"
	b7 "code.cloudfoundry.org/cfdev/cmd/telemetry"
	b1 "code.cloudfoundry.org/cfdev/cmd/version"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/host"
	"code.cloudfoundry.org/cfdev/hypervisor"
	"code.cloudfoundry.org/cfdev/iso"
	"code.cloudfoundry.org/cfdev/network"
	"code.cloudfoundry.org/cfdev/provision"
	"code.cloudfoundry.org/cfdev/resource"
	"code.cloudfoundry.org/cfdev/resource/progress"
	"github.com/spf13/cobra"
)

type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
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

func NewRoot(exit chan struct{}, ui UI, config config.Config, analyticsClient AnalyticsClient, analyticsToggle Toggle) *cobra.Command {
	root := &cobra.Command{Use: "cf", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().Bool("help", false, "")
	root.PersistentFlags().Lookup("help").Hidden = true
	lctl := daemon.NewWinSW(config.CFDevHome)
	vpnkit := &network.VpnKit{Config: config, DaemonRunner: lctl, Powershell: runner.Powershell{}, Label: network.VpnKitLabel}
	isoReader := iso.New()
	hostnet := &network.HostNet{
		VMSwitchName: "cfdev",
	}

	usageTemplate := strings.Replace(root.UsageTemplate(), "\n"+`Use "{{.CommandPath}} [command] --help" for more information about a command.`, "", -1)
	root.SetUsageTemplate(usageTemplate)

	skipVerify := strings.ToLower(os.Getenv("CFDEV_SKIP_ASSET_CHECK"))
	writer := ui.Writer()
	cache := &resource.Cache{
		Dir:                   config.CacheDir,
		HttpDo:                http.DefaultClient.Do,
		SkipAssetVerification: skipVerify == "true",
		Progress:              progress.New(writer),
		RetryWait:             time.Second,
		Writer:                writer,
	}

	analyticsD := &cfanalytics.AnalyticsD{
		Config:       config,
		DaemonRunner: lctl,
	}

	dev := &cobra.Command{
		Use:           "dev",
		Short:         "Start and stop a single vm CF deployment running on your workstation",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(dev)

	for _, cmd := range []cmdBuilder{
		&b1.Version{
			UI:        ui,
			Version:   config.CliVersion,
			Config:    config,
			IsoReader: isoReader,
		},
		&b2.Bosh{
			Exit:        exit,
			UI:          ui,
			StateDir:    config.StateDir,
			Provisioner: provision.NewController(),
			Analytics:   analyticsClient,
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
			LocalExit:       make(chan string, 3),
			UI:              ui,
			Config:          config,
			Cache:           cache,
			Analytics:       analyticsClient,
			AnalyticsToggle: analyticsToggle,
			HostNet:         hostnet,
			Host: &host.Host{
				Powershell: &runner.Powershell{},
			},
			AnalyticsD:  analyticsD,
			CFDevD:      &network.CFDevD{ExecutablePath: filepath.Join(config.CacheDir, "cfdevd")},
			Hypervisor:  &hypervisor.HyperV{Config: config},
			VpnKit:      vpnkit,
			Provisioner: provision.NewController(),
			IsoReader:   isoReader,
			Stop: &b6.Stop{
				Config:     config,
				Analytics:  analyticsClient,
				Hypervisor: &hypervisor.HyperV{Config: config},
				VpnKit:     vpnkit,
				HostNet:    hostnet,
				Host: &host.Host{
					Powershell: &runner.Powershell{},
				},
				AnalyticsD: analyticsD,
			},
			Profiler: &profiler.SystemProfiler{},
		},
		&b6.Stop{
			Config:     config,
			Analytics:  analyticsClient,
			Hypervisor: &hypervisor.HyperV{Config: config},
			VpnKit:     vpnkit,
			HostNet:    hostnet,
			Host: &host.Host{
				Powershell: &runner.Powershell{},
			},
			AnalyticsD: analyticsD,
		},
		&b7.Telemetry{
			UI:              ui,
			AnalyticsToggle: analyticsToggle,
			AnalyticsD:      analyticsD,
		},
		&b8.Logs{
			Provisioner: provision.NewController(),
			UI:          ui,
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
