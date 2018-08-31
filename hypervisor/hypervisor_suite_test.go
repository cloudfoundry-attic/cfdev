package hypervisor_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestHypervisor(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hypervisor Suite")
}
