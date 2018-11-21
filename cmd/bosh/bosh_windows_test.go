package bosh_test

import (
	"code.cloudfoundry.org/cfdev/cfanalytics"
	"io/ioutil"
	"os"
	"strings"

	"code.cloudfoundry.org/cfdev/bosh"
	cmd "code.cloudfoundry.org/cfdev/cmd/bosh"
	"code.cloudfoundry.org/cfdev/cmd/bosh/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Bosh", func() {
	var (
		mockController      *gomock.Controller
		mockProvisioner     *mocks.MockProvisioner
		mockAnalyticsClient *mocks.MockAnalyticsClient
		mockUI              *mocks.MockUI
		tmpDir              string
		boshCmd             *cmd.Bosh
	)

	BeforeEach(func() {
		mockController = gomock.NewController(GinkgoT())
		mockProvisioner = mocks.NewMockProvisioner(mockController)
		mockAnalyticsClient = mocks.NewMockAnalyticsClient(mockController)
		mockUI = mocks.NewMockUI(mockController)

		var err error
		tmpDir, err = ioutil.TempDir("", "cmd-bosh-test")
		Expect(err).NotTo(HaveOccurred())
		boshCmd = &cmd.Bosh{
			Provisioner: mockProvisioner,
			StateDir:    tmpDir,
			UI:          mockUI,
			Analytics:   mockAnalyticsClient,
		}
	})

	AfterEach(func() {
		mockController.Finish()
		os.RemoveAll(tmpDir)
	})

	Describe("Env", func() {
		Context("when no BOSH_* env vars are currently set", func() {
			BeforeEach(func() {
				for _, envvar := range os.Environ() {
					if strings.HasPrefix(envvar, "BOSH_") {
						key := strings.Split(envvar, "=")[0]
						os.Unsetenv(key)
					}
				}
			})

			It("print the export statements", func() {
				mockProvisioner.EXPECT().FetchBOSHConfig().Return(bosh.Config{
					AdminUsername:     "some-admin-username",
					AdminPassword:     "some-admin-password",
					CACertificate:     "some-ca-cert",
					DirectorAddress:   "some-director-address",
					GatewayHost:       "some-gateway-host",
					GatewayPrivateKey: "some-gateway-private-key",
					GatewayUsername:   "some-gateway-username",
				}, nil)

				mockAnalyticsClient.EXPECT().Event(cfanalytics.BOSH_ENV)

				mockUI.EXPECT().Say(`$env:BOSH_ENVIRONMENT="some-director-address";
$env:BOSH_CLIENT="some-admin-username";
$env:BOSH_CLIENT_SECRET="some-admin-password";
$env:BOSH_CA_CERT="some-ca-cert";
$env:BOSH_GW_HOST="some-gateway-host";
$env:BOSH_GW_PRIVATE_KEY="some-gateway-private-key";
$env:BOSH_GW_USER="some-gateway-username";`)
				Expect(boshCmd.Env()).To(Succeed())
			})
		})

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
				mockProvisioner.EXPECT().FetchBOSHConfig().Return(bosh.Config{
					AdminUsername:     "some-admin-username",
					AdminPassword:     "some-admin-password",
					CACertificate:     "some-ca-cert",
					DirectorAddress:   "some-director-address",
					GatewayHost:       "some-gateway-host",
					GatewayPrivateKey: "some-gateway-private-key",
					GatewayUsername:   "some-gateway-username",
				}, nil)

				mockAnalyticsClient.EXPECT().Event(cfanalytics.BOSH_ENV)

				mockUI.EXPECT().Say(`Remove-Item Env:BOSH_SOME_OTHER_VAR;
Remove-Item Env:BOSH_SOME_VAR;
$env:BOSH_ENVIRONMENT="some-director-address";
$env:BOSH_CLIENT="some-admin-username";
$env:BOSH_CLIENT_SECRET="some-admin-password";
$env:BOSH_CA_CERT="some-ca-cert";
$env:BOSH_GW_HOST="some-gateway-host";
$env:BOSH_GW_PRIVATE_KEY="some-gateway-private-key";
$env:BOSH_GW_USER="some-gateway-username";`)
				Expect(boshCmd.Env()).To(Succeed())
			})
		})
	})
})
