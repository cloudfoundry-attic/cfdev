package stop_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"

	"errors"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/cmd/stop"
	"code.cloudfoundry.org/cfdev/cmd/stop/mocks"
	"code.cloudfoundry.org/cfdev/config"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Stop", func() {
	var (
		cfg              config.Config
		stopCmd          *cobra.Command
		mockCfdevdClient *mocks.MockCfdevdClient
		mockAnalytics    *mocks.MockAnalytics
		mockHostNet      *mocks.MockHostNet
		mockHypervisor   *mocks.MockHypervisor
		mockVpnkit       *mocks.MockVpnKit
		mockController   *gomock.Controller
		stateDir         string
		err              error
	)

	BeforeEach(func() {
		stateDir, err = ioutil.TempDir(os.Getenv("TMPDIR"), "state-dir")
		Expect(err).NotTo(HaveOccurred())

		cfg = config.Config{
			StateDir:       stateDir,
			CFRouterIP:     "some-cf-router-ip",
			BoshDirectorIP: "some-bosh-director-ip",
		}

		mockController = gomock.NewController(GinkgoT())
		mockCfdevdClient = mocks.NewMockCfdevdClient(mockController)
		mockAnalytics = mocks.NewMockAnalytics(mockController)
		mockHostNet = mocks.NewMockHostNet(mockController)
		mockHypervisor = mocks.NewMockHypervisor(mockController)
		mockVpnkit = mocks.NewMockVpnKit(mockController)

		subject := &stop.Stop{
			Hypervisor:   mockHypervisor,
			VpnKit:       mockVpnkit,
			Config:       cfg,
			Analytics:    mockAnalytics,
			CfdevdClient: mockCfdevdClient,
			HostNet:      mockHostNet,
		}
		stopCmd = subject.Cmd()
		stopCmd.SetArgs([]string{})
		stopCmd.SetOutput(GinkgoWriter)
	})

	AfterEach(func() {
		mockController.Finish()
		os.RemoveAll(stateDir)
	})

	It("uninstalls linuxkit, vpnkit, and cfdevd, tears down aliases, and sends analytics event", func() {
		mockAnalytics.EXPECT().Event(cfanalytics.STOP)
		mockHypervisor.EXPECT().Stop("cfdev")
		mockHypervisor.EXPECT().Destroy("cfdev")
		mockVpnkit.EXPECT().Stop()
		mockVpnkit.EXPECT().Destroy()

		if runtime.GOOS == "darwin" {
			mockCfdevdClient.EXPECT().Uninstall()
		}
		mockHostNet.EXPECT().RemoveLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip")

		Expect(stopCmd.Execute()).To(Succeed())
	})

	Context("stopping linuxkit fails", func() {
		It("stops the others and returns linuxkit error", func() {
			mockAnalytics.EXPECT().Event(cfanalytics.STOP)
			mockHypervisor.EXPECT().Stop("cfdev").Return(errors.New("test"))
			mockHypervisor.EXPECT().Destroy("cfdev")
			mockVpnkit.EXPECT().Stop()
			mockVpnkit.EXPECT().Destroy()
			if runtime.GOOS == "darwin" {
				mockCfdevdClient.EXPECT().Uninstall()
			}
			mockHostNet.EXPECT().RemoveLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip")

			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to stop linuxkit: test"))
		})
	})

	Context("destroying linuxkit fails", func() {
		It("stops the others and returns linuxkit error", func() {
			mockAnalytics.EXPECT().Event(cfanalytics.STOP)
			mockHypervisor.EXPECT().Stop("cfdev")
			mockHypervisor.EXPECT().Destroy("cfdev").Return(errors.New("test"))
			mockVpnkit.EXPECT().Stop()
			mockVpnkit.EXPECT().Destroy()
			if runtime.GOOS == "darwin" {
				mockCfdevdClient.EXPECT().Uninstall()
			}
			mockHostNet.EXPECT().RemoveLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip")

			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to destroy linuxkit: test"))
		})
	})

	Context("stopping vpnkit fails", func() {
		It("stops the others and returns vpnkit error", func() {
			mockAnalytics.EXPECT().Event(cfanalytics.STOP)
			mockHypervisor.EXPECT().Stop("cfdev")
			mockHypervisor.EXPECT().Destroy("cfdev")
			mockVpnkit.EXPECT().Stop().Return(errors.New("test"))
			mockVpnkit.EXPECT().Destroy()

			if runtime.GOOS == "darwin" {
				mockCfdevdClient.EXPECT().Uninstall()
			}
			mockHostNet.EXPECT().RemoveLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip")

			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to stop vpnkit: test"))
		})
	})

	Context("destroying vpnkit fails", func() {
		It("stops the others and returns vpnkit error", func() {
			mockAnalytics.EXPECT().Event(cfanalytics.STOP)
			mockHypervisor.EXPECT().Stop("cfdev")
			mockHypervisor.EXPECT().Destroy("cfdev")
			mockVpnkit.EXPECT().Stop()
			mockVpnkit.EXPECT().Destroy().Return(errors.New("test"))
			if runtime.GOOS == "darwin" {
				mockCfdevdClient.EXPECT().Uninstall()
			}
			mockHostNet.EXPECT().RemoveLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip")

			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to destroy vpnkit: test"))
		})
	})

	Context("removing aliases fails", func() {
		It("stops the others and returns alias error", func() {
			mockAnalytics.EXPECT().Event(cfanalytics.STOP)
			mockHypervisor.EXPECT().Stop("cfdev")
			mockHypervisor.EXPECT().Destroy("cfdev")
			mockVpnkit.EXPECT().Stop()
			mockVpnkit.EXPECT().Destroy()
			if runtime.GOOS == "darwin" {
				mockCfdevdClient.EXPECT().Uninstall()
			}
			mockHostNet.EXPECT().RemoveLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip").Return(fmt.Errorf("test"))

			Expect(stopCmd.Execute()).To(MatchError(`cf dev stop: failed to remove IP aliases: test`))
		})
	})
})
