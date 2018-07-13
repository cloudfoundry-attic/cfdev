package toggle

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"path/filepath"
)

type Toggle struct {
	defined bool
	value   bool
	path    string
	props   map[string]interface{}
}

const deprecatedTrueVal = "optin"
const deprecatedFalseVal = "optout"
const keyEnabled = "enabled"

func New(path string) *Toggle {
	t := &Toggle{
		defined: false,
		value:   false,
		path:    path,
		props:   make(map[string]interface{}, 1),
	}
	if txt, err := ioutil.ReadFile(path); err == nil {
		if string(txt) == deprecatedFalseVal {
			t.defined = true
			t.value = false
		} else if string(txt) == deprecatedTrueVal {
			t.defined = true
			t.value = true
		} else {
			data := make(map[string]interface{}, 1)
			if err := json.Unmarshal(txt, &data); err == nil {
				if _, t.defined = data[keyEnabled]; t.defined {
					if v, ok := data[keyEnabled].(bool); ok {
						t.value = v
					}
				}
				if v, ok := data["props"]; ok {
					if props, ok := v.(map[string]interface{}); ok {
						t.props = props
					}
				}
			}
		}
	}
	return t
}

func (t *Toggle) Defined() bool {
	return t.defined
}

func (t *Toggle) Get() bool {
	return t.value
}

func (t *Toggle) Set(value bool) error {
	t.defined = true
	t.value = value
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
		hash["enabled"] = t.value
	}
	txt, err := json.Marshal(hash)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(t.path, txt, 0644)
}
