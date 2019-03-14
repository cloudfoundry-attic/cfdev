package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"code.cloudfoundry.org/cfdev/errors"

	"code.cloudfoundry.org/cfdev/resource"
	"code.cloudfoundry.org/cfdev/semver"
	"runtime"
)

var (
	cfdepsUrl  string
	cfdepsMd5  string
	cfdepsSize string

	cfdevdUrl  string
	cfdevdMd5  string
	cfdevdSize string

	servicewUrl  string
	servicewMd5  string
	servicewSize string

	analyticsdUrl  string
	analyticsdMd5  string
	analyticsdSize string

	analyticsKey     string
	testAnalyticsKey string

	cliVersion   string
	buildVersion string
)

type Config struct {
	BoshDirectorIP         string
	CFRouterIP             string
	HostIP                 string
	CFDevHome              string
	StateDir               string
	StateBosh              string
	StateLinuxkit          string
	CacheDir               string
	BinaryDir              string
	VpnKitStateDir         string
	LogDir                 string
	Dependencies           resource.Catalog
	CFDevDSocketPath       string
	CFDevDInstallationPath string
	CliVersion             *semver.Version
	BuildVersion           string
	AnalyticsKey           string
	ServicesDir            string
	DaemonDir              string
	CFDomain               string
}

func NewConfig() (Config, error) {
	cfdevHome := getCfdevHome()

	catalog, err := catalog()
	if err != nil {
		return Config{}, err
	}

	var analytixKey string
	if os.Getenv("CFDEV_MODE") == "debug" || analyticsKey == "" {
		analytixKey = testAnalyticsKey
	} else {
		analytixKey = analyticsKey
	}

	return Config{
		BoshDirectorIP:         "10.144.0.2",
		CFRouterIP:             "10.144.0.34",
		HostIP:                 "192.168.65.2",
		CFDevHome:              cfdevHome,
		StateDir:               filepath.Join(cfdevHome, "state"),
		StateBosh:              filepath.Join(cfdevHome, "state", "bosh"),
		StateLinuxkit:          filepath.Join(cfdevHome, "state", "linuxkit"),
		VpnKitStateDir:         filepath.Join(cfdevHome, "state", "vpnkit"),
		ServicesDir:            filepath.Join(cfdevHome, "services"),
		DaemonDir:              filepath.Join(cfdevHome, "daemons"),
		CacheDir:               filepath.Join(cfdevHome, "cache"),
		BinaryDir:              filepath.Join(cfdevHome, "bin"),
		LogDir:                 filepath.Join(cfdevHome, "log"),
		Dependencies:           catalog,
		CFDevDSocketPath:       filepath.Join("/var", "tmp", "cfdevd.socket"),
		CFDevDInstallationPath: filepath.Join("/Library", "PrivilegedHelperTools", "org.cloudfoundry.cfdevd"),
		CliVersion:             semver.Must(semver.New(cliVersion)),
		BuildVersion:           buildVersion,
		AnalyticsKey:           analytixKey,
		CFDomain:               "dev.cfdev.sh",
	}, nil
}

func aToUint64(a string) uint64 {
	i, err := strconv.ParseUint(a, 10, 64)
	if err != nil {
		return 0
	}
	return i
}

func catalog() (resource.Catalog, error) {
	override := os.Getenv("CFDEV_CATALOG")

	if override != "" {
		var c resource.Catalog
		if err := json.Unmarshal([]byte(override), &c); err != nil {
			return resource.Catalog{}, errors.SafeWrap(err, "Unable to parse CFDEV_CATALOG env variable")
		}
		return c, nil
	}

	catalog := resource.Catalog{
		Items: []resource.Item{
			{
				URL:   cfdepsUrl,
				Name:  "cfdev-deps.tgz",
				MD5:   cfdepsMd5,
				Size:  aToUint64(cfdepsSize),
				InUse: true,
			},
		},
	}

	switch runtime.GOOS {
	case "darwin":
		catalog.Items = append(catalog.Items,
			resource.Item{
				URL:   analyticsdUrl,
				Name:  "analyticsd",
				MD5:   analyticsdMd5,
				Size:  aToUint64(analyticsdSize),
				InUse: true,
			},
			resource.Item{
				URL:   cfdevdUrl,
				Name:  "cfdevd",
				MD5:   cfdevdMd5,
				Size:  aToUint64(cfdevdSize),
				InUse: true,
			})
	case "windows":
		catalog.Items = append(catalog.Items,
			resource.Item{
				URL:   analyticsdUrl,
				Name:  "analyticsd.exe",
				MD5:   analyticsdMd5,
				Size:  aToUint64(analyticsdSize),
				InUse: true,
			})
	case "linux":
		catalog.Items = append(catalog.Items,
			resource.Item{
				URL:   analyticsdUrl,
				Name:  "analyticsd",
				MD5:   analyticsdMd5,
				Size:  aToUint64(analyticsdSize),
				InUse: true,
			},
			resource.Item{
				URL:   serviceWUrl,
				Name:  "servicew",
				MD5:   serviceWMd5,
				Size:  aToUint64(serviceWSize),
				InUse: true,
			})
	}

	sort.Slice(catalog.Items, func(i, j int) bool {
		return catalog.Items[i].Size < catalog.Items[j].Size
	})

	return catalog, nil
}

func getCfdevHome() string {
	cfdevHome := os.Getenv("CFDEV_HOME")
	if cfdevHome != "" {
		return cfdevHome
	}

	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"), ".cfdev")
	} else {
		return filepath.Join(os.Getenv("HOME"), ".cfdev")
	}
}
