package start

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"
	"time"

	"path/filepath"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/env"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/cfdev/resource"
	"github.com/spf13/cobra"
)

//go:generate mockgen -package mocks -destination mocks/ui.go code.cloudfoundry.org/cfdev/cmd/start UI
type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
}

//go:generate mockgen -package mocks -destination mocks/analytics_client.go code.cloudfoundry.org/cfdev/cmd/start AnalyticsClient
type AnalyticsClient interface {
	Event(event string, data ...map[string]interface{}) error
	PromptOptIn() error
}

//go:generate mockgen -package mocks -destination mocks/toggle.go code.cloudfoundry.org/cfdev/cmd/start Toggle
type Toggle interface {
	Get() bool
	SetProp(k, v string) error
}

//go:generate mockgen -package mocks -destination mocks/network.go code.cloudfoundry.org/cfdev/cmd/start HostNet
type HostNet interface {
	AddLoopbackAliases(...string) error
}

//go:generate mockgen -package mocks -destination mocks/cache.go code.cloudfoundry.org/cfdev/cmd/start Cache
type Cache interface {
	Sync(resource.Catalog) error
}

//go:generate mockgen -package mocks -destination mocks/cfdevd.go code.cloudfoundry.org/cfdev/cmd/start CFDevD
type CFDevD interface {
	Install() error
}

//go:generate mockgen -package mocks -destination mocks/vpnkit.go code.cloudfoundry.org/cfdev/cmd/start VpnKit
type VpnKit interface {
	Start() error
	Stop()
	Watch(chan string)
}

//go:generate mockgen -package mocks -destination mocks/linuxkit.go code.cloudfoundry.org/cfdev/cmd/start LinuxKit
type LinuxKit interface {
	Start(int, int, string) error
	Stop()
	Watch(chan string)
	IsRunning() (bool, error)
}

//go:generate mockgen -package mocks -destination mocks/garden.go code.cloudfoundry.org/cfdev/cmd/start GardenClient
type GardenClient interface {
	Ping() error
	DeployBosh() error
	DeployCloudFoundry([]string) error
	DeployService(string, string) error
	GetServices() ([]garden.Service, string, error)
	ReportProgress(garden.UI, string)
}

type Args struct {
	Registries  string
	DepsIsoPath string
	NoProvision bool
	Cpus        int
	Mem         int
}

type Start struct {
	Exit            chan struct{}
	LocalExit       chan string
	UI              UI
	Config          config.Config
	Analytics       AnalyticsClient
	AnalyticsToggle Toggle
	HostNet         HostNet
	Cache           Cache
	CFDevD          CFDevD
	VpnKit          VpnKit
	LinuxKit        LinuxKit
	GardenClient    GardenClient
}

func (s *Start) Cmd() *cobra.Command {
	args := Args{}
	cmd := &cobra.Command{
		Use: "start",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := s.Execute(args); err != nil {
				return errors.SafeWrap(err, "cf dev start")
			}
			return nil
		},
	}

	pf := cmd.PersistentFlags()
	pf.StringVarP(&args.DepsIsoPath, "file", "f", "", "path to .dev file containing bosh & cf bits")
	pf.StringVarP(&args.Registries, "registries", "r", "", "docker registries that skip ssl validation - ie. host:port,host2:port2")
	pf.IntVarP(&args.Cpus, "cpus", "c", 4, "cpus to allocate to vm")
	pf.IntVarP(&args.Mem, "memory", "m", 4096, "memory to allocate to vm in MB")
	pf.BoolVarP(&args.NoProvision, "no-provision", "n", false, "start vm but do not provision")

	pf.MarkHidden("no-provision")
	return cmd
}

