package start

import (
	"io"

	"code.cloudfoundry.org/cfdev/iso"

	"os"

	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/errors"
	"code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/cfdev/process"
	"code.cloudfoundry.org/cfdev/resource"
)

//go:generate mockgen -package mocks -destination mocks/ui.go code.cloudfoundry.org/cfdev/cmd/start UI
type UI interface {
	Say(message string, args ...interface{})
	Writer() io.Writer
}

//go:generate mockgen -package mocks -destination mocks/analytics_client.go code.cloudfoundry.org/cfdev/cmd/start AnalyticsClient
type AnalyticsClient interface {
	Event(event string, data ...map[string]interface{}) error
	PromptOptIn() error
}

//go:generate mockgen -package mocks -destination mocks/toggle.go code.cloudfoundry.org/cfdev/cmd/start Toggle
type Toggle interface {
	Get() bool
	SetProp(k, v string) error
}

//go:generate mockgen -package mocks -destination mocks/network.go code.cloudfoundry.org/cfdev/cmd/start HostNet
type HostNet interface {
	AddLoopbackAliases(...string) error
}

//go:generate mockgen -package mocks -destination mocks/cache.go code.cloudfoundry.org/cfdev/cmd/start Cache
type Cache interface {
	Sync(resource.Catalog) error
}

//go:generate mockgen -package mocks -destination mocks/cfdevd.go code.cloudfoundry.org/cfdev/cmd/start CFDevD
type CFDevD interface {
	Install() error
}

//go:generate mockgen -package mocks -destination mocks/vpnkit.go code.cloudfoundry.org/cfdev/cmd/start VpnKit
type VpnKit interface {
	Start() error
	Stop()
	Watch(chan string)
}

//go:generate mockgen -package mocks -destination mocks/linuxkit.go code.cloudfoundry.org/cfdev/cmd/start LinuxKit
type LinuxKit interface {
	Start(int, int, string) error
	Stop()
	Watch(chan string)
	IsRunning() (bool, error)
}

//go:generate mockgen -package mocks -destination mocks/hyperv.go code.cloudfoundry.org/cfdev/cmd/start HyperV
type HyperV interface {
	Start(vmName string) error
	CreateVM(vm process.VM) error
}

//go:generate mockgen -package mocks -destination mocks/garden.go code.cloudfoundry.org/cfdev/cmd/start GardenClient
type GardenClient interface {
	Ping() error
	DeployBosh() error
	DeployCloudFoundry([]string) error
	GetServices() ([]garden.Service, string, error)
	DeployServices(garden.UI, []garden.Service) error
	ReportProgress(garden.UI, string)
}

//go:generate mockgen -package mocks -destination mocks/isoreader.go code.cloudfoundry.org/cfdev/cmd/start IsoReader
type IsoReader interface {
	Read(isoPath string) (iso.Metadata, error)
}

type Args struct {
	Registries  string
	DepsIsoPath string
	NoProvision bool
	Cpus        int
	Mem         int
}

type Start struct {
	Exit            chan struct{}
	LocalExit       chan string
	UI              UI
	Config          config.Config
	IsoReader       IsoReader
	Analytics       AnalyticsClient
	AnalyticsToggle Toggle
	HostNet         HostNet
	Cache           Cache
	CFDevD          CFDevD
	VpnKit          VpnKit
	HyperV          HyperV
	LinuxKit        LinuxKit
	GardenClient    GardenClient
}

func CleanupStateDir(cfg config.Config) error {
	for _, dir := range []string{cfg.StateDir, cfg.VpnKitStateDir} {
		if err := os.RemoveAll(dir); err != nil {
			return errors.SafeWrap(err, "Unable to clean up .cfdev state directory")
		}
		if err := os.MkdirAll(dir, 0755); err != nil {
			return errors.SafeWrap(err, "Unable to create .cfdev state directory")
		}
	}

	return nil
}
