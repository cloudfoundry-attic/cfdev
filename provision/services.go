package provision

import (
	"code.cloudfoundry.org/cfdev/bosh"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

type Service struct {
	Name          string `yaml:"name"`
	Flagname      string `yaml:"flag_name"`
	DefaultDeploy bool   `yaml:"default_deploy"`
	Handle        string `yaml:"handle"` //TODO <-- remove
	Script        string `yaml:"script"`
	Deployment    string `yaml:"deployment"`
	IsErrand      bool   `yaml:"errand"`
}

func (c *Controller) WhiteListServices(whiteList string, services []Service) ([]Service, error) {
	if services == nil {
		return nil, errors.New("Error whitelisting services")
	}

	if strings.ToLower(whiteList) == "all" {
		return services, nil
	}

	var whiteListed []Service

	if whiteList == "none" {
		for _, service := range services {
			if service.Flagname == "always-include" {
				whiteListed = append(whiteListed, service)
			}
		}

		return whiteListed, nil
	}

	if whiteList == "" {
		for _, service := range services {
			if service.DefaultDeploy {
				whiteListed = append(whiteListed, service)
			}
		}

		return whiteListed, nil
	}

	for _, service := range services {
		if (strings.ToLower(whiteList) == strings.ToLower(service.Flagname)) || (strings.ToLower(service.Flagname) == "always-include") {
			whiteListed = append(whiteListed, service)
		}
	}

	return whiteListed, nil
}

func (c *Controller) DeployServices(ui UI, services []Service) error {
	b, err := bosh.New(c.Config)
	if err != nil {
		return err
	}

	errChan := make(chan error, 1)

	for _, service := range services {
		start := time.Now()

		ui.Say("Deploying %s...", service.Name)

		go func(handle string, serviceManifest string) {
			errChan <- c.DeployService(service)
		}(service.Handle, service.Script)

		err = c.report(start, ui, b, service, errChan)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *Controller) DeployService(service Service) error {
	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.Command("powershell.exe", "-ExecutionPolicy", "Bypass", "-File", filepath.Join(c.Config.ServicesDir, service.Script+".ps1"))
	} else {
		cmd = exec.Command(filepath.Join(c.Config.ServicesDir, service.Script))
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, bosh.Envs(c.Config)...)

	logFile, err := os.Create(filepath.Join(c.Config.LogDir, "deploy-"+strings.ToLower(service.Name)+".log"))
	if err != nil {
		return err
	}
	defer logFile.Close()

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	return cmd.Run()
}