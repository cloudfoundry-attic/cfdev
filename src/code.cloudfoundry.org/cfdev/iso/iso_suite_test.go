package iso_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestIso(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Iso Suite")
}
