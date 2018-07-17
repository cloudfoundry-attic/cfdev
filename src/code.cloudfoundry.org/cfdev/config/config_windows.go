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

	winswUrl  string
	winswMd5  string
	winswSize string

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

type Config struct {
	BoshDirectorIP         string
	CFRouterIP             string
	CFDevHome              string
	StateDir               string
	CacheDir               string
	VpnKitStateDir         string
	Dependencies           resource.Catalog
	CFDevDSocketPath       string
	CFDevDInstallationPath string
	CliVersion             *semver.Version
	AnalyticsKey           string
}

func NewConfig() (Config, error) {
	cfdevHome := getCfdevHome()

	catalog, err := catalog()
	if err != nil {
		return Config{}, err
	}

	return Config{
		BoshDirectorIP:         "10.245.0.2",
		CFRouterIP:             "10.144.0.34",
		CFDevHome:              cfdevHome,
		StateDir:               filepath.Join(cfdevHome, "state", "linuxkit"),
		CacheDir:               filepath.Join(cfdevHome, "cache"),
		VpnKitStateDir:         filepath.Join(cfdevHome, "state", "vpnkit"),
		Dependencies:           catalog,
		CFDevDSocketPath:       filepath.Join("/var", "tmp", "cfdevd.socket"),
		CFDevDInstallationPath: filepath.Join("/Library", "PrivilegedHelperTools", "org.cloudfoundry.cfdevd"),
		CliVersion:             semver.Must(semver.New(cliVersion)),
		AnalyticsKey:           analyticsKey,
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
				Name:  "cf-deps.iso",
				MD5:   cfdepsMd5,
				Size:  aToUint64(cfdepsSize),
				InUse: true,
			},
			{
				URL:   cfdevefiUrl,
				Name:  "cfdev-efi.iso",
				MD5:   cfdevefiMd5,
				Size:  aToUint64(cfdevefiSize),
				InUse: true,
			},
			{
				URL:   vpnkitUrl,
				Name:  "vpnkit.exe",
				MD5:   vpnkitMd5,
				Size:  aToUint64(vpnkitSize),
				InUse: true,
			},
			{
				URL:   hyperkitUrl,
				Name:  "hyperkit",
				MD5:   hyperkitMd5,
				Size:  aToUint64(hyperkitSize),
				InUse: true,
			},
			{
				URL:   linuxkitUrl,
				Name:  "linuxkit",
				MD5:   linuxkitMd5,
				Size:  aToUint64(linuxkitSize),
				InUse: true,
			},
			{
				URL:   winswUrl,
				Name:  "winsw.exe",
				MD5:   winswMd5,
				Size:  aToUint64(winswUrl),
				InUse: true,
			},
			{
				URL:   qcowtoolUrl,
				Name:  "qcow-tool",
				MD5:   qcowtoolMd5,
				Size:  aToUint64(qcowtoolSize),
				InUse: true,
			},
			{
				URL:   uefiUrl,
				Name:  "UEFI.fd",
				MD5:   uefiMd5,
				Size:  aToUint64(uefiSize),
				InUse: true,
			},
			{
				URL:   cfdevdUrl,
				Name:  "cfdevd",
				MD5:   cfdevdMd5,
				Size:  aToUint64(cfdevdSize),
				InUse: true,
			},
		},
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
