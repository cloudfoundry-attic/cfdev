package cmd

import (
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"path/filepath"

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
	"code.cloudfoundry.org/cfdev/hypervisor"
	"code.cloudfoundry.org/cfdev/iso"
	"code.cloudfoundry.org/cfdev/network"
	"code.cloudfoundry.org/cfdev/provision"
	"code.cloudfoundry.org/cfdev/resource"
	"code.cloudfoundry.org/cfdev/resource/progress"
	cfdevdClient "code.cloudfoundry.org/cfdevd/client"
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
	lctl := daemon.New(config.CFDevHome)

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
	linuxkit := &hypervisor.LinuxKit{Config: config, DaemonRunner: lctl}
	vpnkit := &network.VpnKit{Config: config, DaemonRunner: lctl}

	dev := &cobra.Command{
		Use:           "dev",
		Short:         "Start and stop a single vm CF deployment running on your workstation",
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	root.AddCommand(dev)

	for _, cmd := range []cmdBuilder{
		&b1.Version{
			UI:      ui,
			Version: config.CliVersion,
		},
		&b2.Bosh{
			Exit:        exit,
			UI:          ui,
			Config:      config,
			Provisioner: provision.NewController(),
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
			HostNet:         &network.HostNet{},
			CFDevD:          &network.CFDevD{ExecutablePath: filepath.Join(config.CacheDir, "cfdevd")},
			VpnKit:          vpnkit,
			Hypervisor:      linuxkit,
			Provisioner:     provision.NewController(),
			IsoReader:       iso.New(),
		},
		&b6.Stop{
			Config:       config,
			Analytics:    analyticsClient,
			Hypervisor:   linuxkit,
			HostNet:      &network.HostNet{},
			VpnKit:       vpnkit,
			CfdevdClient: cfdevdClient.New("CFD3V", config.CFDevDSocketPath),
		},
		&b7.Telemetry{
			UI:              ui,
			AnalyticsToggle: analyticsToggle,
		},
		&b8.Logs{
			UI: ui,
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
