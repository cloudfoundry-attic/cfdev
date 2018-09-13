package main

import (
	"code.cloudfoundry.org/cfdev/daemon"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func install(programSrc string, args []string) error {
	var(
		lctl        = daemon.New("")
		label       = "org.cloudfoundry.cfdevd"
		program     = "/Library/PrivilegedHelperTools/org.cloudfoundry.cfdevd"
		programArgs = append([]string{program}, args...)
	)

	cfdevdSpec := daemon.DaemonSpec{
		Label:            label,
		Program:          program,
		ProgramArguments: programArgs,
		RunAtLoad:        false,
		Sockets: map[string]string{
			sockName: "/var/tmp/cfdevd.socket",
		},
		StdoutPath: "/var/tmp/cfdevd.stdout.log",
		StderrPath: "/var/tmp/cfdevd.stderr.log",
	}

	isRunning, err := lctl.IsRunning(label)
	if err != nil {
		return fmt.Errorf("checking if analyticsd is running: %s", err)
	}

	if isRunning {
		return nil
	}

	if err := copyExecutable(programSrc, program); err != nil {
		return fmt.Errorf("failed to copy cfdevd: %s", err)
	}

	if err := lctl.AddDaemon(cfdevdSpec); err != nil {
		return fmt.Errorf("failed to install cfdevd: %s", err)
	}

	return nil
}

func copyExecutable(src string, dest string) error {
	if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
		return err
	}

	target, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer target.Close()

	if err = os.Chmod(dest, 0744); err != nil {
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
