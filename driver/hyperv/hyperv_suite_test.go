package hyperv_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHyperv(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hyperv Suite")
}
