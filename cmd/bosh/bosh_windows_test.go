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

var _ = XDescribe("Bosh", func() {
	var (
		mockController      *gomock.Controller
		mockAnalyticsClient *mocks.MockAnalyticsClient
		mockUI              *mocks.MockUI
		tmpDir              string
		boshCmd             *cmd.Bosh
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockAnalyticsClient = mocks.NewMockAnalyticsClient(mockController)
		mockUI = mocks.NewMockUI(mockController)

		var err error
		tmpDir, err = ioutil.TempDir("", "cmd-bosh-test")
		Expect(err).NotTo(HaveOccurred())

		cfg := config.Config{
			StateBosh: tmpDir,
			BoshDirectorIP: "10.0.0.1",
		}

		ioutil.WriteFile(filepath.Join(tmpDir, "secret"), []byte("some-bosh-secret"), 0600)

		boshCmd = &cmd.Bosh{
			UI:          mockUI,
			Analytics:   mockAnalyticsClient,
			Config: cfg,
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
				mockUI.EXPECT().Say(fmt.Sprintf(`Remove-Item Env:BOSH_SOME_OTHER_VAR;
Remove-Item Env:BOSH_SOME_VAR;
$env:BOSH_ENVIRONMENT="10.0.0.1";
$env:BOSH_CLIENT="admin";
$env:BOSH_CLIENT_SECRET="some-bosh-secret";
$env:BOSH_CA_CERT="%s";
$env:BOSH_GW_HOST="10.0.0.1";
$env:BOSH_GW_PRIVATE_KEY="%s";
$env:BOSH_GW_USER="jumpbox";`,
					filepath.Join(tmpDir, "ca.crt"),
					filepath.Join(tmpDir, "jumpbox.key"),
				))

				Expect(boshCmd.Env()).To(Succeed())
			})
		})
	})
})
