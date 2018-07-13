package stop_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestStop(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Stop Suite")
}
