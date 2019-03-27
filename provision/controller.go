package provision

import (
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/driver"
	"code.cloudfoundry.org/cfdev/workspace"
	"context"
	"github.com/aemengo/bosh-runc-cpi/client"
	"io"
	"time"
)

type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
}

type Controller struct {
	Config    config.Config
	Workspace *workspace.Workspace
}

func NewController(config config.Config) *Controller {
	return &Controller{
		Config:    config,
		Workspace: workspace.New(config),
	}
}

func (c *Controller) Ping(duration time.Duration) error {
	var (
		ticker  = time.NewTicker(time.Second)
		timeout = time.After(duration)
		err     error
	)

	for {
		select {
		case <-ticker.C:
			var ip string
			ip, err = driver.IP(c.Config)
			if err != nil {
				continue
			}

			err = client.Ping(context.Background(), ip+":9999")
			if err == nil {
				return nil
			}
		case <-timeout:
			return err
		}
	}
}
