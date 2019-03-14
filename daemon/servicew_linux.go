package daemon

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/pkg/servicew/client"
	swconfig "code.cloudfoundry.org/cfdev/pkg/servicew/config"
	"path/filepath"
)

type ServiceWrapper struct {
	swc *client.ServiceWrapper
}

func NewServiceWrapper(cfg config.Config) *ServiceWrapper {
	var (
		binaryPath = filepath.Join(cfg.CacheDir, "servicew")
		workdir    = cfg.DaemonDir
		swc        = client.New(binaryPath, workdir)
	)

	return &ServiceWrapper{
		swc: swc,
	}
}

func (s *ServiceWrapper) AddDaemon(spec DaemonSpec) error {
	var (
		definition = swconfig.Config{
			Args:       spec.ProgramArguments,
			Env:        spec.EnvironmentVariables,
			Executable: spec.Program,
			Label:      spec.Label,
			Log:        spec.LogPath,
			Options:    spec.Options,
		}
	)

	return s.swc.Install(definition)
}

func (s *ServiceWrapper) RemoveDaemon(label string) error {
	return s.swc.Uninstall(label)
}

func (s *ServiceWrapper) Start(label string) error {
	return s.swc.Start(label)
}

func (s *ServiceWrapper) Stop(label string) error {
	return s.swc.Stop(label)
}

func (s *ServiceWrapper) IsRunning(label string) (bool, error) {
	return s.swc.IsRunning(label)
}
