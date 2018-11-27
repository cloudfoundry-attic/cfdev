package bosh

import (
	"code.cloudfoundry.org/cfdev/config"
	"io/ioutil"
	"path/filepath"
	"strings"

)

type Config struct {
	AdminUsername   string
	AdminPassword   string
	CACertificate   string
	DirectorAddress string

	GatewayHost       string
	GatewayPrivateKey string
	GatewayUsername   string
}

func FetchConfig(cfg config.Config) (Config, error) {
	content, err := ioutil.ReadFile(filepath.Join(cfg.StateBosh, "secret"))
	if err != nil {
		return Config{}, err
	}

	secret := strings.TrimSpace(string(content))

	return Config{
		AdminUsername:     "admin",
		AdminPassword:     secret,
		CACertificate:     filepath.Join(cfg.StateBosh, "ca.crt"),
		DirectorAddress:   cfg.BoshDirectorIP,
		GatewayHost:       cfg.BoshDirectorIP,
		GatewayPrivateKey: filepath.Join(cfg.StateBosh, "jumpbox.key"),
		GatewayUsername:   "jumpbox",
	}, nil
}

func Envs(cfg config.Config) []string {
	 boshConfig, _ := FetchConfig(cfg)

	return []string{
		"BOSH_ENVIRONMENT=" + boshConfig.DirectorAddress,
		"BOSH_CLIENT=" + boshConfig.AdminUsername,
		"BOSH_CLIENT_SECRET=" + boshConfig.AdminPassword,
		"BOSH_CA_CERT=" + boshConfig.CACertificate,
		"BOSH_GW_HOST=" + boshConfig.GatewayHost,
		"BOSH_GW_USER=" + boshConfig.GatewayUsername,
		"BOSH_GW_PRIVATE_KEY=" + boshConfig.GatewayPrivateKey,
		"SERVICES_DIR=" + cfg.ServicesDir,
		"CACHE_DIR=" + cfg.CacheDir,
		"BOSH_STATE=" + cfg.StateBosh,
		"CF_DOMAIN=" + cfg.CFDomain,
	}
}
