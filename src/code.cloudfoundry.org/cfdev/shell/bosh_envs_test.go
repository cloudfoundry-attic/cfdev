package shell_test

import (
	"fmt"
	"io/ioutil"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/cfdev/shell"
)

var _ = Describe("Formatting BOSH Configuration", func() {

	It("formats BOSH configuration for eval'ing", func() {
		config := garden.BOSHConfiguration{
			AdminUsername:   "admin",
			AdminPassword:   "admin-password",
			CACertificate:   "ca-certificate",
			DirectorAddress: "10.245.0.2",

			GatewayHost:       "10.245.0.3",
			GatewayUsername:   "jumpbox",
			GatewayPrivateKey: "ssh-private-key",
		}

		expectedExports := []string{
			`export BOSH_ENVIRONMENT="10.245.0.2"`,
			`export BOSH_CLIENT="admin"`,
			`export BOSH_CLIENT_SECRET="admin-password"`,
			`export BOSH_GW_HOST="10.245.0.3"`,
			`export BOSH_GW_USER="jumpbox"`,

			// The following items will be saved to files so we
			// ignore the value for now
			`export BOSH_CA_CERT=`,
			`export BOSH_GW_PRIVATE_KEY=`,
		}

		dir, err := ioutil.TempDir("", "cfdev-state-dir")
		Expect(err).ToNot(HaveOccurred())

		env := shell.Environment{
			StateDir: dir,
		}

		exports, err := env.Prepare(config)
		Expect(err).ShouldNot(HaveOccurred())

		for _, v := range expectedExports {
			Expect(exports).To(ContainSubstring(v))
		}

		ExpectExportToContainFilePathWithContent(exports, "BOSH_GW_PRIVATE_KEY", "ssh-private-key")
		ExpectExportToContainFilePathWithContent(exports, "BOSH_CA_CERT", "ca-certificate")
	})

	Context("unable save files to the state dir", func() {
		It("returns an error", func() {
			env := shell.Environment{
				StateDir: "/some-garbage-directory",
			}

			_, err := env.Prepare(garden.BOSHConfiguration{})
			Expect(err).To(HaveOccurred())
		})
	})
})

func ExpectExportToContainFilePathWithContent(exports, envVar, fileContent string) {
	keyRe := regexp.MustCompile(fmt.Sprintf(`%s="(.*)"`, envVar))
	matches := keyRe.FindStringSubmatch(exports)

	Expect(matches).To(HaveLen(2))
	Expect(matches[1]).To(BeAnExistingFile(), "export "+envVar+" does not point to valid file")

	content, err := ioutil.ReadFile(matches[1])
	Expect(err).ToNot(HaveOccurred())

	Expect(string(content)).To(Equal(fileContent))

}
