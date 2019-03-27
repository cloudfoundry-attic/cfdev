package bosh

import (
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/workspace"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"runtime"

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

type Bosh struct {
	Exit      chan struct{}
	UI        UI
	Config    config.Config
	Analytics AnalyticsClient
	Workspace *workspace.Workspace
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

	b.Analytics.Event(cfanalytics.BOSH_ENV)

	var output []string
	envsMapping := b.Workspace.EnvsMapping()

	for _, envvar := range os.Environ() {
		if strings.HasPrefix(envvar, "BOSH_") {
			envvar = strings.Split(envvar, "=")[0]
			if runtime.GOOS != "windows" {
				output = append(output, fmt.Sprintf("unset %s;", envvar))
			} else {
				output = append(output, fmt.Sprintf("Remove-Item Env:%s;", envvar))
			}
		}
	}

	for key, value := range envsMapping {
		if key == "BOSH_CA_CERT" || key == "BOSH_GW_PRIVATE_KEY" {
			continue
		}

		if runtime.GOOS != "windows" {
			output = append(output, fmt.Sprintf(`export %s=%q;`, key, value))
		} else {
			output = append(output, fmt.Sprintf(`$env:%s=%q;`, key, value))
		}
	}

	if runtime.GOOS == "windows" {
		output = append(output, fmt.Sprintf(`$env:BOSH_CA_CERT=%q;`, filepath.Join(b.Config.StateBosh, "ca.crt")))
		output = append(output, fmt.Sprintf(`$env:BOSH_GW_PRIVATE_KEY=%q;`, filepath.Join(b.Config.StateBosh, "jumpbox.key")))
	} else {
		output = append(output, fmt.Sprintf(`export BOSH_CA_CERT=%q;`, filepath.Join(b.Config.StateBosh, "ca.crt")))
		output = append(output, fmt.Sprintf(`export BOSH_GW_PRIVATE_KEY=%q;`, filepath.Join(b.Config.StateBosh, "jumpbox.key")))
	}

	b.UI.Say(strings.Join(output, "\n"))
	return nil
}
