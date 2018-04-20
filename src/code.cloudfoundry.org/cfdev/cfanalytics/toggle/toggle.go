package toggle

import (
	"io/ioutil"
	"os"
	"path/filepath"
)

type Toggle struct {
	defined  bool
	value    bool
	path     string
	trueVal  string
	falseVal string
}

func New(path, trueVal, falseVal string) *Toggle {
	t := &Toggle{
		defined:  false,
		value:    false,
		path:     path,
		trueVal:  trueVal,
		falseVal: falseVal,
	}
	if txt, err := ioutil.ReadFile(path); err == nil {
		t.defined = len(txt) > 0
		t.value = string(txt) == trueVal
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
	os.MkdirAll(filepath.Dir(t.path), 0755)
	if value {
		return ioutil.WriteFile(t.path, []byte(t.trueVal), 0644)
	} else {
		return ioutil.WriteFile(t.path, []byte(t.falseVal), 0644)
	}
}
