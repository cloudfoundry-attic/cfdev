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
	"code.cloudfoundry.org/cfdev/iso"
	"code.cloudfoundry.org/cfdev/process"
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
		mockGardenClient    *mocks.MockGardenClient
		mockHyperV          *mocks.MockHyperV
		mockIsoReader       *mocks.MockIsoReader

		startCmd      start.Start
		exitChan      chan struct{}
		localExitChan chan string
		tmpDir        string
		cacheDir      string
		depsIsoPath   string
		metadata      iso.Metadata
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
		mockHyperV = mocks.NewMockHyperV(mockController)
		mockGardenClient = mocks.NewMockGardenClient(mockController)
		mockIsoReader = mocks.NewMockIsoReader(mockController)

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
			HyperV:          mockHyperV,
			GardenClient:    mockGardenClient,
			IsoReader:       mockIsoReader,
		}

		depsIsoPath = filepath.Join(cacheDir, "cf-deps.iso")
		metadata = iso.Metadata{
			Version:       "v1",
			DefaultMemory: 8765,
			Services: []garden.Service{
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
			},
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
					mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
					mockUI.EXPECT().Say("Downloading Resources..."),
					mockCache.EXPECT().Sync(resource.Catalog{
						Items: []resource.Item{
							{Name: "some-item"},
							{Name: "cf-deps.iso"},
						},
					}),
					mockIsoReader.EXPECT().Read(depsIsoPath).Return(metadata, nil),
					mockUI.EXPECT().Say("Creating the VM..."),
					mockHyperV.EXPECT().CreateVM(process.VM{
						DepsIso:  filepath.Join(cacheDir, "cf-deps.iso"),
						MemoryMB: 6666,
						CPUs:     7,
					}),
					mockUI.EXPECT().Say("Starting VPNKit..."),
					mockVpnKit.EXPECT().Start(),
					mockUI.EXPECT().Say("Starting the VM..."),
					mockHyperV.EXPECT().Start("cfdev"),
					mockUI.EXPECT().Say("Waiting for Garden..."),
					mockGardenClient.EXPECT().Ping(),
					mockUI.EXPECT().Say("Deploying the BOSH Director..."),
					mockGardenClient.EXPECT().DeployBosh(),
					mockUI.EXPECT().Say("Deploying CF..."),
					mockGardenClient.EXPECT().ReportProgress(mockUI, "cf"),
					mockGardenClient.EXPECT().DeployCloudFoundry(nil),
					mockGardenClient.EXPECT().DeployServices(mockUI, []garden.Service{
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
					}),

					//welcome message
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END),
				)

				Expect(startCmd.Execute(start.Args{
					Cpus: 7,
					Mem:  6666,
				})).To(Succeed())
			})
		})

		Context("-f flag provided", func() {
			It("starts the vm with provided iso", func() {
				customIso := filepath.Join(tmpDir, "custom.iso")
				ioutil.WriteFile(customIso, []byte{}, 0644)

				gomock.InOrder(
					mockToggle.EXPECT().SetProp("type", "custom.iso"),
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN),
					mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
					mockUI.EXPECT().Say("Downloading Resources..."),
					mockCache.EXPECT().Sync(resource.Catalog{
						Items: []resource.Item{
							{Name: "some-item"},
						},
					}),
					mockIsoReader.EXPECT().Read(customIso).Return(metadata, nil),
					mockUI.EXPECT().Say("Creating the VM..."),
					mockHyperV.EXPECT().CreateVM(process.VM{
						DepsIso:  customIso,
						MemoryMB: 6666,
						CPUs:     7,
					}),
					mockUI.EXPECT().Say("Starting VPNKit..."),
					mockVpnKit.EXPECT().Start(),
					mockUI.EXPECT().Say("Starting the VM..."),
					mockHyperV.EXPECT().Start("cfdev"),
					mockUI.EXPECT().Say("Waiting for Garden..."),
					mockGardenClient.EXPECT().Ping(),
					mockUI.EXPECT().Say("Deploying the BOSH Director..."),
					mockGardenClient.EXPECT().DeployBosh(),
					mockUI.EXPECT().Say("Deploying CF..."),
					mockGardenClient.EXPECT().ReportProgress(mockUI, "cf"),
					mockGardenClient.EXPECT().DeployCloudFoundry(nil),
					mockGardenClient.EXPECT().DeployServices(mockUI, []garden.Service{
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
					}),

					//welcome message
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END),
				)

				Expect(startCmd.Execute(start.Args{
					DepsIsoPath: customIso,
					Cpus:        7,
					Mem:         6666,
				})).To(Succeed())
			})
		})

		Context("when the --no-provision flag is provided", func() {
			It("starts the VM and garden but does not provision", func() {
				gomock.InOrder(
					mockToggle.EXPECT().SetProp("type", "cf"),
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN),
					mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
					mockUI.EXPECT().Say("Downloading Resources..."),
					mockCache.EXPECT().Sync(resource.Catalog{
						Items: []resource.Item{
							{Name: "some-item"},
							{Name: "cf-deps.iso"},
						},
					}),
					mockIsoReader.EXPECT().Read(depsIsoPath).Return(metadata, nil),
					mockUI.EXPECT().Say("Creating the VM..."),
					mockHyperV.EXPECT().CreateVM(process.VM{
						DepsIso:  filepath.Join(cacheDir, "cf-deps.iso"),
						MemoryMB: 4096,
						CPUs:     4,
					}),
					mockUI.EXPECT().Say("Starting VPNKit..."),
					mockVpnKit.EXPECT().Start(),
					mockUI.EXPECT().Say("Starting the VM..."),
					mockHyperV.EXPECT().Start("cfdev"),
					mockUI.EXPECT().Say("Waiting for Garden..."),
					mockGardenClient.EXPECT().Ping(),
				)

				//no provision message message
				mockUI.EXPECT().Say(gomock.Any())

				Expect(startCmd.Execute(start.Args{
					Cpus:        4,
					Mem:         4096,
					NoProvision: true,
				})).To(Succeed())
			})
		})

		Context("when the -f flag is provided with a non-existing filepath", func() {
			It("returns an error message and does not execute start command", func() {

				wrongPath := "C:\\wrong-path-to-some-deps.iso"

				Expect(startCmd.Execute(start.Args{
					Cpus:        7,
					Mem:         6666,
					DepsIsoPath: wrongPath,
				})).To(MatchError("no file found at: " + wrongPath))
			})
		})
	})
})
