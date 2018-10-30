package provision

import (
	"code.cloudfoundry.org/cfdev/config"
	garden "code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
	"context"
	"github.com/aemengo/bosh-runc-cpi/client"
)

type Controller struct {
	Client garden.Client
	Config config.Config
}

func NewController(config config.Config) *Controller {
	return &Controller{
		Client: garden.New(connection.New("tcp", "localhost:8888")),
		Config: config,
	}
}

func (c *Controller) Ping() error {
	ctx := context.Background()
	return client.Ping(ctx, "127.0.0.1:9999")
}

