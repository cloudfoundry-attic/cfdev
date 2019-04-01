package hyperkit_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestHyperkit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Hyperkit Suite")
}
