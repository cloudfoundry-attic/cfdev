package shell_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/cfdev/shell"
)

var _ = Describe("Formatting BOSH Configuration", func() {

	It("formats BOSH configuration for eval'ing", func() {
		config := garden.BOSHConfiguration{
			AdminUsername: "admin",
			AdminPassword: "admin-password",
			CACertificate: "ca-certificate",
			SSHPrivateKey: "ssh-private-key",
			IPAddress:     "10.245.0.2",
		}

		expected := `export BOSH_ENVIRONMENT="10.245.0.2"
export BOSH_CLIENT="admin"
export BOSH_CLIENT_SECRET="admin-password"
export BOSH_CA_CERT="ca-certificate"
export BOSH_GW_PRIVATE_KEY="ssh-private-key"
`

		exports, err := shell.FormatConfig(config)
		Expect(err).ShouldNot(HaveOccurred())
		Expect(exports).To(Equal(expected))
	})
})
