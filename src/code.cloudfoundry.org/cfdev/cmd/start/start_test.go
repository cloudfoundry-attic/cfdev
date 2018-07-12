package start_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/cmd/start"
	"code.cloudfoundry.org/cfdev/cmd/start/mocks"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/cfdev/resource"
	"github.com/golang/mock/gomock"
)

var _ = Describe("Start", func() {

	var (
		mockController      *gomock.Controller
		mockUI              *mocks.MockUI
		mockAnalyticsClient *mocks.MockAnalyticsClient
		mockToggle          *mocks.MockToggle
		mockHostNet         *mocks.MockHostNet
		mockCache           *mocks.MockCache
		mockCFDevD          *mocks.MockCFDevD
		mockVpnKit          *mocks.MockVpnKit
		mockLinuxKit        *mocks.MockLinuxKit
		mockGardenClient    *mocks.MockGardenClient

		startCmd      start.Start
		exitChan      chan struct{}
		localExitChan chan string
		tmpDir        string
		cacheDir      string
	)

	BeforeEach(func() {
		var err error
		mockController = gomock.NewController(GinkgoT())
		mockUI = mocks.NewMockUI(mockController)
		mockAnalyticsClient = mocks.NewMockAnalyticsClient(mockController)
		mockToggle = mocks.NewMockToggle(mockController)
		mockHostNet = mocks.NewMockHostNet(mockController)
		mockCache = mocks.NewMockCache(mockController)
		mockCFDevD = mocks.NewMockCFDevD(mockController)
		mockVpnKit = mocks.NewMockVpnKit(mockController)
		mockLinuxKit = mocks.NewMockLinuxKit(mockController)
		mockGardenClient = mocks.NewMockGardenClient(mockController)

		localExitChan = make(chan string, 3)
		tmpDir, err = ioutil.TempDir("", "start-test-home")
		cacheDir = filepath.Join(tmpDir, "some-cache-dir")
		Expect(err).NotTo(HaveOccurred())

		startCmd = start.Start{
			Config: config.Config{
				CFDevHome:      tmpDir,
				StateDir:       filepath.Join(tmpDir, "some-state-dir"),
				VpnKitStateDir: filepath.Join(tmpDir, "some-vpnkit-state-dir"),
				CacheDir:       cacheDir,
				CFRouterIP:     "some-cf-router-ip", BoshDirectorIP: "some-bosh-director-ip", Dependencies: resource.Catalog{
					Items: []resource.Item{
						{Name: "some-item"},
						{Name: "cf-deps.iso"},
					},
				},
			},
			Exit:            exitChan,
			LocalExit:       localExitChan,
			UI:              mockUI,
			Analytics:       mockAnalyticsClient,
			AnalyticsToggle: mockToggle,
			HostNet:         mockHostNet,
			Cache:           mockCache,
			CFDevD:          mockCFDevD,
			VpnKit:          mockVpnKit,
			LinuxKit:        mockLinuxKit,
			GardenClient:    mockGardenClient,
		}
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
		mockController.Finish()
	})

	Describe("Execute", func() {
		Context("when no args are provided", func() {
			It("starts the vm with default settings", func() {
				gomock.InOrder(
					mockToggle.EXPECT().SetProp("type", "cf"),
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN),
					mockLinuxKit.EXPECT().IsRunning().Return(false, nil),
					mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
					mockUI.EXPECT().Say("Downloading Resources..."),
					mockCache.EXPECT().Sync(resource.Catalog{
						Items: []resource.Item{
							{Name: "some-item"},
							{Name: "cf-deps.iso"},
						},
					}),
					mockUI.EXPECT().Say("Installing cfdevd network helper..."),
					mockCFDevD.EXPECT().Install(),
					mockUI.EXPECT().Say("Starting VPNKit..."),
					mockVpnKit.EXPECT().Start(),
					mockVpnKit.EXPECT().Watch(localExitChan),
					mockUI.EXPECT().Say("Starting the VM..."),
					mockLinuxKit.EXPECT().Start(7, 6666, filepath.Join(cacheDir, "cf-deps.iso")),
					mockLinuxKit.EXPECT().Watch(localExitChan),
					mockUI.EXPECT().Say("Waiting for Garden..."),
					mockGardenClient.EXPECT().Ping(),
					mockUI.EXPECT().Say("Deploying the BOSH Director..."),
					mockGardenClient.EXPECT().DeployBosh(),
					mockUI.EXPECT().Say("Deploying CF..."),
					mockGardenClient.EXPECT().ReportProgress(mockUI, "cf"),
					mockGardenClient.EXPECT().DeployCloudFoundry(nil),
					mockGardenClient.EXPECT().GetServices().Return([]garden.Service{
						{
							Name:       "some-service",
							Handle:     "some-handle",
							Script:     "/path/to/some-script",
							Deployment: "some-deployment",
						},
						{
							Name:       "some-other-service",
							Handle:     "some-other-handle",
							Script:     "/path/to/some-other-script",
							Deployment: "some-other-deployment",
						},
					}, "", nil),
					mockUI.EXPECT().Say("Deploying %s...", "some-service"),
					mockGardenClient.EXPECT().ReportProgress(mockUI, "some-deployment"),
					mockGardenClient.EXPECT().DeployService("some-handle", "/path/to/some-script"),
					mockUI.EXPECT().Say("Deploying %s...", "some-other-service"),
					mockGardenClient.EXPECT().ReportProgress(mockUI, "some-other-deployment"),
					mockGardenClient.EXPECT().DeployService("some-other-handle", "/path/to/some-other-script"),

					//welcome message
					mockUI.EXPECT().Say(gomock.Any()),
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END),
				)

				Expect(startCmd.Execute(start.Args{
					Cpus: 7,
					Mem:  6666,
				})).To(Succeed())
			})
		})

		Context("when the --no-provision flag is provided", func() {
			It("starts the VM and garden but does not provision", func() {
				gomock.InOrder(
					mockToggle.EXPECT().SetProp("type", "cf"),
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN),
					mockLinuxKit.EXPECT().IsRunning().Return(false, nil),
					mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
					mockUI.EXPECT().Say("Downloading Resources..."),
					mockCache.EXPECT().Sync(resource.Catalog{
						Items: []resource.Item{
							{Name: "some-item"},
							{Name: "cf-deps.iso"},
						},
					}),
					mockUI.EXPECT().Say("Installing cfdevd network helper..."),
					mockCFDevD.EXPECT().Install(),
					mockUI.EXPECT().Say("Starting VPNKit..."),
					mockVpnKit.EXPECT().Start(),
					mockVpnKit.EXPECT().Watch(localExitChan),
					mockUI.EXPECT().Say("Starting the VM..."),
					mockLinuxKit.EXPECT().Start(7, 6666, filepath.Join(cacheDir, "cf-deps.iso")),
					mockLinuxKit.EXPECT().Watch(localExitChan),
					mockUI.EXPECT().Say("Waiting for Garden..."),
					mockGardenClient.EXPECT().Ping(),
				)

				//no provision message message
				mockUI.EXPECT().Say(gomock.Any())

				Expect(startCmd.Execute(start.Args{
					Cpus:        7,
					Mem:         6666,
					NoProvision: true,
				})).To(Succeed())
			})
		})

		Context("when the -f flag is provided", func() {
			It("starts the given iso, doesn't download cf-deps.iso, adds the iso name as an analytics property", func() {
				gomock.InOrder(
					mockToggle.EXPECT().SetProp("type", "some-deps.iso"),
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN),
					mockLinuxKit.EXPECT().IsRunning().Return(false, nil),
					mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
					mockUI.EXPECT().Say("Downloading Resources..."),
					// don't download cf-deps.iso that we won't use
					mockCache.EXPECT().Sync(resource.Catalog{
						Items: []resource.Item{
							{Name: "some-item"},
						},
					}),
					mockUI.EXPECT().Say("Installing cfdevd network helper..."),
					mockCFDevD.EXPECT().Install(),
					mockUI.EXPECT().Say("Starting VPNKit..."),
					mockVpnKit.EXPECT().Start(),
					mockVpnKit.EXPECT().Watch(localExitChan),
					mockUI.EXPECT().Say("Starting the VM..."),
					mockLinuxKit.EXPECT().Start(7, 6666, "/path/to/some-deps.iso"),
					mockLinuxKit.EXPECT().Watch(localExitChan),
					mockUI.EXPECT().Say("Waiting for Garden..."),
					mockGardenClient.EXPECT().Ping(),
					mockUI.EXPECT().Say("Deploying the BOSH Director..."),
					mockGardenClient.EXPECT().DeployBosh(),
					mockUI.EXPECT().Say("Deploying CF..."),
					mockGardenClient.EXPECT().ReportProgress(mockUI, "cf"),
					mockGardenClient.EXPECT().DeployCloudFoundry(nil),
					mockGardenClient.EXPECT().GetServices().Return([]garden.Service{
						{
							Name:       "some-service",
							Handle:     "some-handle",
							Script:     "/path/to/some-script",
							Deployment: "some-deployment",
						},
						{
							Name:       "some-other-service",
							Handle:     "some-other-handle",
							Script:     "/path/to/some-other-script",
							Deployment: "some-other-deployment",
						},
					}, "", nil),
					mockUI.EXPECT().Say("Deploying %s...", "some-service"),
					mockGardenClient.EXPECT().ReportProgress(mockUI, "some-deployment"),
					mockGardenClient.EXPECT().DeployService("some-handle", "/path/to/some-script"),
					mockUI.EXPECT().Say("Deploying %s...", "some-other-service"),
					mockGardenClient.EXPECT().ReportProgress(mockUI, "some-other-deployment"),
					mockGardenClient.EXPECT().DeployService("some-other-handle", "/path/to/some-other-script"),

					//welcome message
					mockUI.EXPECT().Say(gomock.Any()),
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END),
				)

				Expect(startCmd.Execute(start.Args{
					Cpus:        7,
					Mem:         6666,
					DepsIsoPath: "/path/to/some-deps.iso",
				})).To(Succeed())
			})
		})

		Context("when linuxkit is already running", func() {
			It("says cf dev is already running", func() {
				gomock.InOrder(
					mockToggle.EXPECT().SetProp("type", "cf"),
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN),
					mockLinuxKit.EXPECT().IsRunning().Return(true, nil),
					mockUI.EXPECT().Say("CF Dev is already running..."),
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END, map[string]interface{}{"alreadyrunning": true}),
				)

				Expect(startCmd.Execute(start.Args{})).To(Succeed())
			})
		})
	})
})
