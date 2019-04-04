package runner

import (
	"os"
	"os/exec"
)

type Sudo struct{}

func (s *Sudo) Run(args ...string) error {
	var (
		invocation = append([]string{"-S"}, args...)
		cmd        = exec.Command("sudo", invocation...)
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
