package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/resource"
)

var (
	cfdepsUrl string
	cfdepsMd5 string

	cfdevefiUrl string
	cfdevefiMd5 string

	vpnkitUrl string
	vpnkitMd5 string

	hyperkitUrl string
	hyperkitMd5 string

	linuxkitUrl string
	linuxkitMd5 string

	qcowtoolUrl string
	qcowtoolMd5 string

	uefiUrl string
	uefiMd5 string
)

type Config struct {
	BoshDirectorIP  string
	CFRouterIP      string
	CFDevHome       string
	StateDir        string
	CacheDir        string
	LinuxkitPidFile string
	VpnkitPidFile   string
	HyperkitPidFile string
	Dependencies    resource.Catalog
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

	return Config{
		BoshDirectorIP:  "10.245.0.2",
		CFRouterIP:      "10.144.0.34",
		CFDevHome:       cfdevHome,
		StateDir:        filepath.Join(cfdevHome, "state"),
		CacheDir:        filepath.Join(cfdevHome, "cache"),
		LinuxkitPidFile: filepath.Join(cfdevHome, "state", "linuxkit.pid"),
		VpnkitPidFile:   filepath.Join(cfdevHome, "state", "vpnkit.pid"),
		HyperkitPidFile: filepath.Join(cfdevHome, "state", "hyperkit.pid"),
		Dependencies:    catalog,
	}, nil
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

	return resource.Catalog{
		Items: []resource.Item{
			{
				URL:  cfdepsUrl,
				Name: "cf-oss-deps.iso",
				MD5:  cfdepsMd5,
			},
			{
				URL:  cfdevefiUrl,
				Name: "cfdev-efi.iso",
				MD5:  cfdevefiMd5,
			},
			{
				URL:  vpnkitUrl,
				Name: "vpnkit",
				MD5:  vpnkitMd5,
			},
			{
				URL:  hyperkitUrl,
				Name: "hyperkit",
				MD5:  hyperkitMd5,
			},
			{
				URL:  linuxkitUrl,
				Name: "linuxkit",
				MD5:  linuxkitMd5,
			},
			{
				URL:  qcowtoolUrl,
				Name: "qcow-tool",
				MD5:  qcowtoolMd5,
			},
			{
				URL:  uefiUrl,
				Name: "UEFI.fd",
				MD5:  uefiMd5,
			},
		},
	}, nil
}
