package program

import (
	"code.cloudfoundry.org/cfdev/servicew/config"
	"fmt"
	"github.com/kardianos/service"
	"io"
	"log"
	"os"
	"os/exec"
	"syscall"
	"time"
)

const (
	StatusUnknown = "unknown"
	StatusRunning = "running"
	StatusStopped = "stopped"
)

type Program struct {
	conf    config.Config
	cmd     *exec.Cmd
	logger  *log.Logger
	Service service.Service
}

func New(conf config.Config, out io.Writer) (*Program, error) {
	svConfig := &service.Config{
		Name:        conf.Label,
		DisplayName: conf.Label,
		Description: fmt.Sprintf("CF Dev managed service for '%s'", conf.Label),
		Option:      conf.Options,
	}

	prog := &Program{
		conf:   conf,
		logger: log.New(out, "[PROG-MGR] ", log.LstdFlags),
	}

	svc, err := service.New(prog, svConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize service for '%s': %s", conf.Label, err)
	}

	prog.Service = svc
	return prog, nil
}

func (p *Program) Start(s service.Service) error {
	execPath, err := exec.LookPath(p.conf.Executable)
	if err != nil {
		return fmt.Errorf("failed to find executable '%s': %s", p.conf.Executable, err)
	}

	p.cmd = exec.Command(execPath, p.conf.Args...)
	p.cmd.Env = p.parseEnvs()

	go p.run()
	return nil
}

func (p *Program) run() {
	if p.conf.Log != "" {
		f, err := os.Create(p.conf.Log)
		if err != nil {
			return
		}

		defer f.Close()

		p.cmd.Stdout = f
		p.cmd.Stderr = f
	}

	p.cmd.Run()
}

func (p *Program) Stop(s service.Service) error {
	if p.cmd == nil {
		return nil
	}

	// Invoke a graceful shutdown
	// but then kill if still alive
	p.cmd.Process.Signal(syscall.SIGINT)

	for {
		select {
		case <-time.After(10 * time.Second):
			p.cmd.Process.Signal(syscall.SIGKILL)
		default:
			if p.cmd.ProcessState == nil || p.cmd.ProcessState.Exited() {
				return nil
			}
		}
	}
}

func (p *Program) StartService() error {
	p.logger.Printf("Starting '%s'...\n", p.conf.Label)

	return p.Service.Start()
}

func (p *Program) StopService() error {
	p.logger.Printf("Stopping '%s'...\n", p.conf.Label)

	return p.Service.Stop()
}

func (p *Program) Uninstall() error {
	p.logger.Printf("Uninstalling '%s'...\n", p.conf.Label)

	return p.Service.Uninstall()
}

func (p *Program) Install() error {
	p.logger.Printf("Installing '%s'...\n", p.conf.Label)

	return p.Service.Install()
}

func (p *Program) Status() string {
	status, err := p.Service.Status()
	if err != nil {
		return fmt.Sprintf("error: %s", err)
	}

	switch status {
	case service.StatusRunning:
		return StatusRunning
	case service.StatusStopped:
		return StatusStopped
	default:
		return StatusUnknown
	}
}

func (p *Program) parseEnvs() []string {
	var envs = os.Environ()
	for k, v := range p.conf.Env {
		envs = append(envs, fmt.Sprintf(`%s=%s`, k, v))
	}
	return envs
}
