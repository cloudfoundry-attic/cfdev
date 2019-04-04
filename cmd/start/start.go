package start

import (
	"code.cloudfoundry.org/cfdev/driver"
	"code.cloudfoundry.org/cfdev/workspace"
	"io"
	"time"

	"code.cloudfoundry.org/cfdev/config"
	e "code.cloudfoundry.org/cfdev/errors"
	cfdevos "code.cloudfoundry.org/cfdev/os"
	"code.cloudfoundry.org/cfdev/resource"
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"strings"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"path/filepath"
)

//go:generate mockgen -package mocks -destination mocks/ui.go code.cloudfoundry.org/cfdev/cmd/start UI
type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
}

//go:generate mockgen -package mocks -destination mocks/analytics_client.go code.cloudfoundry.org/cfdev/cmd/start AnalyticsClient
type AnalyticsClient interface {
	Event(event string, data ...map[string]interface{}) error
	PromptOptInIfNeeded(message string) error
}

//go:generate mockgen -package mocks -destination mocks/toggle.go code.cloudfoundry.org/cfdev/cmd/start Toggle
type Toggle interface {
	Enabled() bool
	SetCFAnalyticsEnabled(value bool) error
	SetCustomAnalyticsEnabled(value bool) error
	SetProp(k, v string) error
}

//go:generate mockgen -package mocks -destination mocks/analyticsd.go code.cloudfoundry.org/cfdev/cmd/start AnalyticsD
type AnalyticsD interface {
	Start() error
	Stop() error
	IsRunning() (bool, error)
}

//go:generate mockgen -package mocks -destination mocks/os.go code.cloudfoundry.org/cfdev/cmd/start OS
type OS interface {
	Stats() (cfdevos.Stats, error)
}

//go:generate mockgen -package mocks -destination mocks/provisioner.go code.cloudfoundry.org/cfdev/cmd/start Provisioner
type Provisioner interface {
	Ping(duration time.Duration) error
}

//go:generate mockgen -package mocks -destination mocks/provision.go code.cloudfoundry.org/cfdev/cmd/start Provision
type Provision interface {
	Execute(args Args) error
}

//go:generate mockgen -package mocks -destination mocks/stop.go code.cloudfoundry.org/cfdev/cmd/start Stop
type Stop interface {
	RunE(cmd *cobra.Command, args []string) error
}

//go:generate mockgen -package mocks -destination mocks/env.go code.cloudfoundry.org/cfdev/cmd/start Workspace
type Workspace interface {
	CreateDirs() error
	SetupState(depsFile string) error
	Metadata() (workspace.Metadata, error)
}

//go:generate mockgen -package mocks -destination mocks/cache.go code.cloudfoundry.org/cfdev/cmd/start Cache
type Cache interface {
	Sync(resource.Catalog) error
}

type Args struct {
	Registries          string
	DeploySingleService string
	DepsPath            string
	EFIPath             string
	NoProvision         bool
	Cpus                int
	Mem                 int
}

type Start struct {
	Exit            chan struct{}
	UI              UI
	Config          config.Config
	Analytics       AnalyticsClient
	AnalyticsToggle Toggle
	Cache           Cache
	AnalyticsD      AnalyticsD
	Driver          driver.Driver
	Stop            Stop
	Provisioner     Provisioner
	Provision       Provision
	Workspace       Workspace
	OS              OS
}

const (
	compatibilityVersion = "v5"
	defaultMemory        = 4192
)

func (s *Start) Cmd() *cobra.Command {
	args := Args{}
	cmd := &cobra.Command{
		Use: "start",
		RunE: func(_ *cobra.Command, _ []string) error {
			if err := s.Execute(args); err != nil {
				return e.SafeWrap(err, "cf dev start")
			}
			return nil
		},
	}

	pf := cmd.PersistentFlags()
	pf.StringVarP(&args.DepsPath, "file", "f", "", "path to .dev file containing bosh & cf bits")
	pf.StringVarP(&args.Registries, "registries", "r", "", "docker registries that skip ssl validation - ie. host:port,host2:port2")
	pf.IntVarP(&args.Cpus, "cpus", "c", 4, "cpus to allocate to vm")
	pf.IntVarP(&args.Mem, "memory", "m", 0, "memory to allocate to vm in MB")
	pf.BoolVarP(&args.NoProvision, "no-provision", "n", false, "start vm but do not provision")
	pf.StringVarP(&args.DeploySingleService, "white-listed-services", "s", "", "list of supported services to deploy")
	pf.StringVarP(&args.EFIPath, "efi", "e", filepath.Join(s.Config.BinaryDir, "cfdev-efi-v2.iso"), "path to efi boot iso")

	pf.MarkHidden("no-provision")
	pf.MarkHidden("efi")
	return cmd
}

