package stop_test

import (
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

var _ = XDescribe("Stop", func() {
	var (
		cfg              config.Config
		stopCmd          *cobra.Command
		mockCfdevdClient *mocks.MockCfdevdClient
		mockAnalytics    *mocks.MockAnalytics
		mockHostNet      *mocks.MockHostNet
		mockHost         *mocks.MockHost
		mockHypervisor   *mocks.MockHypervisor
		mockAnalyticsD   *mocks.MockAnalyticsD
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
		mockHost = mocks.NewMockHost(mockController)
		mockAnalyticsD = mocks.NewMockAnalyticsD(mockController)
		mockHypervisor = mocks.NewMockHypervisor(mockController)
		mockVpnkit = mocks.NewMockVpnKit(mockController)

		subject := &stop.Stop{
			Hypervisor:   mockHypervisor,
			VpnKit:       mockVpnkit,
			Config:       cfg,
			Analytics:    mockAnalytics,
			CfdevdClient: mockCfdevdClient,
			AnalyticsD:   mockAnalyticsD,
			HostNet:      mockHostNet,
			Host:         mockHost,
		}
		stopCmd = subject.Cmd()
		stopCmd.SetArgs([]string{})
		stopCmd.SetOutput(GinkgoWriter)
	})

	AfterEach(func() {
		mockController.Finish()
		os.RemoveAll(stateDir)
	})

	It("destroys the VM, uninstalls vpnkit, analyticsd, and cfdevd, tears down aliases, and sends analytics event", func() {
		mockAnalytics.EXPECT().Event(cfanalytics.STOP)
		mockHost.EXPECT().CheckRequirements()
		mockAnalyticsD.EXPECT().Stop()
		mockAnalyticsD.EXPECT().Destroy()
		mockHypervisor.EXPECT().Stop("cfdev")
		mockHypervisor.EXPECT().Destroy("cfdev")
		mockVpnkit.EXPECT().Stop()
		mockVpnkit.EXPECT().Destroy()

		mockHostNet.EXPECT().RemoveLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip")
		if runtime.GOOS == "darwin" {
			mockCfdevdClient.EXPECT().Uninstall()
		}

		Expect(stopCmd.Execute()).To(Succeed())
	})

	Context("stopping the VM fails", func() {
		It("stops the others and returns VM error", func() {
			mockAnalytics.EXPECT().Event(cfanalytics.STOP)
			mockHost.EXPECT().CheckRequirements()
			mockAnalyticsD.EXPECT().Stop()
			mockAnalyticsD.EXPECT().Destroy()
			mockHypervisor.EXPECT().Stop("cfdev").Return(errors.New("test"))
			mockHypervisor.EXPECT().Destroy("cfdev")
			mockVpnkit.EXPECT().Stop()
			mockVpnkit.EXPECT().Destroy()

			mockHostNet.EXPECT().RemoveLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip")
			if runtime.GOOS == "darwin" {
				mockCfdevdClient.EXPECT().Uninstall()
			}

			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to stop the VM: test"))
		})
	})

	Context("destroying the VM fails", func() {
		It("stops the others and returns VM error", func() {
			mockAnalytics.EXPECT().Event(cfanalytics.STOP)
			mockHost.EXPECT().CheckRequirements()
			mockAnalyticsD.EXPECT().Stop()
			mockAnalyticsD.EXPECT().Destroy()
			mockHypervisor.EXPECT().Stop("cfdev")
			mockHypervisor.EXPECT().Destroy("cfdev").Return(errors.New("test"))
			mockVpnkit.EXPECT().Stop()
			mockVpnkit.EXPECT().Destroy()

			mockHostNet.EXPECT().RemoveLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip")
			if runtime.GOOS == "darwin" {
				mockCfdevdClient.EXPECT().Uninstall()
			}

			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to destroy the VM: test"))
		})
	})

	Context("stopping vpnkit fails", func() {
		It("stops the others and returns vpnkit error", func() {
			mockAnalytics.EXPECT().Event(cfanalytics.STOP)
			mockHost.EXPECT().CheckRequirements()
			mockAnalyticsD.EXPECT().Stop()
			mockAnalyticsD.EXPECT().Destroy()
			mockHypervisor.EXPECT().Stop("cfdev")
			mockHypervisor.EXPECT().Destroy("cfdev")
			mockVpnkit.EXPECT().Stop().Return(errors.New("test"))
			mockVpnkit.EXPECT().Destroy()

			mockHostNet.EXPECT().RemoveLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip")
			if runtime.GOOS == "darwin" {
				mockCfdevdClient.EXPECT().Uninstall()
			}

			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to stop vpnkit: test"))
		})
	})

	Context("destroying vpnkit fails", func() {
		It("stops the others and returns vpnkit error", func() {
			mockAnalytics.EXPECT().Event(cfanalytics.STOP)
			mockHost.EXPECT().CheckRequirements()
			mockAnalyticsD.EXPECT().Stop()
			mockAnalyticsD.EXPECT().Destroy()
			mockHypervisor.EXPECT().Stop("cfdev")
			mockHypervisor.EXPECT().Destroy("cfdev")
			mockVpnkit.EXPECT().Stop()
			mockVpnkit.EXPECT().Destroy().Return(errors.New("test"))

			mockHostNet.EXPECT().RemoveLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip")
			if runtime.GOOS == "darwin" {
				mockCfdevdClient.EXPECT().Uninstall()
			}

			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to destroy vpnkit: test"))
		})
	})

	Context("stopping analyticsd fails", func() {
		It("stops the others and returns analyticsd error", func() {
			mockAnalytics.EXPECT().Event(cfanalytics.STOP)
			mockHost.EXPECT().CheckRequirements()
			mockAnalyticsD.EXPECT().Stop().Return(errors.New("test"))
			mockAnalyticsD.EXPECT().Destroy()
			mockHypervisor.EXPECT().Stop("cfdev")
			mockHypervisor.EXPECT().Destroy("cfdev")
			mockVpnkit.EXPECT().Stop()
			mockVpnkit.EXPECT().Destroy()

			mockHostNet.EXPECT().RemoveLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip")
			if runtime.GOOS == "darwin" {
				mockCfdevdClient.EXPECT().Uninstall()
			}

			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to stop analyticsd: test"))
		})
	})

	Context("destroying analyticsd fails", func() {
		It("stops the others and returns analyticsd error", func() {
			mockAnalytics.EXPECT().Event(cfanalytics.STOP)
			mockHost.EXPECT().CheckRequirements()
			mockAnalyticsD.EXPECT().Stop()
			mockAnalyticsD.EXPECT().Destroy().Return(errors.New("test"))
			mockHypervisor.EXPECT().Stop("cfdev")
			mockHypervisor.EXPECT().Destroy("cfdev")
			mockVpnkit.EXPECT().Stop()
			mockVpnkit.EXPECT().Destroy()

			mockHostNet.EXPECT().RemoveLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip")
			if runtime.GOOS == "darwin" {
				mockCfdevdClient.EXPECT().Uninstall()
			}

			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to destroy analyticsd: test"))
		})
	})

	Context("removing aliases fails", func() {
		It("stops the others and returns alias error", func() {
			mockAnalytics.EXPECT().Event(cfanalytics.STOP)
			mockHost.EXPECT().CheckRequirements()
			mockAnalyticsD.EXPECT().Stop()
			mockAnalyticsD.EXPECT().Destroy()
			mockHypervisor.EXPECT().Stop("cfdev")
			mockHypervisor.EXPECT().Destroy("cfdev")
			mockVpnkit.EXPECT().Stop()
			mockVpnkit.EXPECT().Destroy()

			mockHostNet.EXPECT().RemoveLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip").Return(errors.New("test"))
			if runtime.GOOS == "darwin" {
				mockCfdevdClient.EXPECT().Uninstall()
			}

			Expect(stopCmd.Execute()).To(MatchError(`cf dev stop: failed to remove IP aliases: test`))
		})
	})
})
