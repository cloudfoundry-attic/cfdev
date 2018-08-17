package version

import (
	"code.cloudfoundry.org/cfdev/semver"
	"github.com/spf13/cobra"
)

type UI interface {
	Say(message string, args ...interface{})
}

type Version struct {
	UI      UI
	Version *semver.Version
}

func (v *Version) Execute() {
	v.UI.Say("Version: %s", v.Version.Original)
}

func (v *Version) Cmd() *cobra.Command {
	return &cobra.Command{
		Use: "version",
		Run: func(_ *cobra.Command, _ []string) {
			v.Execute()
		},
	}
}
