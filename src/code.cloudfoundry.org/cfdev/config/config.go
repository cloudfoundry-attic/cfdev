package config

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strconv"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/cfanalytics/toggle"
	"code.cloudfoundry.org/cfdev/resource"
	"code.cloudfoundry.org/cfdev/semver"
	analytics "gopkg.in/segmentio/analytics-go.v3"
)

var (
	cfdepsUrl  string
	cfdepsMd5  string
	cfdepsSize string

	cfdevefiUrl  string
	cfdevefiMd5  string
	cfdevefiSize string

	vpnkitUrl  string
	vpnkitMd5  string
	vpnkitSize string

	hyperkitUrl  string
	hyperkitMd5  string
	hyperkitSize string

	linuxkitUrl  string
	linuxkitMd5  string
	linuxkitSize string

	qcowtoolUrl  string
	qcowtoolMd5  string
	qcowtoolSize string

	uefiUrl  string
	uefiMd5  string
	uefiSize string

	cfdevdUrl  string
	cfdevdMd5  string
	cfdevdSize string

	cliVersion   string
	analyticsKey string
)

type Analytics interface {
	Event(string, map[string]interface{}) error
	Close()
	PromptOptIn(chan struct{}, cfanalytics.UI) error
}

type Toggle interface {
	Get() bool
	Set(value bool) error
}

type Config struct {
	BoshDirectorIP         string
	CFRouterIP             string
	CFDevHome              string
	StateDir               string
	CacheDir               string
	LinuxkitPidFile        string
	VpnkitPidFile          string
	HyperkitPidFile        string
	Dependencies           resource.Catalog
	CFDevDSocketPath       string
	CFDevDInstallationPath string
	CliVersion             *semver.Version
	Analytics              Analytics
	AnalyticsToggle        Toggle
}

func NewConfig() (Config, error) {
	cfdevHome := os.Getenv("CFDEV_HOME")
	if cfdevHome == "" {
		cfdevHome = filepath.Join(os.Getenv("HOME"), ".cfdev")
	}

	catalog, err := catalog()
	if err != nil {
		return Config{}, err
	}

	analyticsToggle := toggle.New(filepath.Join(cfdevHome, "analytics", "analytics.txt"), "optin", "optout")
	analyticsClient, _ := analytics.NewWithConfig(analyticsKey, analytics.Config{
		Logger: analytics.StdLogger(log.New(ioutil.Discard, "", 0)),
	})

	return Config{
		BoshDirectorIP:         "10.245.0.2",
		CFRouterIP:             "10.144.0.34",
		CFDevHome:              cfdevHome,
		StateDir:               filepath.Join(cfdevHome, "state"),
		CacheDir:               filepath.Join(cfdevHome, "cache"),
		LinuxkitPidFile:        filepath.Join(cfdevHome, "state", "linuxkit.pid"),
		VpnkitPidFile:          filepath.Join(cfdevHome, "state", "vpnkit.pid"),
		HyperkitPidFile:        filepath.Join(cfdevHome, "state", "hyperkit.pid"),
		Dependencies:           catalog,
		CFDevDSocketPath:       filepath.Join("/var", "tmp", "cfdevd.socket"),
		CFDevDInstallationPath: filepath.Join("/Library", "PrivilegedHelperTools", "org.cloudfoundry.cfdevd"),
		CliVersion:             semver.Must(semver.New(cliVersion)),
		Analytics:              cfanalytics.New(analyticsToggle, analyticsClient, cliVersion),
		AnalyticsToggle:        analyticsToggle,
	}, nil
}

func (c Config) Close() {
	c.Analytics.Close()
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
			return resource.Catalog{}, fmt.Errorf("Unable to parse CFDEV_CATALOG env variable: %v\n", err)
		}
		return c, nil
	}

	catalog := resource.Catalog{
		Items: []resource.Item{
			{
				URL:  cfdepsUrl,
				Name: "cf-oss-deps.iso",
				MD5:  cfdepsMd5,
				Size: aToUint64(cfdepsSize),
			},
			{
				URL:  cfdevefiUrl,
				Name: "cfdev-efi.iso",
				MD5:  cfdevefiMd5,
				Size: aToUint64(cfdevefiSize),
			},
			{
				URL:  vpnkitUrl,
				Name: "vpnkit",
				MD5:  vpnkitMd5,
				Size: aToUint64(vpnkitSize),
			},
			{
				URL:  hyperkitUrl,
				Name: "hyperkit",
				MD5:  hyperkitMd5,
				Size: aToUint64(hyperkitSize),
			},
			{
				URL:  linuxkitUrl,
				Name: "linuxkit",
				MD5:  linuxkitMd5,
				Size: aToUint64(linuxkitSize),
			},
			{
				URL:  qcowtoolUrl,
				Name: "qcow-tool",
				MD5:  qcowtoolMd5,
				Size: aToUint64(qcowtoolSize),
			},
			{
				URL:  uefiUrl,
				Name: "UEFI.fd",
				MD5:  uefiMd5,
				Size: aToUint64(uefiSize),
			},
			{
				URL:  cfdevdUrl,
				Name: "cfdevd",
				MD5:  cfdevdMd5,
				Size: aToUint64(cfdevdSize),
			},
		},
	}
	sort.Slice(catalog.Items, func(i, j int) bool {
		return catalog.Items[i].Size < catalog.Items[j].Size
	})
	return catalog, nil
}
