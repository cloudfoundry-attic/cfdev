package daemon

import (
	"code.cloudfoundry.org/cfdev/config"
	swconfig "code.cloudfoundry.org/cfdev/servicew/config"
	"code.cloudfoundry.org/cfdev/servicew/program"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ServiceWrapper struct {
	cfg config.Config
}

func NewServiceWrapper(cfg config.Config) *ServiceWrapper {
	return &ServiceWrapper{
		cfg: cfg,
	}
}

func (s *ServiceWrapper) AddDaemon(spec DaemonSpec) error {
	var (
		swrapperPath     = s.swrapperPath(spec.Label)
		definitionConfig = swrapperPath + ".yml"
		definition       = swconfig.Config{
			Args:       spec.ProgramArguments,
			Env:        spec.EnvironmentVariables,
			Executable: spec.Program,
			Label:      spec.Label,
			Log:        spec.LogPath,
			Options:    spec.Options,
		}
	)

	// TODO this is probably gonna be the cache directory
	err := copyBinary(filepath.Join(s.cfg.BinaryDir, "servicew"), swrapperPath)
	if err != nil {
		return err
	}

	f, err := os.Create(definitionConfig)
	if err != nil {
		return err
	}
	defer f.Close()

	err = yaml.NewEncoder(f).Encode(definition)
	if err != nil {
		return err
	}

	command := exec.Command(swrapperPath, "install")
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to install '%s': %s: %s", spec.Label, err, output)
	}

	return nil
}

func (s *ServiceWrapper) RemoveDaemon(label string) error {
	var (
		swrapperPath     = s.swrapperPath(label)
		definitionConfig = swrapperPath + ".yml"
	)

	if s.swrapperNotExist(label) {
		return nil
	}

	command := exec.Command(swrapperPath, "uninstall")
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to uninstall '%s': %s: %s", label, err, output)
	}

	err = os.RemoveAll(swrapperPath)
	if err != nil {
		return err
	}

	return os.RemoveAll(definitionConfig)
}

func (s *ServiceWrapper) Start(label string) error {
	command := exec.Command(s.swrapperPath(label), "start")
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start '%s': %s: %s", label, err, output)
	}

	return nil
}

func (s *ServiceWrapper) Stop(label string) error {
	if s.swrapperNotExist(label) {
		return nil
	}

	command := exec.Command(s.swrapperPath(label), "stop")
	output, err := command.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to stop '%s': %s: %s", label, err, output)
	}

	return nil
}

func (s *ServiceWrapper) IsRunning(label string) (bool, error) {
	if s.swrapperNotExist(label) {
		return false, nil
	}

	command := exec.Command(s.swrapperPath(label), "status")
	output, err := command.Output()
	if err != nil {
		return false, fmt.Errorf("failed to fetch status of '%s': %s: %s", label, err, output)
	}

	return strings.TrimSpace(string(output)) == program.StatusRunning, nil
}

func (s *ServiceWrapper) swrapperNotExist(label string) bool {
	_, err := os.Stat(s.swrapperPath(label))
	return os.IsNotExist(err)
}

func (s *ServiceWrapper) swrapperPath(label string) string {
	splits := strings.Split(label, ".")
	return filepath.Join(s.cfg.DaemonDir, splits[len(splits) - 1])
}

func copyBinary(src string, dest string) error {
	target, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer target.Close()

	err = os.Chmod(dest, 0744)
	if err != nil {
		return err
	}

	binData, err := os.Open(src)
	if err != nil {
		return err
	}
	defer binData.Close()

	_, err = io.Copy(target, binData)
	return err
}
