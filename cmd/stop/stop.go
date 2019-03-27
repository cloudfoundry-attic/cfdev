package stop

import (
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/driver"
	"code.cloudfoundry.org/cfdev/errors"
	"github.com/spf13/cobra"
)

//go:generate mockgen -package mocks -destination mocks/analytics.go code.cloudfoundry.org/cfdev/cmd/stop Analytics
type Analytics interface {
	Event(event string, data ...map[string]interface{}) error
}

//go:generate mockgen -package mocks -destination mocks/analyticsd.go code.cloudfoundry.org/cfdev/cmd/stop AnalyticsD
type AnalyticsD interface {
	Stop() error
	Destroy() error
}

type Stop struct {
	Driver       driver.Driver
	Analytics    Analytics
	AnalyticsD   AnalyticsD
}

func (s *Stop) Cmd() *cobra.Command {
	return &cobra.Command{
		Use:  "stop",
		RunE: s.RunE,
	}
}

func (s *Stop) RunE(cmd *cobra.Command, args []string) error {
	s.Analytics.Event(cfanalytics.STOP)

	if err := s.Driver.CheckRequirements(); err != nil {
		return err
	}

	var reterr error

	if err := s.AnalyticsD.Stop(); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop analyticsd")
	}

	if err := s.AnalyticsD.Destroy(); err != nil {
		reterr = errors.SafeWrap(err, "failed to destroy analyticsd")
	}

	if err := s.Driver.Stop(); err != nil {
		reterr = errors.SafeWrap(err, "failed to stop the VM")
	}

	if reterr != nil {
		return errors.SafeWrap(reterr, "cf dev stop")
	}

	return nil
}
