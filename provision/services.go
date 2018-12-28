package provision

import (
	"code.cloudfoundry.org/cfdev/bosh"
	"errors"
	"fmt"
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
	var whiteListed []Service

	for _, service := range services {
		if service.Flagname == "always-include" {
			whiteListed = append(whiteListed, service)
		}
	}

	switch strings.TrimSpace(strings.ToLower(whiteList)) {
	case "all":
		return services, nil
	case "none":
		return whiteListed, nil
	case "":
		for _, service := range services {
			if service.DefaultDeploy && !contains(whiteListed, service.Name) {
				whiteListed = append(whiteListed, service)
			}
		}

		return whiteListed, nil
	default:
		for _, service := range services {
			if strings.Contains(strings.ToLower(whiteList), strings.ToLower(service.Flagname)) && !contains(whiteListed, service.Name) {
				whiteListed = append(whiteListed, service)
			}
		}

		return whiteListed, nil
	}
}

func (c *Controller) GetWhiteListedService(serviceName string, whiteList []Service) (*Service, error) {
	for _, service := range whiteList {
		if strings.Contains(strings.ToLower(serviceName), strings.ToLower(service.Flagname)) {
			return &service, nil
		}
	}
	return nil, errors.New(fmt.Sprintf("The service '%s' is not a valid service", serviceName))
}

func contains(services []Service, name string) bool {
	for _, s := range services {
		if s.Name == name {
			return true
		}
	}

	return false
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
