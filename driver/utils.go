package driver

import (
	"code.cloudfoundry.org/cfdev/config"
	"encoding/json"
	"io/ioutil"
	"path/filepath"
)

func WriteHttpConfig(cfg config.Config) error {
	var (
		httpProxyPath    = filepath.Join(cfg.VpnKitStateDir, "http_proxy.json")
		proxyConfig      = cfg.BuildProxyConfig()
		proxyContents, _ = json.Marshal(proxyConfig)
		httpProxyConfig  = []byte(proxyContents)
	)

	return ioutil.WriteFile(httpProxyPath, httpProxyConfig, 0600)
}
