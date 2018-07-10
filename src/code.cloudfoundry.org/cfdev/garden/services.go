package garden

import (
	"archive/tar"
	"fmt"
	"io"
	"io/ioutil"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/garden"
	yaml "gopkg.in/yaml.v2"
)

func (g *Garden) DeployService(handle, script string) error {
	container, err := g.Client.Create(containerSpec(handle))
	if err != nil {
		return err
	}

	process, err := container.Run(garden.ProcessSpec{
		ID:   handle,
		Path: "/bin/bash",
		Args: []string{filepath.Join("/var/vcap/cache", script)},
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
		return errors.SafeWrap(nil, fmt.Sprintf("process exited with status %d", exitCode))
	}

	g.Client.Destroy(handle)

	return nil
}

type Service struct {
	Name       string `yaml:"name"`
	Handle     string `yaml:"handle"`
	Script     string `yaml:"script"`
	Deployment string `yaml:"deployment"`
}

func (g *Garden) GetServices() ([]Service, string, error) {
	container, err := g.Client.Create(containerSpec("get-services"))
	if err != nil {
		return nil, "", err
	}
	r, err := container.StreamOut(garden.StreamOutSpec{Path: "/var/vcap/cache/metadata.yml"})
	if err != nil {
		return nil, "", err
	}
	defer r.Close()
	tr := tar.NewReader(r)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return nil, "", err
		}
		b, err := ioutil.ReadAll(tr)
		if err != nil {
			return nil, "", fmt.Errorf("parsing %s: %s", hdr.Name, err)
		}
		services := struct {
			Message  string    `yaml:"splash_message"`
			Services []Service `yaml:"services"`
		}{}
		err = yaml.Unmarshal(b, &services)
		return services.Services, services.Message, err
	}
	return nil, "", fmt.Errorf("metadata.yml not found in container")
}

func containerSpec(handle string) garden.ContainerSpec {
	return garden.ContainerSpec{
		Handle:     handle,
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
}
