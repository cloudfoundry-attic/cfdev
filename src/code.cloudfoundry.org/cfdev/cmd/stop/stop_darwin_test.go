package stop_test

import (
	"fmt"
	"io/ioutil"
	"os"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/cmd/stop"
	"code.cloudfoundry.org/cfdev/cmd/stop/mocks"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/launchd"
	"code.cloudfoundry.org/cfdev/process"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/spf13/cobra"
)

var _ = Describe("Stop", func() {
	var (
		cfg              config.Config
		stopCmd          *cobra.Command
		mockLaunchd      *mocks.MockLaunchd
		mockProcManager  *mocks.MockProcManager
		mockCfdevdClient *mocks.MockCfdevdClient
		mockAnalytics    *mocks.MockAnalytics
		mockController   *gomock.Controller
		stateDir         string
		err              error
	)

	BeforeEach(func() {
		stateDir, err = ioutil.TempDir(os.Getenv("TMPDIR"), "state-dir")
		Expect(err).NotTo(HaveOccurred())

		cfg = config.Config{
			StateDir: stateDir,
		}

		mockController = gomock.NewController(GinkgoT())
		mockLaunchd = mocks.NewMockLaunchd(mockController)
		mockProcManager = mocks.NewMockProcManager(mockController)
		mockCfdevdClient = mocks.NewMockCfdevdClient(mockController)
		mockAnalytics = mocks.NewMockAnalytics(mockController)

		subject := &stop.Stop{
			Config:       cfg,
			Analytics:    mockAnalytics,
			Launchd:      mockLaunchd,
			ProcManager:  mockProcManager,
			CfdevdClient: mockCfdevdClient,
		}
		stopCmd = subject.Cmd()
		stopCmd.SetArgs([]string{})
		stopCmd.SetOutput(GinkgoWriter)
	})

	AfterEach(func() {
		mockController.Finish()
		os.RemoveAll(stateDir)
	})

	It("uninstalls linuxkit, vpnkit, and cfdevd, and sends analytics event", func() {
		mockAnalytics.EXPECT().Event(cfanalytics.STOP)
		mockLaunchd.EXPECT().RemoveDaemon(launchd.DaemonSpec{
			Label: process.LinuxKitLabel,
		})
		mockLaunchd.EXPECT().RemoveDaemon(launchd.DaemonSpec{
			Label: process.VpnKitLabel,
		})
		mockProcManager.EXPECT().SafeKill(gomock.Any(), "hyperkit")
		mockCfdevdClient.EXPECT().Uninstall()
		Expect(stopCmd.Execute()).To(Succeed())
	})

	Context("stopping linuxkit fails", func() {
		It("stops the others and returns linuxkit error", func() {
			mockAnalytics.EXPECT().Event(cfanalytics.STOP)
			mockLaunchd.EXPECT().RemoveDaemon(launchd.DaemonSpec{
				Label: process.VpnKitLabel,
			})
			mockLaunchd.EXPECT().RemoveDaemon(launchd.DaemonSpec{
				Label: process.LinuxKitLabel,
			}).Return(fmt.Errorf("test"))
			mockProcManager.EXPECT().SafeKill(gomock.Any(), "hyperkit")
			mockCfdevdClient.EXPECT().Uninstall()

			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to stop linuxkit: test"))
		})
	})

	Context("stopping vpnkit fails", func() {
		It("stops the others and returns vpnkit error", func() {
			mockAnalytics.EXPECT().Event(cfanalytics.STOP)
			mockLaunchd.EXPECT().RemoveDaemon(launchd.DaemonSpec{
				Label: process.LinuxKitLabel,
			})
			mockLaunchd.EXPECT().RemoveDaemon(launchd.DaemonSpec{
				Label: process.VpnKitLabel,
			}).Return(fmt.Errorf("test"))
			mockProcManager.EXPECT().SafeKill(gomock.Any(), "hyperkit")
			mockCfdevdClient.EXPECT().Uninstall()
			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to stop vpnkit: test"))
		})
	})

	Context("stopping hyperkit fails", func() {
		It("stops the others and returns vpnkit error", func() {
			mockAnalytics.EXPECT().Event(cfanalytics.STOP)
			mockLaunchd.EXPECT().RemoveDaemon(launchd.DaemonSpec{
				Label: process.LinuxKitLabel,
			})
			mockLaunchd.EXPECT().RemoveDaemon(launchd.DaemonSpec{
				Label: process.VpnKitLabel,
			})
			mockProcManager.EXPECT().SafeKill(gomock.Any(), "hyperkit").Return(fmt.Errorf("test"))
			mockCfdevdClient.EXPECT().Uninstall()
			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to kill hyperkit: test"))
		})
	})

	Context("stopping cfdevd fails", func() {
		It("stops the others and returns cfdevd error", func() {
			mockAnalytics.EXPECT().Event(cfanalytics.STOP)
			mockLaunchd.EXPECT().RemoveDaemon(launchd.DaemonSpec{
				Label: process.LinuxKitLabel,
			})
			mockLaunchd.EXPECT().RemoveDaemon(launchd.DaemonSpec{
				Label: process.VpnKitLabel,
			})
			mockProcManager.EXPECT().SafeKill(gomock.Any(), "hyperkit")
			mockCfdevdClient.EXPECT().Uninstall().Return("test", fmt.Errorf("test"))

			Expect(stopCmd.Execute()).To(MatchError("cf dev stop: failed to uninstall cfdevd: test"))
		})
	})
})
