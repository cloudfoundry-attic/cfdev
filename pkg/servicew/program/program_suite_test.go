package program_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestProgram(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "ServiceWrapper Program Suite")
}
