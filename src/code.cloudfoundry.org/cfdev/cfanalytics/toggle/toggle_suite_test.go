package toggle_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestToggle(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Toggle Suite")
}
