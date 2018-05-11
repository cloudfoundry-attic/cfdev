package vpnkit_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestVpnkit(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Vpnkit Suite")
}
