package garden

import (
	gdn "code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
)

type Garden struct {
	Client gdn.Client
}

func New() *Garden {
	return &Garden{
		Client: gdn.New(connection.New("tcp", "localhost:8888")),
	}
}

func (g *Garden) Ping() error {
	return g.Client.Ping()
}
