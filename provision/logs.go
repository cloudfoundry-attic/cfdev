package provision

import (
	"io"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/garden"
)

var (
	LogsFileName        = "cfdev-logs.tgz"
	logsContainerHandle = "fetch-logs"
)

func (c *Controller) FetchLogs(destinationDir string) error {
	containerSpec := garden.ContainerSpec{
		Handle:     logsContainerHandle,
		Privileged: true,
		Network:    "10.246.0.0/16",
		Image: garden.ImageRef{
			URI: "/var/vcap/cache/workspace.tar",
		},
		BindMounts: []garden.BindMount{
			{
				SrcPath: "/var/vcap",
				DstPath: "/var/vcap",
				Mode:    garden.BindMountModeRO,
			},
		},
	}

	container, err := c.Client.Create(containerSpec)
	if err != nil {
		return err
	}
	defer c.Client.Destroy(logsContainerHandle)

	tr, err := container.StreamOut(garden.StreamOutSpec{Path: "/var/vcap/logs"})
	if err != nil {
		return err
	}
	defer tr.Close()

	err = os.MkdirAll(destinationDir, os.ModePerm)
	if err != nil {
		return nil
	}

	destinationPath := filepath.Join(destinationDir, LogsFileName)

	f, err := os.Create(destinationPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, tr)
	return err
}
