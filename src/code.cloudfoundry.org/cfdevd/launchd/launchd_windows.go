package launchd

import (
	"github.com/kardianos/service"
	"os/exec"
)

type program struct {
	executable string
	args       []string
}

func (p *program) Start(s service.Service) error {
	command := exec.Command(p.executable, p.args...)
	return command.Start()
}

func (p *program) Stop(s service.Service) error {
	return nil
}

func (l *Launchd) AddDaemon(spec DaemonSpec) error {
	srvConfig := &service.Config{
		Name: spec.Label,
	}

	prg := &program{
		executable: spec.Program,
		args:       spec.ProgramArguments,
	}

	s, err := service.New(prg, srvConfig)
	if err != nil {
		return err
	}

	s.Install()
	go s.Run()
	return nil
}

func (l *Launchd) RemoveDaemon(label string) error {
	srvConfig := &service.Config{
		Name: label,
	}

	prg := &program{}
	s, err := service.New(prg, srvConfig)
	if err != nil {
		return err
	}

	return s.Uninstall()
}

func (l *Launchd) Start(label string) error {
	srvConfig := &service.Config{
		Name: label,
	}

	prg := &program{}
	s, err := service.New(prg, srvConfig)
	if err != nil {
		return err
	}

	return s.Start()
}

func (l *Launchd) Stop(label string) error {
	srvConfig := &service.Config{
		Name: label,
	}

	prg := &program{}
	s, err := service.New(prg, srvConfig)
	if err != nil {
		return err
	}

	return s.Stop()
}

func (l *Launchd) IsRunning(label string) (bool, error) {
	return false, nil
}
