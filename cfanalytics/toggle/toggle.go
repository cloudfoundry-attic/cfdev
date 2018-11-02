package toggle

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Toggle struct {
	defined                bool
	CfAnalyticsEnabled     bool `json:"cfAnalyticsEnabled"`
	CustomAnalyticsEnabled bool `json:"customAnalyticsEnabled"`
	path                   string
	props                  map[string]interface{}
}

func New(path string) *Toggle {
	t := &Toggle{
		defined:                false,
		CfAnalyticsEnabled:     false,
		CustomAnalyticsEnabled: false,
		path:                   path,
		props:                  make(map[string]interface{}, 1),
	}

	if txt, err := ioutil.ReadFile(path); err == nil {
		if err := json.Unmarshal(txt, &t); err == nil {
			t.path = path
			t.defined = true
		} else {
			fmt.Printf("Error unmarshalling json: %v", err)
		}
	}

	return t
}

func (t *Toggle) Defined() bool {
	return t.defined
}

func (t *Toggle) CustomAnalyticsDefined() bool {
	if !t.defined {
		return false
	} else {
		return !(t.CfAnalyticsEnabled && !t.CustomAnalyticsEnabled)
	}
}

func (t *Toggle) Enabled() bool {
	return t.CfAnalyticsEnabled || t.CustomAnalyticsEnabled
}

func (t *Toggle) IsCustom() bool {
	return t.CustomAnalyticsEnabled
}

func (t *Toggle) SetCFAnalyticsEnabled(value bool) error {
	t.defined = true
	if !value {
		t.CfAnalyticsEnabled = value
		t.CustomAnalyticsEnabled = value
	} else {
		t.CfAnalyticsEnabled = value
	}

	return t.save()
}

func (t *Toggle) SetCustomAnalyticsEnabled(value bool) error {
	t.defined = true
	t.CfAnalyticsEnabled = value
	t.CustomAnalyticsEnabled = value
	return t.save()
}

func (t *Toggle) GetProps() map[string]interface{} {
	return t.props
}

func (t *Toggle) SetProp(k, v string) error {
	t.props[k] = v
	return t.save()
}

func (t *Toggle) save() error {
	os.MkdirAll(filepath.Dir(t.path), 0755)
	hash := map[string]interface{}{"props": t.props}
	if t.defined {
		hash["cfAnalyticsEnabled"] = t.CfAnalyticsEnabled
		hash["customAnalyticsEnabled"] = t.CustomAnalyticsEnabled
	}
	txt, err := json.Marshal(hash)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(t.path, txt, 0644)
}
