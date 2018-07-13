package telemetry

import (
	"code.cloudfoundry.org/cfdev/errors"
	"github.com/spf13/cobra"
)

type UI interface {
	Say(message string, args ...interface{})
}
type Toggle interface {
	Get() bool
	Set(value bool) error
}

type Telemetry struct {
	UI              UI
	AnalyticsToggle Toggle
	Args            struct {
		FlagOff bool
		FlagOn  bool
	}
}

func (t *Telemetry) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "telemetry",
		Short: "Show status for collecting anonymous usage telemetry",
		RunE:  t.RunE,
	}

	cmd.PersistentFlags().BoolVar(&t.Args.FlagOff, "off", false, "Disable the collection of anonymous usage telemetry")
	cmd.PersistentFlags().BoolVar(&t.Args.FlagOn, "on", false, "Enable the collection of anonymous usage telemetry")
	return cmd
}

func (t *Telemetry) RunE(cmd *cobra.Command, args []string) error {
	if t.Args.FlagOff {
		if err := t.AnalyticsToggle.Set(false); err != nil {
			return errors.SafeWrap(err, "turning off telemetry")
		}
	} else if t.Args.FlagOn {
		if err := t.AnalyticsToggle.Set(true); err != nil {
			return errors.SafeWrap(err, "turning on telemetry")
		}
	}

	if t.AnalyticsToggle.Get() {
		t.UI.Say("Telemetry is turned ON")
	} else {
		t.UI.Say("Telemetry is turned OFF")
	}
	return nil
}
