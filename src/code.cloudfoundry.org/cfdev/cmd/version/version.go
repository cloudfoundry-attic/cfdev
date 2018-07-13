package version

import (
	"code.cloudfoundry.org/cfdev/config"
	"github.com/spf13/cobra"
)

type UI interface {
	Say(message string, args ...interface{})
}

type Version struct {
	UI     UI
	Config config.Config
}

func (v *Version) Run(_ *cobra.Command, _rgs []string) {
	v.UI.Say("Version: %s", v.Config.CliVersion.Original)
}

func (v *Version) Cmd() *cobra.Command {
	return &cobra.Command{
		Use: "version",
		Run: v.Run,
	}
}
