package provision

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

func (c *Controller) DeployService(service Service) error {
	script := service.Script
	if runtime.GOOS == "windows" {
		script = fmt.Sprintf("%s.ps1", script)
	}

	cmd := exec.Command(filepath.Join(c.Config.ServicesDir, script))

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, c.boshEnvs()...)

	logFile, err := os.Create(filepath.Join(c.Config.LogDir, "deploy-"+strings.ToLower(service.Name)+".log"))
	if err != nil {
		return err
	}
	defer logFile.Close()

	cmd.Stdout = logFile
	cmd.Stderr = logFile

	return cmd.Run()
}

type Service struct {
	Name          string `yaml:"name"`
	Flagname      string `yaml:"flag_name"`
	DefaultDeploy bool   `yaml:"default_deploy"`
	Handle        string `yaml:"handle"` //TODO <-- remove
	Script        string `yaml:"script"`
	Deployment    string `yaml:"deployment"`
	IsErrand      bool   `yaml:"errand"`
}
