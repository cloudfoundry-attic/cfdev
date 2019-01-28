package provision

import (
	"code.cloudfoundry.org/cfdev/cmd/start"
	"code.cloudfoundry.org/cfdev/config"
	e "code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/metadata"
	"code.cloudfoundry.org/cfdev/provision"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

//go:generate mockgen -package mocks -destination mocks/ui.go code.cloudfoundry.org/cfdev/cmd/provision UI
type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
}

//go:generate mockgen -package mocks -destination mocks/metadata_reader.go code.cloudfoundry.org/cfdev/cmd/provision MetaDataReader
type MetaDataReader interface {
	Read(tarballPath string) (metadata.Metadata, error)
}

//go:generate mockgen -package mocks -destination mocks/provisioner.go code.cloudfoundry.org/cfdev/cmd/provision Provisioner
type Provisioner interface {
	Ping() error
	DeployBosh() error
	WhiteListServices(string, []provision.Service) ([]provision.Service, error)
	DeployServices(provision.UI, []provision.Service, []string) error
}

const compatibilityVersion = "v4"

type Provision struct {
	Exit           chan struct{}
	UI             UI
	Provisioner    Provisioner
	MetaDataReader MetaDataReader
	Config         config.Config
}

func (c *Provision) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:  "provision",
		RunE: c.RunE,
	}
	cmd.Hidden = true
	return cmd
}

func (c *Provision) RunE(cmd *cobra.Command, args []string) error {
	go func() {
		<-c.Exit
		os.Exit(128)
	}()

	return c.Execute(start.Args{})
}

func (c *Provision) Execute(args start.Args) error {
	metadataConfig, err := c.MetaDataReader.Read(filepath.Join(c.Config.StateDir, "metadata.yml"))
	if err != nil {
		return e.SafeWrap(err, fmt.Sprintf("something went wrong while reading the assets. Please execute 'cf dev start'"))
	}

	if metadataConfig.Version != compatibilityVersion {
		return fmt.Errorf("asset version is incompatible with the current version of the plugin. Please execute 'cf dev start'")
	}

	registries, err := c.parseDockerRegistriesFlag(args.Registries)
	if err != nil {
		return e.SafeWrap(err, "Unable to parse docker registries")
	}

	return c.provision(metadataConfig, registries, args.DeploySingleService)
}

func (c *Provision) provision(metadataConfig metadata.Metadata, registries []string, deploySingleService string) error {
	err := c.Provisioner.Ping()
	if err != nil {
		return e.SafeWrap(err, "VM is not running. Please execute 'cf dev start'")
	}

	c.UI.Say("Deploying the BOSH Director...")
	if err := c.Provisioner.DeployBosh(); err != nil {
		return e.SafeWrap(err, "Failed to deploy the BOSH Director")
	}

	//c.UI.Say("Deploying CF...")
	//if err := c.Provisioner.DeployCloudFoundry(c.UI, registries); err != nil {
	//	return e.SafeWrap(err, "Failed to deploy the Cloud Foundry")
	//}

	services, err := c.Provisioner.WhiteListServices(deploySingleService, metadataConfig.Services)
	if err != nil {
		return e.SafeWrap(err, "Failed to whitelist services")
	}

	if err := c.Provisioner.DeployServices(c.UI, services, registries); err != nil {
		return e.SafeWrap(err, "Failed to deploy services")
	}

	if metadataConfig.Message != "" {
		t := template.Must(template.New("message").Parse(metadataConfig.Message))
		err := t.Execute(c.UI.Writer(), map[string]string{"SYSTEM_DOMAIN": c.Config.CFDomain})
		if err != nil {
			return e.SafeWrap(err, "Failed to print deps file provided message")
		}
	}

	return nil
}

func (c *Provision) parseDockerRegistriesFlag(flag string) ([]string, error) {
	if flag == "" {
		return nil, nil
	}

	values := strings.Split(flag, ",")

	registries := make([]string, 0, len(values))

	for _, value := range values {
		// Including the // will cause url.Parse to validate 'value' as a host:port
		u, err := url.Parse("//" + value)

		if err != nil {
			// Grab the more succinct error message
			if urlErr, ok := err.(*url.Error); ok {
				err = urlErr.Err
			}
			return nil, fmt.Errorf("'%v' - %v", value, err)
		}
		registries = append(registries, u.Host)
	}
	return registries, nil
}
