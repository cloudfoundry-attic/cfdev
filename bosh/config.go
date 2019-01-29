package bosh

import (
	"code.cloudfoundry.org/cfdev/config"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

func EnvsMapping(cfg config.Config) map[string]string {
	data, _ := ioutil.ReadFile(filepath.Join(cfg.StateBosh, "env.yml"))

	mapping := map[string]string{}
	yaml.Unmarshal(data, &mapping)

	return mapping
}

func Envs(cfg config.Config) []string {
	var results []string
	for k, v := range EnvsMapping(cfg) {
		results = append(results, k+"="+v)
	}

	return results
}