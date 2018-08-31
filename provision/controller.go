package provision

import (
	garden "code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
)

type Controller struct {
	Client garden.Client
}

func NewController() *Controller {
	return &Controller{
		Client: garden.New(connection.New("tcp", "localhost:8888")),
	}
}

func (c *Controller) Ping() error {
	return c.Client.Ping()
}
