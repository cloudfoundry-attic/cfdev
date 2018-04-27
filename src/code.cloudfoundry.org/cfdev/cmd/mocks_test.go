package cmd_test

import (
	"fmt"
	"io"
	"io/ioutil"
)

type MockUI struct {
	WasCalledWith string
}

func (m *MockUI) Say(message string, args ...interface{}) {
	m.WasCalledWith = fmt.Sprintf(message, args...)
}
func (m *MockUI) Writer() io.Writer { return ioutil.Discard }

type MockToggle struct {
	val bool
}

func (t *MockToggle) Get() bool        { return t.val }
func (t *MockToggle) Set(v bool) error { t.val = v; return nil }
