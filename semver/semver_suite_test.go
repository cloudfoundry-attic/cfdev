package semver_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSemver(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Semver Suite")
}
