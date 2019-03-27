package workspace

import (
	"code.cloudfoundry.org/cfdev/config"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"path/filepath"
)

type Workspace struct {
	Config config.Config
}

func New(config config.Config) *Workspace {
	return &Workspace{
		Config: config,
	}
}

func (w *Workspace) EnvsMapping() map[string]string {
	mapping := map[string]string{}

	data, err := ioutil.ReadFile(filepath.Join(w.Config.StateBosh, "env.yml"))
	if err != nil {
		return mapping
	}

	yaml.Unmarshal(data, &mapping)
	return mapping
}

func (w *Workspace) Envs() []string {
	var results []string
	for k, v := range w.EnvsMapping() {
		results = append(results, k+"="+v)
	}

	return results
}