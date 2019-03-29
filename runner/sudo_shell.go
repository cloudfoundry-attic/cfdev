package runner

import (
	"os"
	"os/exec"
)

type SudoShell struct{}

func (s *SudoShell) Run(args ...string) error {
	var (
		invocation = append([]string{"-S"}, args...)
		cmd        = exec.Command("sudo", invocation...)
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}
