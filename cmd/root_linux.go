package cmd

import (
	"code.cloudfoundry.org/cfdev/env"
	"code.cloudfoundry.org/cfdev/profiler"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"path/filepath"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	cfdevdClient "code.cloudfoundry.org/cfdev/cfdevd/client"
	b2 "code.cloudfoundry.org/cfdev/cmd/bosh"
	b3 "code.cloudfoundry.org/cfdev/cmd/catalog"
	b9 "code.cloudfoundry.org/cfdev/cmd/deploy-service"
	b4 "code.cloudfoundry.org/cfdev/cmd/download"
	b8 "code.cloudfoundry.org/cfdev/cmd/provision"
	b5 "code.cloudfoundry.org/cfdev/cmd/start"
	b6 "code.cloudfoundry.org/cfdev/cmd/stop"
	b7 "code.cloudfoundry.org/cfdev/cmd/telemetry"
	b1 "code.cloudfoundry.org/cfdev/cmd/version"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/daemon"
	"code.cloudfoundry.org/cfdev/host"
	"code.cloudfoundry.org/cfdev/hypervisor"
	"code.cloudfoundry.org/cfdev/metadata"
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
	PromptOptInIfNeeded(customMessage string) error
}

type Toggle interface {
	Defined() bool
	Enabled() bool
	CustomAnalyticsDefined() bool
	IsCustom() bool
	SetCFAnalyticsEnabled(value bool) error
	SetCustomAnalyticsEnabled(value bool) error
	GetProps() map[string]interface{}
	SetProp(k, v string) error
}

func NewRoot(exit chan struct{}, ui UI, config config.Config, analyticsClient AnalyticsClient, analyticsToggle Toggle) *cobra.Command {
	root := &cobra.Command{Use: "cf", SilenceUsage: true, SilenceErrors: true}
	root.PersistentFlags().Bool("help", false, "")
	root.PersistentFlags().Lookup("help").Hidden = true

	daemonWrapper := daemon.NewServiceWrapper(config)

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

	linuxkit := &hypervisor.LinuxKit{Config: config, DaemonRunner: daemonWrapper}
	vpnkit := &network.VpnKit{Config: config, DaemonRunner: daemonWrapper, Label: network.VpnKitLabel}
	metaDataReader := metadata.New()
	analyticsD := &cfanalytics.AnalyticsD{
		Config:       config,
		DaemonRunner: daemonWrapper,
	}
	provisionCmd := &b8.Provision{
		Exit:           exit,
		UI:             ui,
		Provisioner:    provision.NewController(config),
		MetaDataReader: metaDataReader,
		Config:         config,
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
			UI:             ui,
			Version:        config.CliVersion,
			BuildVersion:   config.BuildVersion,
			Config:         config,
			MetaDataReader: metaDataReader,
		},
		&b2.Bosh{
			Exit:      exit,
			UI:        ui,
			Config:    config,
			Analytics: analyticsClient,
		},
		&b3.Catalog{
			UI:     ui,
			Config: config,
		},
		&b4.Download{
			Exit:   exit,
			UI:     ui,
			Config: config,
			Env:    &env.Env{Config: config},
		},
		&b5.Start{
			Exit:            exit,
			LocalExit:       make(chan string, 3),
			UI:              ui,
			Config:          config,
			Cache:           cache,
			Env:             &env.Env{Config: config},
			Analytics:       analyticsClient,
			AnalyticsToggle: analyticsToggle,
			HostNet: &network.HostNet{
				CfdevdClient: cfdevdClient.New("CFD3V", config.CFDevDSocketPath),
			},
			Host: &host.Host{},
			CFDevD: &network.CFDevD{
				ExecutablePath: filepath.Join(config.CacheDir, "cfdevd"),
				TimeSyncSocket: filepath.Join(config.StateLinuxkit, "00000003.0000f3a4"),
			},
			VpnKit:         vpnkit,
			AnalyticsD:     analyticsD,
			Hypervisor:     linuxkit,
			Provisioner:    provision.NewController(config),
			Provision:      provisionCmd,
			MetaDataReader: metaDataReader,
			Stop: &b6.Stop{
				Config:     config,
				Analytics:  analyticsClient,
				Hypervisor: linuxkit,
				HostNet: &network.HostNet{
					CfdevdClient: cfdevdClient.New("CFD3V", config.CFDevDSocketPath),
				},
				Host:         &host.Host{},
				AnalyticsD:   analyticsD,
				VpnKit:       vpnkit,
				CfdevdClient: cfdevdClient.New("CFD3V", config.CFDevDSocketPath),
			},
			Profiler: &profiler.SystemProfiler{},
		},
		&b6.Stop{
			Config:     config,
			Analytics:  analyticsClient,
			Hypervisor: linuxkit,
			HostNet: &network.HostNet{
				CfdevdClient: cfdevdClient.New("CFD3V", config.CFDevDSocketPath),
			},
			Host:         &host.Host{},
			AnalyticsD:   analyticsD,
			VpnKit:       vpnkit,
			CfdevdClient: cfdevdClient.New("CFD3V", config.CFDevDSocketPath),
		},
		&b7.Telemetry{
			Config:          config,
			UI:              ui,
			Analytics:       analyticsClient,
			AnalyticsToggle: analyticsToggle,
			AnalyticsD:      analyticsD,
		},
		provisionCmd,
		&b9.DeployService{
			UI:             ui,
			Provisioner:    provision.NewController(config),
			MetaDataReader: metaDataReader,
			Analytics:      analyticsClient,
			Config:         config,
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
