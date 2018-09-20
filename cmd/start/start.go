package start

import (
	"io"

	"code.cloudfoundry.org/cfdev/iso"

	"fmt"
	"net/url"
	"os"
	"strings"
	"time"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/provision"
	"code.cloudfoundry.org/cfdev/resource"
	"github.com/spf13/cobra"

	"path/filepath"
	"text/template"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/env"
	"code.cloudfoundry.org/cfdev/hypervisor"
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

//go:generate mockgen -package mocks -destination mocks/host.go code.cloudfoundry.org/cfdev/cmd/start Host
type Host interface {
	CheckRequirements() error
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
	Stop() error
	Watch(chan string)
}

//go:generate mockgen -package mocks -destination mocks/analyticsd.go code.cloudfoundry.org/cfdev/cmd/start AnalyticsD
type AnalyticsD interface {
	Start() error
	Stop() error
	IsRunning() (bool, error)
}

//go:generate mockgen -package mocks -destination mocks/hypervisor.go code.cloudfoundry.org/cfdev/cmd/start Hypervisor
type Hypervisor interface {
	CreateVM(vm hypervisor.VM) error
	Start(vmName string) error
	Stop(vmName string) error
	IsRunning(vmName string) (bool, error)
}

//go:generate mockgen -package mocks -destination mocks/provision.go code.cloudfoundry.org/cfdev/cmd/start Provisioner
type Provisioner interface {
	Ping() error
	DeployBosh() error
	DeployCloudFoundry([]string) error
	GetServices() ([]provision.Service, string, error)
	DeployServices(provision.UI, []provision.Service) error
	ReportProgress(provision.UI, string)
}

//go:generate mockgen -package mocks -destination mocks/isoreader.go code.cloudfoundry.org/cfdev/cmd/start IsoReader
type IsoReader interface {
	Read(isoPath string) (iso.Metadata, error)
}

//go:generate mockgen -package mocks -destination mocks/stop.go code.cloudfoundry.org/cfdev/cmd/start Stop
type Stop interface {
	RunE(cmd *cobra.Command, args []string) error
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
	IsoReader       IsoReader
	Analytics       AnalyticsClient
	AnalyticsToggle Toggle
	HostNet         HostNet
	Host            Host
	Cache           Cache
	CFDevD          CFDevD
	VpnKit          VpnKit
	AnalyticsD      AnalyticsD
	Hypervisor      Hypervisor
	Provisioner     Provisioner
	Stop            Stop
}

const compatibilityVersion = "v1"
const defaultMemory = 4192

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
	pf.IntVarP(&args.Mem, "memory", "m", 0, "memory to allocate to vm in MB")
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
		s.Hypervisor.Stop("cfdev")
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
		if _, err := os.Stat(depsIsoPath); os.IsNotExist(err) {
			return fmt.Errorf("no file found at: %s", depsIsoPath)
		}

		s.Config.Dependencies.Remove("cf-deps.iso")
	}

	s.AnalyticsToggle.SetProp("type", depsIsoName)
	s.Analytics.Event(cfanalytics.START_BEGIN)
	if err := s.Host.CheckRequirements(); err != nil {
		return err
	}

	if running, err := s.Hypervisor.IsRunning("cfdev"); err != nil {
		return errors.SafeWrap(err, "is running")
	} else if running {
		s.UI.Say("CF Dev is already running...")
		s.Analytics.Event(cfanalytics.START_END, map[string]interface{}{"alreadyrunning": true})
		return nil
	}

	if err := s.Stop.RunE(nil,nil); err != nil {
		return errors.SafeWrap(err, "stopping cfdev")
	}

	if err := env.SetupHomeDir(s.Config); err != nil {
		return errors.SafeWrap(err, "setting up cfdev home dir")
	}

	if cfdevd := s.Config.Dependencies.Lookup("cfdevd"); cfdevd != nil {
		s.UI.Say("Downloading Network Helper...")
		if err := s.Cache.Sync(resource.Catalog{
			Items: []resource.Item{*cfdevd},
		}); err != nil {
			return errors.SafeWrap(err, "Unable to download network helper")
		}
		s.Config.Dependencies.Remove("cfdevd")
	}

	if err := s.osSpecificSetup(); err != nil {
		return err
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

	isoConfig, err := s.IsoReader.Read(depsIsoPath)
	if err != nil {
		return errors.SafeWrap(err, fmt.Sprintf("%s is not compatible with CF Dev. Please use a compatible file.", depsIsoName))
	}
	if isoConfig.Version != compatibilityVersion {
		return fmt.Errorf("%s is not compatible with CF Dev. Please use a compatible file", depsIsoName)
	}

	if args.Mem <= 0 {
		if isoConfig.DefaultMemory > 0 {
			args.Mem = isoConfig.DefaultMemory
		} else {
			args.Mem = defaultMemory
		}
	}

	s.UI.Say("Creating the VM...")
	if err := s.Hypervisor.CreateVM(hypervisor.VM{
		Name:     "cfdev",
		CPUs:     args.Cpus,
		MemoryMB: args.Mem,
		DepsIso:  depsIsoPath,
	}); err != nil {
		return errors.SafeWrap(err, "creating the vm")
	}
	s.UI.Say("Starting VPNKit...")
	if err := s.VpnKit.Start(); err != nil {
		return errors.SafeWrap(err, "starting vpnkit")
	}
	s.VpnKit.Watch(s.LocalExit)

	s.UI.Say("Starting the VM...")
	if err := s.Hypervisor.Start("cfdev"); err != nil {
		return errors.SafeWrap(err, "starting the vm")
	}

	s.UI.Say("Waiting for Garden...")
	s.waitForGarden()

	if args.NoProvision {
		s.UI.Say("VM will not be provisioned because '-n' (no-provision) flag was specified.")
		return nil
	}

	if err := s.provision(isoConfig, registries); err != nil {
		return err
	}

	if s.AnalyticsToggle.Get() {
		err = s.AnalyticsD.Start()
	}

	s.Analytics.Event(cfanalytics.START_END)

	return nil
}

func (s *Start) provision(isoConfig iso.Metadata, registries []string) error {
	s.UI.Say("Deploying the BOSH Director...")
	if err := s.Provisioner.DeployBosh(); err != nil {
		return errors.SafeWrap(err, "Failed to deploy the BOSH Director")
	}

	s.UI.Say("Deploying CF...")
	s.Provisioner.ReportProgress(s.UI, "cf")
	if err := s.Provisioner.DeployCloudFoundry(registries); err != nil {
		return errors.SafeWrap(err, "Failed to deploy the Cloud Foundry")
	}

	if err := s.Provisioner.DeployServices(s.UI, isoConfig.Services); err != nil {
		return errors.SafeWrap(err, "Failed to deploy services")
	}

	if isoConfig.Message != "" {
		t := template.Must(template.New("message").Parse(isoConfig.Message))
		err := t.Execute(s.UI.Writer(), map[string]string{"SYSTEM_DOMAIN": "dev.cfdev.sh"})
		if err != nil {
			return errors.SafeWrap(err, "Failed to print deps file provided message")
		}
	}
	return nil
}

func (s *Start) waitForGarden() {
	for {
		if err := s.Provisioner.Ping(); err == nil {
			return
		}
		time.Sleep(time.Second)
	}
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
