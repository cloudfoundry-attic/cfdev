package launchd

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLaunchd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Launchd Suite")
}
