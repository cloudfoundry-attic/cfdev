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

//go:generate mockgen -package mocks -destination mocks/analyticsd.go code.cloudfoundry.org/cfdev/cmd/telemetry AnalyticsD
type AnalyticsD interface {
	Start() error
	Stop() error
	Destroy() error
	IsRunning() (bool, error)
}

type Telemetry struct {
	UI              UI
	AnalyticsToggle Toggle
	AnalyticsD      AnalyticsD
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
		isRunning, err := t.AnalyticsD.IsRunning()
		if err != nil {
			return errors.SafeWrap(err, "checking if analyticsd is running")
		}
		if isRunning {
			if err := t.AnalyticsD.Stop(); err != nil {
				return errors.SafeWrap(err, "turning off analyticsd")
			}
			if err := t.AnalyticsD.Destroy(); err != nil {
				return errors.SafeWrap(err, "destroying analyticsd")
			}
		}
	} else if t.Args.FlagOn {
		if err := t.AnalyticsToggle.Set(true); err != nil {
			return errors.SafeWrap(err, "turning on telemetry")
		}
		isRunning, err := t.AnalyticsD.IsRunning()
		if err != nil {
			return errors.SafeWrap(err,"checking if analyticsd is running")
		}
		if !isRunning {
			if err := t.AnalyticsD.Start(); err != nil {
				return errors.SafeWrap(err,"turning on analyticsd")
			}
		}
	}

	if t.AnalyticsToggle.Get() {
		t.UI.Say("Telemetry is turned ON")
	} else {
		t.UI.Say("Telemetry is turned OFF")
	}
	return nil
}
