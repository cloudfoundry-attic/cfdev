package provision_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestProvision(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Provision Suite")
}
