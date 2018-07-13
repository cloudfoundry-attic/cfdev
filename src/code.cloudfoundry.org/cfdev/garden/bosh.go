package garden

import (
	"fmt"

	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/garden"
)

func (g *Garden) DeployBosh() error {
	containerSpec := garden.ContainerSpec{
		Handle:     "deploy-bosh",
		Privileged: true,
		Network:    "10.246.0.0/16",
		Image: garden.ImageRef{
			URI: "/var/vcap/cache/workspace.tar",
		},
		BindMounts: []garden.BindMount{
			{
				SrcPath: "/var/vcap",
				DstPath: "/var/vcap",
				Mode:    garden.BindMountModeRW,
			},
			{
				SrcPath: "/var/vcap/cache",
				DstPath: "/var/vcap/cache",
				Mode:    garden.BindMountModeRO,
			},
		},
	}

	container, err := g.Client.Create(containerSpec)
	if err != nil {
		return err
	}

	process, err := container.Run(garden.ProcessSpec{
		ID:   "deploy-bosh",
		Path: "/bin/bash",
		Args: []string{"/var/vcap/cache/bin/deploy-bosh"},
		User: "root",
	}, garden.ProcessIO{})

	if err != nil {
		return err
	}

	exitCode, err := process.Wait()
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return errors.SafeWrap(nil, fmt.Sprintf("process exited with status %v", exitCode))
	}

	g.Client.Destroy("deploy-bosh")

	return nil
}
