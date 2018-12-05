package provision

import (
	"code.cloudfoundry.org/cfdev/config"
	"context"
	"github.com/aemengo/bosh-runc-cpi/client"
	"io"
)

type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
}

type Controller struct {
	Config config.Config
}

func NewController(config config.Config) *Controller {
	return &Controller{
		Config: config,
	}
}

func (c *Controller) Ping() error {
	ctx := context.Background()
	return client.Ping(ctx, "127.0.0.1:9999")
}