func (s *Start) Execute(args Args) error {
	go func() {
		<-s.Exit

		s.Driver.Stop()
		os.Exit(128)
	}()

	stats, _ := s.OS.Stats()
	depsPath := filepath.Join(s.Config.CacheDir, "cfdev-deps.tgz")

	if args.DepsPath != "" {
		var err error
		depsPath, err = filepath.Abs(args.DepsPath)
		if err != nil {
			return e.SafeWrap(err, "determining absolute path to deps iso")
		}

		if _, err := os.Stat(depsPath); os.IsNotExist(err) {
			return fmt.Errorf("no file found at: %s", depsPath)
		}

		s.Config.Dependencies.Remove("cfdev-deps.tgz")
	}

	if err := s.Driver.CheckRequirements(); err != nil {
		return err
	}

	if running, err := s.Driver.IsRunning(); err != nil {
		return e.SafeWrap(err, "is running")
	} else if running {
		s.UI.Say("CF Dev is already running...")
		s.Analytics.Event(cfanalytics.START_END, map[string]interface{}{"alreadyrunning": true})
		return nil
	}

	if err := s.Stop.RunE(nil, nil); err != nil {
		return e.SafeWrap(err, "stopping cfdev")
	}

	if err := s.Workspace.CreateDirs(); err != nil {
		return e.SafeWrap(err, "setting up cfdev home dir")
	}

	if cfdevd := s.Config.Dependencies.Lookup("cfdevd"); cfdevd != nil {
		s.UI.Say("Downloading Network Helper...")

		if err := s.Cache.Sync(resource.Catalog{
			Items: []resource.Item{*cfdevd},
		}); err != nil {
			return e.SafeWrap(err, "Unable to download network helper")
		}

		s.Config.Dependencies.Remove("cfdevd")
	}

	err := s.Driver.Prestart()
	if err != nil {
		return e.SafeWrap(err, "Unable to invoke pre-start")
	}

	s.UI.Say("Downloading Resources...")
	if err := s.Cache.Sync(s.Config.Dependencies); err != nil {
		return e.SafeWrap(err, "Unable to sync assets")
	}

	s.UI.Say("Setting State...")
	if err := s.Workspace.SetupState(depsPath); err != nil {
		return e.SafeWrap(err, "Unable to setup directories")
	}

	metaData, err := s.Workspace.Metadata()
	if err != nil {
		return e.SafeWrap(err, fmt.Sprintf("%s is not compatible with CF Dev. Please use a compatible file.", depsPath))
	}

	s.AnalyticsToggle.SetProp("type", metaData.DeploymentName)
	s.AnalyticsToggle.SetProp("artifact", metaData.ArtifactVersion)

	if metaData.Version != compatibilityVersion {
		return fmt.Errorf("%s is not compatible with CF Dev. Please use a compatible file", depsPath)
	}

	s.Analytics.PromptOptInIfNeeded(metaData.AnalyticsMessage)

	s.Analytics.Event(cfanalytics.START_BEGIN, map[string]interface{}{
		"total memory":     stats.TotalMemory,
		"available memory": stats.AvailableMemory,
	})

	if args.DeploySingleService != "" {
		if !s.isServiceSupported(args.DeploySingleService, metaData.Services) {
			return e.SafeWrap(err, fmt.Sprintf("Service: '%v' is not supported", args.DeploySingleService))
		}

		s.Analytics.Event(cfanalytics.SELECTED_SERVICE, map[string]interface{}{"services_requested": args.DeploySingleService})
	}

	memoryToAllocate, err := s.allocateMemory(metaData, stats, args.Mem)
	if err != nil {
		return err
	}

	err = s.Driver.Start(args.Cpus, memoryToAllocate, args.EFIPath)
	if err != nil {
		return err
	}

	s.UI.Say("Waiting for the VM...")
	err = s.Provisioner.Ping(2 * time.Minute)
	if err != nil {
		return e.SafeWrap(err, "Timed out waiting for the VM")
	}

	if args.NoProvision {
		s.UI.Say("VM will not be provisioned because '-n' (no-provision) flag was specified.")
		return nil
	}

	if err := s.Provision.Execute(args); err != nil {
		return err
	}

	if s.AnalyticsToggle.Enabled() {
		err = s.AnalyticsD.Start()
	}

	s.Analytics.Event(cfanalytics.START_END)
	return nil
}

func (s *Start) isServiceSupported(service string, services []workspace.Service) bool {
	if strings.ToLower(service) == "all" || strings.ToLower(service) == "none" {
		return true
	}

	for _, s := range strings.Split(service, ",") {
		if !contains(services, s) {
			return false
		}
	}

	return true
}

func contains(services []workspace.Service, service string) bool {
	for _, s := range services {
		if strings.ToLower(s.Flagname) == strings.ToLower(service) {
			return true
		}
	}

	return false
}

func (s *Start) allocateMemory(metaData workspace.Metadata, stats cfdevos.Stats, requestedMem int) (int, error) {
	baseMem := defaultMemory
	if metaData.DefaultMemory > 0 {
		baseMem = metaData.DefaultMemory
	}

	customMemProvided := requestedMem > 0
	if customMemProvided {
		if requestedMem >= baseMem {
			if stats.AvailableMemory >= uint64(requestedMem) {
				return requestedMem, nil
			}

			if stats.AvailableMemory < uint64(requestedMem) {
				s.UI.Say("WARNING: This machine may not have enough available RAM to run with what is specified.")
				return requestedMem, nil
			}
		}

		if requestedMem < baseMem {
			s.UI.Say(fmt.Sprintf("WARNING: It is recommended that you run %s Dev with at least %v MB of RAM.", strings.ToUpper(metaData.DeploymentName), baseMem))
			if stats.AvailableMemory >= uint64(requestedMem) {
				return requestedMem, nil
			}

			if stats.AvailableMemory < uint64(requestedMem) {
				s.UI.Say("WARNING: This machine may not have enough available RAM to run with what is specified.")
				return requestedMem, nil
			}
		}
	} else {
		if stats.AvailableMemory >= uint64(baseMem) {
			return baseMem, nil
		} else {
			s.UI.Say(fmt.Sprintf("WARNING: %s Dev requires %v MB of RAM to run. This machine may not have enough free RAM.", strings.ToUpper(metaData.DeploymentName), baseMem))
			return baseMem, nil
		}
	}

	return 0, nil
}
