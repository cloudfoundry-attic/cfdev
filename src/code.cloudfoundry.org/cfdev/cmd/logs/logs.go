package logs

import (
	"github.com/spf13/cobra"
	"code.cloudfoundry.org/garden/client/connection"
	"code.cloudfoundry.org/garden/client"
	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/cfdev/errors"
	"path/filepath"
)

type UI interface {
	Say(message string, args ...interface{})
}

type Logs struct {
	UI UI
	Args struct {
		DestDir string
	}
}

func (l *Logs) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "logs",
		RunE: l.RunE,
	}

	cmd.PersistentFlags().StringVarP(&l.Args.DestDir, "dir", "d", ".", "Destination directory")
	cmd.Hidden = true
	return cmd
}

func (l *Logs) RunE(cmd *cobra.Command, args []string) error {
	gClient := client.New(connection.New("tcp", "localhost:8888"))

	err := gdn.FetchLogs(gClient, l.Args.DestDir)
	if err != nil {
		return errors.SafeWrap(err, "failed to fetch cfdev logs")
	}

	dir, _ := filepath.Abs(l.Args.DestDir)

	destinationPath := filepath.Join(dir, gdn.LogsFileName)

	l.UI.Say("Logs downloaded to " + destinationPath)
	return nil
}