package deploy_service

import (
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	e "code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/metadata"
	"code.cloudfoundry.org/cfdev/provision"
	"fmt"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"io"
	"os"
	"path/filepath"
	"time"
)

//go:generate mockgen -package mocks -destination mocks/ui.go code.cloudfoundry.org/cfdev/cmd/deploy-service UI
type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
}

//go:generate mockgen -package mocks -destination mocks/metadata_reader.go code.cloudfoundry.org/cfdev/cmd/deploy-service MetaDataReader
type MetaDataReader interface {
	Read(tarballPath string) (metadata.Metadata, error)
}

//go:generate mockgen -package mocks -destination mocks/provisioner.go code.cloudfoundry.org/cfdev/cmd/deploy-service Provisioner
type Provisioner interface {
	Ping(duration time.Duration) error
	DeployServices(provision.UI, []provision.Service, []string) error
	GetWhiteListedService(string, []provision.Service) (*provision.Service, error)
}

//go:generate mockgen -package mocks -destination mocks/analytics.go code.cloudfoundry.org/cfdev/cmd/stop Analytics
type Analytics interface {
	Event(event string, data ...map[string]interface{}) error
}

const compatibilityVersion = "v4"

type DeployService struct {
	Exit           chan struct{}
	UI             UI
	Provisioner    Provisioner
	MetaDataReader MetaDataReader
	Config         config.Config
	Analytics      Analytics
}

type Args struct {
	Service string
}

func (c *DeployService) Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deploy-service",
		RunE:  c.RunE,
		Short: "Deploy a new service",
		Long:  "Command deploy a new service provided as a parameter",
	}

	return cmd
}

func (c *DeployService) RunE(cmd *cobra.Command, args []string) error {
	go func() {
		<-c.Exit
		os.Exit(128)
	}()

	if len(args) != 1 {
		return errors.New("A service name need to be passed as a argument")
	}

	return c.Execute(Args{
		Service: args[0],
	})
}

func (c *DeployService) Execute(args Args) error {
	metadataConfig, err := c.MetaDataReader.Read(filepath.Join(c.Config.StateDir, "metadata.yml"))
	if err != nil {
		return e.SafeWrap(err, fmt.Sprintf("something went wrong while reading the assets. Please execute 'cf dev start'"))
	}

	if metadataConfig.Version != compatibilityVersion {
		return fmt.Errorf("asset version is incompatible with the current version of the plugin. Please execute 'cf dev start'")
	}

	if c.Provisioner.Ping(10*time.Second) != nil {
		return fmt.Errorf("cf dev is not running. Please execute 'cf dev start'")
	}

	var service *provision.Service
	service, err = c.Provisioner.GetWhiteListedService(args.Service, metadataConfig.Services)
	if err != nil {
		return e.SafeWrap(err, "Failed to whitelist service")
	}

	if err := c.Provisioner.DeployServices(c.UI, []provision.Service{*service}, []string{}); err != nil {
		return e.SafeWrap(err, "Failed to deploy services")
	}

	extraProperties := map[string]interface{}{"name": args.Service}
	c.Analytics.Event(cfanalytics.DEPLOY_SERVICE, extraProperties)

	return nil
}
