package bosh

import (
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"os"

	"runtime"

	"code.cloudfoundry.org/cfdev/bosh"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/shell"
	"github.com/spf13/cobra"
)

//go:generate mockgen -package mocks -destination mocks/ui.go code.cloudfoundry.org/cfdev/cmd/bosh UI
type UI interface {
	Say(message string, args ...interface{})
}

//go:generate mockgen -package mocks -destination mocks/analytics_client.go code.cloudfoundry.org/cfdev/cmd/bosh AnalyticsClient
type AnalyticsClient interface {
	Event(event string, data ...map[string]interface{}) error
	PromptOptInIfNeeded(string) error
}

//go:generate mockgen -package mocks -destination mocks/provision.go code.cloudfoundry.org/cfdev/cmd/bosh Provisioner
type Provisioner interface {
	FetchBOSHConfig() (bosh.Config, error)
}

type Bosh struct {
	Exit        chan struct{}
	UI          UI
	StateDir    string
	Provisioner Provisioner
	Analytics   AnalyticsClient
}

func (b *Bosh) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use: "bosh",
		Run: func(cmd *cobra.Command, args []string) {
			if runtime.GOOS != "windows" {
				b.UI.Say(`Usage: eval $(cf dev bosh env)`)
			} else {
				b.UI.Say(`Usage: cf dev bosh env | Invoke-Expression`)
			}
		},
	}
	envCmd := &cobra.Command{
		Use: "env",
		RunE: func(cmd *cobra.Command, args []string) error {
			return b.Env()
		},
	}
	cmd.AddCommand(envCmd)
	return cmd
}

func (b *Bosh) Env() error {
	go func() {
		<-b.Exit
		os.Exit(128)
	}()

	config, err := b.Provisioner.FetchBOSHConfig()
	if err != nil {
		return errors.SafeWrap(err, "failed to fetch bosh configuration")
	}

	b.Analytics.Event(cfanalytics.BOSH_ENV)

	env := shell.Environment{StateDir: b.StateDir}
	shellScript, err := env.Prepare(config)
	if err != nil {
		return errors.SafeWrap(err, "failed to prepare bosh configuration")
	}

	b.UI.Say(shellScript)
	return nil
}
