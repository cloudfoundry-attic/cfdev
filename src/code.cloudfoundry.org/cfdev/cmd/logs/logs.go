package logs

import (
	"path/filepath"

	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/provision"
	"github.com/spf13/cobra"
)

//go:generate mockgen -package mocks -destination mocks/ui.go code.cloudfoundry.org/cfdev/cmd/logs UI
type UI interface {
	Say(message string, args ...interface{})
}

//go:generate mockgen -package mocks -destination mocks/provisioner.go code.cloudfoundry.org/cfdev/cmd/logs Provisioner
type Provisioner interface {
	FetchLogs(string) error
}

type Logs struct {
	UI          UI
	Provisioner Provisioner
}

type Args struct {
	DestDir string
}

func (l *Logs) Cmd() *cobra.Command {
	args := Args{}
	cmd := &cobra.Command{
		Use: "logs",
		RunE: func(_ *cobra.Command, _ []string) error {
			return l.Logs(args)
		},
	}
	cmd.PersistentFlags().StringVarP(&args.DestDir, "dir", "d", ".", "Destination directory")
	cmd.Hidden = true
	return cmd
}

func (l *Logs) Logs(args Args) error {
	err := l.Provisioner.FetchLogs(args.DestDir)
	if err != nil {
		return errors.SafeWrap(err, "failed to fetch cfdev logs")
	}

	dir, _ := filepath.Abs(args.DestDir)
	destinationPath := filepath.Join(dir, provision.LogsFileName)

	l.UI.Say("Logs downloaded to " + destinationPath)
	return nil
}
