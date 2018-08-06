package start

import (
	"fmt"
	"net/url"
	"os"
	"strings"
	"text/template"
	"time"

	"path/filepath"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/env"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/resource"
	"github.com/spf13/cobra"
)

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
		s.LinuxKit.Stop()
		s.VpnKit.Stop()
		os.Exit(128)
	}()

	depsIsoName := "cf"
	depsIsoPath := filepath.Join(s.Config.CacheDir, "cf-deps.iso")
	depsToDownload := s.Config.Dependencies
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

		depsToDownload = resource.Catalog{}
		for _, item := range s.Config.Dependencies.Items {
			if item.Name != "cf-deps.iso" {
				depsToDownload.Items = append(depsToDownload.Items, item)
			}
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
	if err := s.Cache.Sync(depsToDownload); err != nil {
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

	err = s.GardenClient.DeployServices(s.UI, isoConfig.Services)
	if err != nil {
		return err
	}

	if isoConfig.Message != "" {
		t := template.Must(template.New("message").Parse(isoConfig.Message))
		err := t.Execute(s.UI.Writer(), map[string]string{"SYSTEM_DOMAIN": "dev.cfdev.sh"})
		if err != nil {
			return errors.SafeWrap(err, "Failed to print deps file provided message")
		}
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
