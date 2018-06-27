package garden

import (
	"bytes"
	"fmt"

	"code.cloudfoundry.org/cfdev/bosh"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/garden"
	"gopkg.in/yaml.v2"
)

func (g *Garden) FetchBOSHConfig() (bosh.Config, error) {
	containerSpec := garden.ContainerSpec{
		Handle:     "fetch-bosh-config",
		Privileged: true,
		Network:    "10.246.0.0/16",
		Image: garden.ImageRef{
			URI: "/var/vcap/cache/workspace.tar",
		},
		BindMounts: []garden.BindMount{
			{
				SrcPath: "/var/vcap/director",
				DstPath: "/var/vcap/director",
				Mode:    garden.BindMountModeRW,
			},
		},
	}

	container, err := g.Client.Create(containerSpec)
	if err != nil {
		return bosh.Config{}, err
	}

	buffer := &bytes.Buffer{}
	process, err := container.Run(garden.ProcessSpec{
		Path: "cat",
		Args: []string{"/var/vcap/director/creds.yml"},
		User: "root",
	}, garden.ProcessIO{
		Stdout: buffer,
		Stderr: buffer,
	})

	if err != nil {
		return bosh.Config{}, err
	}

	exitCode, err := process.Wait()
	if err != nil {
		return bosh.Config{}, err
	}

	if exitCode != 0 {
		return bosh.Config{}, errors.SafeWrap(nil, fmt.Sprintf("process exited with status %v", exitCode))
	}

	g.Client.Destroy("fetch-bosh-config")

	var resp yamlResponse
	if err := yaml.Unmarshal(buffer.Bytes(), &resp); err != nil {
		return bosh.Config{}, errors.SafeWrap(err, "unable to parse bosh config")
	}

	return resp.convert()
}

type yamlResponse struct {
	AdminPassword string `yaml:"admin_password"`
	DirectorSSL   struct {
		CACertificate string `yaml:"ca"`
	} `yaml:"director_ssl"`
	JumpboxSSH struct {
		PrivateKey string `yaml:"private_key"`
	} `yaml:"jumpbox_ssh"`
}

func (r *yamlResponse) convert() (bosh.Config, error) {
	conf := bosh.Config{}

	if r.AdminPassword == "" {
		return conf, errors.SafeWrap(nil, "admin password was not returned")
	}

	if r.DirectorSSL.CACertificate == "" {
		return conf, errors.SafeWrap(nil, "ca certificate was not returned")
	}

	if r.JumpboxSSH.PrivateKey == "" {
		return conf, errors.SafeWrap(nil, "jumpbox ssh key was not returned")
	}

	conf.DirectorAddress = "10.245.0.2"
	conf.AdminUsername = "admin"
	conf.AdminPassword = r.AdminPassword
	conf.CACertificate = r.DirectorSSL.CACertificate

	conf.GatewayHost = conf.DirectorAddress
	conf.GatewayUsername = "jumpbox"
	conf.GatewayPrivateKey = r.JumpboxSSH.PrivateKey

	return conf, nil
}