func (s *Start) Execute(args Args) error {
	go func() {
		select {
		case <-s.Exit:
			// no-op
		case name := <-s.LocalExit:
			s.UI.Say("ERROR: %s has stopped", name)
		}
		s.LinuxKit.Stop()
		s.VpnKit.Stop()
		os.Exit(128)
	}()

	depsIsoName := "cf"
	depsIsoPath := filepath.Join(s.Config.CacheDir, "cf-deps.iso")
	if args.DepsIsoPath != "" {
		depsIsoName = filepath.Base(args.DepsIsoPath)
		var err error
		depsIsoPath, err = filepath.Abs(args.DepsIsoPath)
		if err != nil {
			return errors.SafeWrap(err, "determining absolute path to deps iso")
		}
	}
	s.AnalyticsToggle.SetProp("type", depsIsoName)
	s.Analytics.Event(cfanalytics.START_BEGIN)

	if running, err := s.LinuxKit.IsRunning(); err != nil {
		return errors.SafeWrap(err, "is linuxkit running")
	} else if running {
		s.UI.Say("CF Dev is already running...")
		s.Analytics.Event(cfanalytics.START_END, map[string]interface{}{"alreadyrunning": true})
		return nil
	}

	if err := env.Setup(s.Config); err != nil {
		return errors.SafeWrap(err, "environment setup")
	}

	if err := cleanupStateDir(s.Config); err != nil {
		return errors.SafeWrap(err, "cleaning state directory")
	}

	if err := s.HostNet.AddLoopbackAliases(s.Config.BoshDirectorIP, s.Config.CFRouterIP); err != nil {
		return errors.SafeWrap(err, "adding aliases")
	}

	registries, err := s.parseDockerRegistriesFlag(args.Registries)
	if err != nil {
		return errors.SafeWrap(err, "Unable to parse docker registries")
	}

	s.UI.Say("Downloading Resources...")
	if err := s.Cache.Sync(s.Config.Dependencies); err != nil {
		return errors.SafeWrap(err, "Unable to sync assets")
	}

	s.UI.Say("Installing cfdevd network helper...")
	if err := s.CFDevD.Install(); err != nil {
		return errors.SafeWrap(err, "installing cfdevd")
	}

	s.UI.Say("Starting VPNKit...")
	if err := s.VpnKit.Start(); err != nil {
		return errors.SafeWrap(err, "starting vpnkit")
	}
	s.VpnKit.Watch(s.LocalExit)

	s.UI.Say("Starting the VM...")
	if err := s.LinuxKit.Start(args.Cpus, args.Mem, depsIsoPath); err != nil {
		return errors.SafeWrap(err, "starting linuxkit")
	}
	s.LinuxKit.Watch(s.LocalExit)

	s.UI.Say("Waiting for Garden...")
	s.waitForGarden()

	if args.NoProvision {
		s.UI.Say("VM will not be provisioned because '-n' (no-provision) flag was specified.")
		return nil
	}

	s.UI.Say("Deploying the BOSH Director...")
	if err := s.GardenClient.DeployBosh(); err != nil {
		return errors.SafeWrap(err, "Failed to deploy the BOSH Director")
	}

	s.UI.Say("Deploying CF...")
	s.GardenClient.ReportProgress(s.UI, "cf")
	if err := s.GardenClient.DeployCloudFoundry(registries); err != nil {
		return errors.SafeWrap(err, "Failed to deploy the Cloud Foundry")
	}

	services, message, err := s.GardenClient.GetServices()
	if err != nil {
		return errors.SafeWrap(err, "Failed to get list of services to deploy")
	}
	for _, service := range services {
		s.UI.Say("Deploying %s...", service.Name)
		s.GardenClient.ReportProgress(s.UI, service.Deployment)
		if err := s.GardenClient.DeployService(service.Handle, service.Script); err != nil {
			return errors.SafeWrap(err, fmt.Sprintf("Failed to deploy %s", service.Name))
		}
	}

	s.UI.Say(`

	  ██████╗███████╗██████╗ ███████╗██╗   ██╗
	 ██╔════╝██╔════╝██╔══██╗██╔════╝██║   ██║
	 ██║     █████╗  ██║  ██║█████╗  ██║   ██║
	 ██║     ██╔══╝  ██║  ██║██╔══╝  ╚██╗ ██╔╝
	 ╚██████╗██║     ██████╔╝███████╗ ╚████╔╝
	  ╚═════╝╚═╝     ╚═════╝ ╚══════╝  ╚═══╝
	             is now running!

	To begin using CF Dev, please run:
	    cf login -a https://api.v3.pcfdev.io --skip-ssl-validation

	Admin user => Email: admin / Password: admin
	Regular user => Email: user / Password: pass`)

	if message != "" {
		s.UI.Say(message)
	}

	s.Analytics.Event(cfanalytics.START_END)

	return nil
}

func (s *Start) waitForGarden() {
	for {
		if err := s.GardenClient.Ping(); err == nil {
			return
		}
		time.Sleep(time.Second)
	}
}

func cleanupStateDir(cfg config.Config) error {
	for _, dir := range []string{cfg.StateDir, cfg.VpnKitStateDir} {
		if err := os.RemoveAll(dir); err != nil {
			return errors.SafeWrap(err, "Unable to clean up .cfdev state directory")
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.SafeWrap(err, "Unable to create .cfdev state directory")
		}
	}

	return nil
}

func (s *Start) parseDockerRegistriesFlag(flag string) ([]string, error) {
	if flag == "" {
		return nil, nil
	}

	values := strings.Split(flag, ",")

	registries := make([]string, 0, len(values))

	for _, value := range values {
		// Including the // will cause url.Parse to validate 'value' as a host:port
		u, err := url.Parse("//" + value)

		if err != nil {
			// Grab the more succinct error message
			if urlErr, ok := err.(*url.Error); ok {
				err = urlErr.Err
			}
			return nil, fmt.Errorf("'%v' - %v", value, err)
		}
		registries = append(registries, u.Host)
	}
	return registries, nil
}
