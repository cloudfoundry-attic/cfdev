package shell_test

import (
	"fmt"
	"io/ioutil"
	"regexp"

	. "github.com/onsi/gomega"
)

//var _ = Describe("Formatting BOSH Configuration", func() {
//	var config bosh.Config
//	var env shell.Environment
//	BeforeEach(func() {
//		config = bosh.Config{
//			AdminUsername:   "admin",
//			AdminPassword:   "admin-password",
//			CACertificate:   "ca-certificate",
//			DirectorAddress: "10.144.0.2",
//
//			GatewayHost:       "10.245.0.3",
//			GatewayUsername:   "jumpbox",
//			GatewayPrivateKey: "ssh-private-key",
//		}
//
//		env = shell.Environment{}
//	})
//
//	It("formats BOSH configuration for eval'ing", func() {
//		var expectedExports []string
//		if runtime.GOOS != "windows" {
//			expectedExports = []string{
//				`export BOSH_ENVIRONMENT="10.144.0.2";`,
//				`export BOSH_CLIENT="admin";`,
//				`export BOSH_CLIENT_SECRET="admin-password";`,
//				`export BOSH_GW_HOST="10.245.0.3";`,
//				`export BOSH_GW_USER="jumpbox";`,
//
//				// The following items will be saved to files so we
//				// ignore the value for now
//				`export BOSH_CA_CERT="ca-certificate";`,
//				`export BOSH_GW_PRIVATE_KEY="ssh-private-key";`,
//			}
//		} else {
//			expectedExports = []string{
//				`$env:BOSH_ENVIRONMENT="10.144.0.2";`,
//				`$env:BOSH_CLIENT="admin";`,
//				`$env:BOSH_CLIENT_SECRET="admin-password";`,
//				`$env:BOSH_GW_HOST="10.245.0.3";`,
//				`$env:BOSH_GW_USER="jumpbox";`,
//
//				// The following items will be saved to files so we
//				// ignore the value for now
//				`$env:BOSH_CA_CERT="ca-certificate";`,
//				`$env:BOSH_GW_PRIVATE_KEY="ssh-private-key";`,
//			}
//		}
//
//		exports, err := env.Prepare(config)
//		Expect(err).ShouldNot(HaveOccurred())
//
//		for _, v := range expectedExports {
//			Expect(exports).To(ContainSubstring(v))
//		}
//	})
//
//	Context("previous BOSH environment variables are set", func() {
//		BeforeEach(func() {
//			os.Setenv("BOSH_ALL_PROXY", "something")
//			os.Setenv("RANDOM_ENV_FOR_TEST", "something")
//		})
//		AfterEach(func() {
//			os.Unsetenv("BOSH_ALL_PROXY")
//			os.Unsetenv("RANDOM_ENV_FOR_TEST")
//		})
//		It("unsets any other previously set BOSH environment variables", func() {
//			exports, err := env.Prepare(config)
//			Expect(err).ShouldNot(HaveOccurred())
//			if runtime.GOOS != "windows" {
//				Expect(exports).To(MatchRegexp(`(?m)^unset BOSH_ALL_PROXY;$`))
//			} else {
//				Expect(exports).To(MatchRegexp(`(?m)^Remove-Item Env:BOSH_ALL_PROXY;$`))
//			}
//		})
//		It("only unsets BOSH_ALL_PROXY if it is currently set", func() {
//			os.Unsetenv("BOSH_ALL_PROXY")
//			exports, err := env.Prepare(config)
//			Expect(err).ShouldNot(HaveOccurred())
//			Expect(exports).ToNot(ContainSubstring("BOSH_ALL_PROXY;"))
//		})
//		It("does not unset  other environment variables", func() {
//			exports, err := env.Prepare(config)
//			Expect(err).ShouldNot(HaveOccurred())
//			Expect(exports).ToNot(ContainSubstring("RANDOM_ENV_FOR_TEST"))
//		})
//	})
//})

func ExpectExportToContainFilePathWithContent(exports, envVar, fileContent string) {
	keyRe := regexp.MustCompile(fmt.Sprintf(`%s="(.*)"`, envVar))
	matches := keyRe.FindStringSubmatch(exports)

	Expect(matches).To(HaveLen(2))
	Expect(matches[1]).To(BeAnExistingFile(), "export "+envVar+" does not point to valid file")

	content, err := ioutil.ReadFile(matches[1])
	Expect(err).ToNot(HaveOccurred())

	Expect(string(content)).To(Equal(fileContent))

}
