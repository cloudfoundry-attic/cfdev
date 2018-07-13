package start_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestStart(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Start Suite")
}
