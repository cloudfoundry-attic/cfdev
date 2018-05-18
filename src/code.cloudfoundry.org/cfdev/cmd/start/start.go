package start

import (
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/cmd/download"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/env"
	"code.cloudfoundry.org/cfdev/errors"
	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/cfdev/network"
	"code.cloudfoundry.org/cfdev/process"
	"code.cloudfoundry.org/cfdev/vpnkit"
	launchdModels "code.cloudfoundry.org/cfdevd/launchd/models"
	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
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
type ProcManager interface {
	SafeKill(pidfile, name string) error
}
type AnalyticsClient interface {
	Event(event string, data ...map[string]interface{}) error
	PromptOptIn() error
}
type Toggle interface {
	Get() bool
	SetProp(k, v string) error
}

type Start struct {
	Exit            chan struct{}
	LocalExit       chan struct{}
	UI              UI
	Config          config.Config
	Launchd         Launchd
	ProcManager     ProcManager
	Analytics       AnalyticsClient
	AnalyticsToggle Toggle
	Args            struct {
		Registries  string
		DepsIsoPath string
		Cpus        int
		Mem         int
	}
}

func (s *Start) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "start",
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := s.RunE(cmd, args); err != nil {
				return errors.SafeWrap(err, "cf dev start")
			}
			return nil
		},
	}

	pf := cmd.PersistentFlags()
	pf.StringVarP(&s.Args.DepsIsoPath, "file", "f", "", "path to .dev file containing bosh & cf bits")
	pf.StringVarP(&s.Args.Registries, "registries", "r", "", "docker registries that skip ssl validation - ie. host:port,host2:port2")
	pf.IntVarP(&s.Args.Cpus, "cpus", "c", 4, "cpus to allocate to vm")
	pf.IntVarP(&s.Args.Mem, "memory", "m", 4096, "memory to allocate to vm in MB")

	return cmd
}

func (s *Start) RunE(_ *cobra.Command, _ []string) error {
	go func() {
		select {
		case <-s.Exit:
			// no-op
		case <-s.LocalExit:
			// no-op
		}
		s.Launchd.Stop(process.LinuxKitLabel)
		s.Launchd.Stop(process.VpnKitLabel)
		s.ProcManager.SafeKill(filepath.Join(s.Config.StateDir, "hyperkit.pid"), "hyperkit")
		os.Exit(128)
	}()

	depsIsoName := "cf"
	if s.Args.DepsIsoPath != "" {
		depsIsoName = filepath.Base(s.Args.DepsIsoPath)
		s.Args.DepsIsoPath, err = filepath.Abs(s.Args.DepsIsoPath)
		if err != nil {
			return errors.SafeWrap(err, "determining absolute path to deps iso")
		}
	}
	s.AnalyticsToggle.SetProp("type", depsIsoName)
	s.Analytics.Event(cfanalytics.START_BEGIN)

	if running, err := s.Launchd.IsRunning(process.LinuxKitLabel); err != nil {
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

	if err := s.setupNetworking(); err != nil {
		return errors.SafeWrap(err, "setting up network")
	}

	registries, err := s.parseDockerRegistriesFlag(s.Args.Registries)
	if err != nil {
		return errors.SafeWrap(err, "Unable to parse docker registries")
	}

	if s.Args.DepsIsoPath != "" {
		item := s.Config.Dependencies.Lookup("cf-deps.iso")
		item.InUse = false
	}

	if err = download.CacheSync(s.Config.Dependencies, s.Config.CacheDir, s.UI.Writer()); err != nil {
		return errors.SafeWrap(err, "downloading")
	}

	if !process.IsCFDevDInstalled(s.Config.CFDevDSocketPath, s.Config.CFDevDInstallationPath, s.Config.Dependencies.Lookup("cfdevd").MD5) {
		if err := process.InstallCFDevD(s.Config.CacheDir); err != nil {
			return errors.SafeWrap(err, "installing cfdevd")
		}
	}

	s.UI.Say("Starting VPNKit ...")
	vpnkit.Start(s.Config, s.Launchd)
	s.watchLaunchd(process.VpnKitLabel)

	s.UI.Say("Starting the VM...")
	linuxKit := process.LinuxKit{
		Config:      s.Config,
		DepsIsoPath: s.Args.DepsIsoPath,
	}
	daemonSpec, err := linuxKit.DaemonSpec(s.Args.Cpus, s.Args.Mem)
	if err != nil {
		return err
	}
	if err := s.Launchd.AddDaemon(daemonSpec); err != nil {
		return errors.SafeWrap(err, "install linuxkit")
	}
	if err := s.Launchd.Start(process.LinuxKitLabel); err != nil {
		return errors.SafeWrap(err, "start linuxkit")
	}
	s.watchLaunchd(process.LinuxKitLabel)

	s.UI.Say("Waiting for Garden...")
	garden := client.New(connection.New("tcp", "localhost:8888"))
	waitForGarden(garden)

	s.UI.Say("Deploying the BOSH Director...")
	if err := gdn.DeployBosh(garden); err != nil {
		return errors.SafeWrap(err, "Failed to deploy the BOSH Director")
	}

	s.UI.Say("Deploying CF...")
	if err := gdn.DeployCloudFoundry(garden, registries); err != nil {
		return errors.SafeWrap(err, "Failed to deploy the Cloud Foundry")
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
Regular user => Email: user / Password: pass
`)

	s.Analytics.Event(cfanalytics.START_END)

	return nil
}

func waitForGarden(client garden.Client) {
	for {
		if err := client.Ping(); err == nil {
			return
		}

		time.Sleep(time.Second)
	}
}

func cleanupStateDir(cfg config.Config) error {
	for _, dir := range []string{cfg.StateDir, cfg.VpnkitStateDir} {
		if err := os.RemoveAll(dir); err != nil {
			return errors.SafeWrap(err, "Unable to clean up .cfdev state directory")
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.SafeWrap(err, "Unable to create .cfdev state directory")
		}
	}

	return nil
}

func (s *Start) setupNetworking() error {
	err := network.AddLoopbackAliases(s.Config.BoshDirectorIP, s.Config.CFRouterIP)

	if err != nil {
		return errors.SafeWrap(err, "Unable to alias BOSH Director/CF Router IP")
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

func (s *Start) watchLaunchd(label string) {
	go func() {
		for {
			running, err := s.Launchd.IsRunning(label)
			if !running && err == nil {
				s.UI.Say("ERROR: %s has stopped", label)
				s.LocalExit <- struct{}{}
				return
			}
			time.Sleep(5 * time.Second)
		}
	}()
}
