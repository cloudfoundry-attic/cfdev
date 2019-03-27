// +build !windows

package bosh_test

import (
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/config"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	cmd "code.cloudfoundry.org/cfdev/cmd/bosh"
	"code.cloudfoundry.org/cfdev/cmd/bosh/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bosh", func() {
	var (
		mockController      *gomock.Controller
		mockUI              *mocks.MockUI
		tmpDir              string
		boshCmd             *cmd.Bosh
		mockAnalyticsClient *mocks.MockAnalyticsClient
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAnalyticsClient = mocks.NewMockAnalyticsClient(mockController)
		mockUI = mocks.NewMockUI(mockController)

		var err error
		tmpDir, err = ioutil.TempDir("", "cfdev-bosh-env-")
		Expect(err).NotTo(HaveOccurred())

		cfg := config.Config{
			StateBosh: tmpDir,
		}

		content := `---
BOSH_ENVIRONMENT: 10.0.0.1
BOSH_CLIENT: admin
BOSH_CA_CERT: |
  some ca cert
  end pem block
BOSH_GW_PRIVATE_KEY: |
  some gw key
  end pem block`

		ioutil.WriteFile(filepath.Join(tmpDir, "env.yml"), []byte(content), 0600)

		boshCmd = &cmd.Bosh{
			UI:        mockUI,
			Analytics: mockAnalyticsClient,
			Config:    cfg,
		}
	})

	AfterEach(func() {
		mockController.Finish()
		os.RemoveAll(tmpDir)
	})

	Describe("Env", func() {
		Context("when the environment has BOSH_* env vars set", func() {
			BeforeEach(func() {
				for _, envvar := range os.Environ() {
					if strings.HasPrefix(envvar, "BOSH_") {
						key := strings.Split(envvar, "=")[0]
						os.Unsetenv(key)
					}
				}
				os.Setenv("BOSH_SOME_VAR", "some-val")
				os.Setenv("BOSH_SOME_OTHER_VAR", "some-other-val")
			})

			It("prints unset and export statements", func() {
				mockAnalyticsClient.EXPECT().Event(cfanalytics.BOSH_ENV)
				mockUI.EXPECT().Say(gomock.Any()).Do(func(arg string) {
					Expect(arg).To(ContainSubstring(`unset BOSH_SOME_VAR;`))
					Expect(arg).To(ContainSubstring(`unset BOSH_SOME_OTHER_VAR;`))
					Expect(arg).To(ContainSubstring(`export BOSH_ENVIRONMENT="10.0.0.1";`))
					Expect(arg).To(ContainSubstring(`export BOSH_CLIENT="admin";`))
				})

				Expect(boshCmd.Env()).To(Succeed())
			})
		})

		It("replaces the certificates with their file paths", func() {
			mockAnalyticsClient.EXPECT().Event(cfanalytics.BOSH_ENV)
			mockUI.EXPECT().Say(gomock.Any()).Do(func(arg string) {
				Expect(strings.Count(arg, "BOSH_CA_CERT")).To(Equal(1))
				Expect(strings.Count(arg, "BOSH_GW_PRIVATE_KEY")).To(Equal(1))

				Expect(arg).To(ContainSubstring(fmt.Sprintf(`export BOSH_CA_CERT="%s";`, filepath.Join(tmpDir, "ca.crt"))))
				Expect(arg).To(ContainSubstring(fmt.Sprintf(`export BOSH_GW_PRIVATE_KEY="%s";`, filepath.Join(tmpDir, "jumpbox.key"))))
			})

			Expect(boshCmd.Env()).To(Succeed())
		})
	})
})
