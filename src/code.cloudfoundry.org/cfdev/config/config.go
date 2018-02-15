package config

import (
	"path/filepath"
	"os"
)

type Config struct {
	BoshDirectorIP string
	CFRouterIP string
	CFDevHome string
	StateDir string
	CacheDir string
	LinuxkitPidFile string
	VpnkitPidFile string
	HyperkitPidFile string
}

func NewConfig() Config{
	cfdevHome := os.Getenv("CFDEV_HOME")
	if cfdevHome == "" {
		cfdevHome = filepath.Join(os.Getenv("HOME"), ".cfdev")
	}
	
	return Config{
		BoshDirectorIP: "10.245.0.2",
		CFRouterIP: "10.144.0.34",
		CFDevHome: cfdevHome,
		StateDir: filepath.Join(cfdevHome, "state"),
		CacheDir: filepath.Join(cfdevHome, "cache"),
		LinuxkitPidFile: filepath.Join(cfdevHome, "state", "linuxkit.pid"),
		VpnkitPidFile: filepath.Join(cfdevHome, "state", "vpnkit.pid"),
		HyperkitPidFile: filepath.Join(cfdevHome, "state", "hyperkit.pid"),
	}
}