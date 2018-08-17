package provision_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestProvision(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Provision Suite")
}
