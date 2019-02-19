package telemetry

import (
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/errors"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"runtime"
)

type UI interface {
	Say(message string, args ...interface{})
}

type Toggle interface {
	Enabled() bool
	SetCustomAnalyticsEnabled(value bool) error
	SetCFAnalyticsEnabled(value bool) error
}

type Analytics interface {
	Event(event string, data ...map[string]interface{}) error
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
	Analytics       Analytics
	AnalyticsToggle Toggle
	AnalyticsD      AnalyticsD
	Config          config.Config
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
		err := t.turnTelemetryOff()
		if err != nil {
			return err
		}
	} else if t.Args.FlagOn {
		if runtime.GOOS == "windows" {
			if _, err := os.Stat(filepath.Join(t.Config.BinaryDir)); os.IsNotExist(err) {
				t.UI.Say("Please run 'cf dev start' before attempting to turn on telemetry")
				return nil
			}
		}

		err := t.turnTelemetryOn()
		if err != nil {
			return err
		}
	}

	if t.AnalyticsToggle.Enabled() {
		t.UI.Say("Telemetry is turned ON")
	} else {
		t.UI.Say("Telemetry is turned OFF")
	}

	return nil
}

func (t *Telemetry) turnTelemetryOff() error {
	t.Analytics.Event(cfanalytics.STOP_TELEMETRY)

	if err := t.AnalyticsToggle.SetCustomAnalyticsEnabled(false); err != nil {
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
	return nil
}

func (t *Telemetry) turnTelemetryOn() error {
	if err := t.AnalyticsToggle.SetCFAnalyticsEnabled(true); err != nil {
		return errors.SafeWrap(err, "turning on telemetry")
	}
	isRunning, err := t.AnalyticsD.IsRunning()
	if err != nil {
		return errors.SafeWrap(err, "checking if analyticsd is running")
	}
	if !isRunning {
		if err := t.AnalyticsD.Start(); err != nil {
			return errors.SafeWrap(err, "turning on analyticsd")
		}
	}
	return nil
}
