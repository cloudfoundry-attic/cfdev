package start_test

import (
	"runtime"

	mdata "code.cloudfoundry.org/cfdev/metadata"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/cfanalytics"
	"code.cloudfoundry.org/cfdev/cmd/start"
	"code.cloudfoundry.org/cfdev/cmd/start/mocks"
	"code.cloudfoundry.org/cfdev/config"
	"code.cloudfoundry.org/cfdev/hypervisor"
	"code.cloudfoundry.org/cfdev/provision"
	"code.cloudfoundry.org/cfdev/resource"
	"github.com/golang/mock/gomock"
)

var _ = XDescribe("Start", func() {

	var (
		mockController      *gomock.Controller
		mockUI              *mocks.MockUI
		mockAnalyticsClient *mocks.MockAnalyticsClient
		mockToggle          *mocks.MockToggle
		mockHostNet         *mocks.MockHostNet
		mockHost            *mocks.MockHost
		mockCache           *mocks.MockCache
		mockCFDevD          *mocks.MockCFDevD
		mockVpnKit          *mocks.MockVpnKit
		mockAnalyticsD      *mocks.MockAnalyticsD
		mockHypervisor      *mocks.MockHypervisor
		mockProvisioner     *mocks.MockProvisioner
		mockProvision       *mocks.MockProvision
		mockSystemProfiler  *mocks.MockSystemProfiler
		mockMetadataReader  *mocks.MockMetaDataReader
		mockEnv             *mocks.MockEnv
		mockStop            *mocks.MockStop

		startCmd      start.Start
		exitChan      chan struct{}
		localExitChan chan string
		tmpDir        string
		cacheDir      string
		stateDir      string
		metadata      mdata.Metadata
	)

	services := []provision.Service{
		{
			Name:          "some-service",
			Flagname:      "some-service-flagname",
			Script:        "/path/to/some-script",
			Deployment:    "some-deployment",
		},
		{
			Name:          "some-other-service",
			Flagname:      "some-other-service-flagname",
			Script:        "/path/to/some-other-script",
			Deployment:    "some-other-deployment",
		},
	}
	BeforeEach(func() {
		var err error
		mockController = gomock.NewController(GinkgoT())
		mockUI = mocks.NewMockUI(mockController)
		mockAnalyticsClient = mocks.NewMockAnalyticsClient(mockController)
		mockToggle = mocks.NewMockToggle(mockController)
		mockHostNet = mocks.NewMockHostNet(mockController)
		mockHost = mocks.NewMockHost(mockController)
		mockCache = mocks.NewMockCache(mockController)
		mockCFDevD = mocks.NewMockCFDevD(mockController)
		mockVpnKit = mocks.NewMockVpnKit(mockController)
		mockAnalyticsD = mocks.NewMockAnalyticsD(mockController)
		mockHypervisor = mocks.NewMockHypervisor(mockController)
		mockProvisioner = mocks.NewMockProvisioner(mockController)
		mockProvision = mocks.NewMockProvision(mockController)
		mockSystemProfiler = mocks.NewMockSystemProfiler(mockController)
		mockMetadataReader = mocks.NewMockMetaDataReader(mockController)
		mockEnv = mocks.NewMockEnv(mockController)
		mockStop = mocks.NewMockStop(mockController)

		localExitChan = make(chan string, 3)
		tmpDir, err = ioutil.TempDir("", "start-test-home")
		cacheDir = filepath.Join(tmpDir, "some-cache-dir")
		stateDir = filepath.Join(tmpDir, "some-state-dir")
		Expect(err).NotTo(HaveOccurred())

		startCmd = start.Start{
			Config: config.Config{
				CFDevHome:      tmpDir,
				StateDir:       stateDir,
				StateBosh:      filepath.Join(tmpDir, "some-bosh-state-dir"),
				StateLinuxkit:  filepath.Join(tmpDir, "some-linuxkit-state-dir"),
				VpnKitStateDir: filepath.Join(tmpDir, "some-vpnkit-state-dir"),
				CacheDir:       cacheDir,
				CFRouterIP:     "some-cf-router-ip",
				BoshDirectorIP: "some-bosh-director-ip",
				Dependencies: resource.Catalog{
					Items: []resource.Item{
						{Name: "some-item"},
						{Name: "cfdev-deps.tgz"},
					},
				},
			},
			Exit:            exitChan,
			LocalExit:       localExitChan,
			UI:              mockUI,
			Analytics:       mockAnalyticsClient,
			AnalyticsToggle: mockToggle,
			HostNet:         mockHostNet,
			Host:            mockHost,
			Cache:           mockCache,
			CFDevD:          mockCFDevD,
			VpnKit:          mockVpnKit,
			AnalyticsD:      mockAnalyticsD,
			Hypervisor:      mockHypervisor,
			Provisioner:     mockProvisioner,
			Provision:       mockProvision,
			MetaDataReader:  mockMetadataReader,
			Env:             mockEnv,
			Stop:            mockStop,
			Profiler:        mockSystemProfiler,
		}

		metadata = mdata.Metadata{
			Version:          "v4",
			DefaultMemory:    8765,
			DeploymentName:   "cf",
			ArtifactVersion:  "some-artifact-version",
			AnalyticsMessage: "",
			Services:         services,
		}
	})

	AfterEach(func() {
		os.RemoveAll(tmpDir)
		mockController.Finish()
	})

	Describe("Execute", func() {
		Context("when no args are provided", func() {
			It("starts the vm with default settings", func() {
				if runtime.GOOS == "darwin" {
					mockUI.EXPECT().Say("Installing cfdevd network helper...")
					mockCFDevD.EXPECT().Install()
				}

				gomock.InOrder(
					mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(111), nil),
					mockSystemProfiler.EXPECT().GetTotalMemory().Return(uint64(222), nil),

					mockHost.EXPECT().CheckRequirements(),
					mockHypervisor.EXPECT().IsRunning("cfdev").Return(false, nil),
					mockStop.EXPECT().RunE(nil, nil),
					mockEnv.EXPECT().CreateDirs(),

					mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
					mockUI.EXPECT().Say("Downloading Resources..."),
					mockCache.EXPECT().Sync(resource.Catalog{
						Items: []resource.Item{
							{Name: "some-item"},
							{Name: "cfdev-deps.tgz"},
						},
					}),
					mockUI.EXPECT().Say("Setting State..."),
					mockEnv.EXPECT().SetupState(filepath.Join(cacheDir, "cfdev-deps.tgz")),
					mockMetadataReader.EXPECT().Read(filepath.Join(stateDir, "metadata.yml")).Return(metadata, nil),

					mockToggle.EXPECT().SetProp("type", "cf"),
					mockToggle.EXPECT().SetProp("artifact", "some-artifact-version"),
					mockAnalyticsClient.EXPECT().PromptOptInIfNeeded(""),
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN, map[string]interface{}{
						"total memory":     uint64(222),
						"available memory": uint64(111),
					}),
					mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(10000), nil),
					mockUI.EXPECT().Say("Creating the VM..."),
					mockHypervisor.EXPECT().CreateVM(hypervisor.VM{
						Name:     "cfdev",
						CPUs:     7,
						MemoryMB: 8765,
					}),
					mockUI.EXPECT().Say("Starting VPNKit..."),
					mockVpnKit.EXPECT().Start(),
					mockVpnKit.EXPECT().Watch(localExitChan),
					mockUI.EXPECT().Say("Starting the VM..."),
					mockHypervisor.EXPECT().Start("cfdev"),
					mockUI.EXPECT().Say("Waiting for the VM..."),
					mockProvisioner.EXPECT().Ping(),
					mockProvision.EXPECT().Execute(start.Args{Cpus: 7, Mem: 0}),

					mockToggle.EXPECT().Enabled().Return(true),
					mockAnalyticsD.EXPECT().Start(),
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END),
				)

				Expect(startCmd.Execute(start.Args{
					Cpus: 7,
					Mem:  0,
				})).To(Succeed())
			})

			It("starts the vm with analytics toggled off", func() {
				if runtime.GOOS == "darwin" {
					mockUI.EXPECT().Say("Installing cfdevd network helper...")
					mockCFDevD.EXPECT().Install()
				}

				gomock.InOrder(
					mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(111), nil),
					mockSystemProfiler.EXPECT().GetTotalMemory().Return(uint64(222), nil),
					mockHost.EXPECT().CheckRequirements(),
					mockHypervisor.EXPECT().IsRunning("cfdev").Return(false, nil),
					mockStop.EXPECT().RunE(nil, nil),
					mockEnv.EXPECT().CreateDirs(),

					mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
					mockUI.EXPECT().Say("Downloading Resources..."),
					mockCache.EXPECT().Sync(resource.Catalog{
						Items: []resource.Item{
							{Name: "some-item"},
							{Name: "cfdev-deps.tgz"},
						},
					}),
					mockUI.EXPECT().Say("Setting State..."),
					mockEnv.EXPECT().SetupState(filepath.Join(cacheDir, "cfdev-deps.tgz")),
					mockMetadataReader.EXPECT().Read(filepath.Join(stateDir, "metadata.yml")).Return(metadata, nil),

					mockToggle.EXPECT().SetProp("type", "cf"),
					mockToggle.EXPECT().SetProp("artifact", "some-artifact-version"),
					mockAnalyticsClient.EXPECT().PromptOptInIfNeeded(""),
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN, map[string]interface{}{
						"total memory":     uint64(222),
						"available memory": uint64(111),
					}),
					mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(10000), nil),
					mockUI.EXPECT().Say("Creating the VM..."),
					mockHypervisor.EXPECT().CreateVM(hypervisor.VM{
						Name:     "cfdev",
						CPUs:     7,
						MemoryMB: 8765,
					}),
					mockUI.EXPECT().Say("Starting VPNKit..."),
					mockVpnKit.EXPECT().Start(),
					mockVpnKit.EXPECT().Watch(localExitChan),
					mockUI.EXPECT().Say("Starting the VM..."),
					mockHypervisor.EXPECT().Start("cfdev"),
					mockUI.EXPECT().Say("Waiting for the VM..."),
					mockProvisioner.EXPECT().Ping(),
					mockProvision.EXPECT().Execute(start.Args{Cpus: 7, Mem: 0}),

					mockToggle.EXPECT().Enabled().Return(false),
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END),
				)

				Expect(startCmd.Execute(start.Args{
					Cpus: 7,
					Mem:  0,
				})).To(Succeed())
			})

			Context("when catalog includes cfdevd", func() {
				BeforeEach(func() {
					startCmd.Config.Dependencies = resource.Catalog{
						Items: []resource.Item{
							{Name: "some-item"},
							{Name: "cfdevd"},
							{Name: "cfdev-deps.tgz"},
						},
					}
				})
				It("downloads cfdevd first", func() {
					if runtime.GOOS == "darwin" {
						mockUI.EXPECT().Say("Installing cfdevd network helper...")
						mockCFDevD.EXPECT().Install()
					}

					gomock.InOrder(
						mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(111), nil),
						mockSystemProfiler.EXPECT().GetTotalMemory().Return(uint64(222), nil),
						mockHost.EXPECT().CheckRequirements(),
						mockHypervisor.EXPECT().IsRunning("cfdev").Return(false, nil),
						mockStop.EXPECT().RunE(nil, nil),
						mockEnv.EXPECT().CreateDirs(),

						mockUI.EXPECT().Say("Downloading Network Helper..."),
						mockCache.EXPECT().Sync(resource.Catalog{
							Items: []resource.Item{
								{Name: "cfdevd"},
							},
						}),
						mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
						mockUI.EXPECT().Say("Downloading Resources..."),
						mockCache.EXPECT().Sync(resource.Catalog{
							Items: []resource.Item{
								{Name: "some-item"},
								{Name: "cfdev-deps.tgz"},
							},
						}),
						mockUI.EXPECT().Say("Setting State..."),
						mockEnv.EXPECT().SetupState(filepath.Join(cacheDir, "cfdev-deps.tgz")),
						mockMetadataReader.EXPECT().Read(filepath.Join(stateDir, "metadata.yml")).Return(metadata, nil),

						mockToggle.EXPECT().SetProp("type", "cf"),
						mockToggle.EXPECT().SetProp("artifact", "some-artifact-version"),
						mockAnalyticsClient.EXPECT().PromptOptInIfNeeded(""),
						mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN, map[string]interface{}{
							"total memory":     uint64(222),
							"available memory": uint64(111),
						}),
						mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(110000), nil),
						mockUI.EXPECT().Say("Creating the VM..."),
						mockHypervisor.EXPECT().CreateVM(hypervisor.VM{
							Name:     "cfdev",
							CPUs:     7,
							MemoryMB: 8765,
						}),
						mockUI.EXPECT().Say("Starting VPNKit..."),
						mockVpnKit.EXPECT().Start(),
						mockVpnKit.EXPECT().Watch(localExitChan),
						mockUI.EXPECT().Say("Starting the VM..."),
						mockHypervisor.EXPECT().Start("cfdev"),
						mockUI.EXPECT().Say("Waiting for the VM..."),
						mockProvisioner.EXPECT().Ping(),
						mockProvision.EXPECT().Execute(start.Args{Cpus: 7, Mem: 0}),

						mockToggle.EXPECT().Enabled().Return(true),
						mockAnalyticsD.EXPECT().Start(),
						mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END),
					)

					Expect(startCmd.Execute(start.Args{
						Cpus: 7,
						Mem:  0,
					})).To(Succeed())
				})
			})

			Context("when no args are provided AND deps.iso does not have a default memory field", func() {
				It("starts the vm with a default memory setting", func() {
					if runtime.GOOS == "darwin" {
						mockUI.EXPECT().Say("Installing cfdevd network helper...")
						mockCFDevD.EXPECT().Install()
					}

					gomock.InOrder(
						mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(111), nil),
						mockSystemProfiler.EXPECT().GetTotalMemory().Return(uint64(222), nil),
						mockHost.EXPECT().CheckRequirements(),
						mockHypervisor.EXPECT().IsRunning("cfdev").Return(false, nil),
						mockStop.EXPECT().RunE(nil, nil),
						mockEnv.EXPECT().CreateDirs(),

						mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
						mockUI.EXPECT().Say("Downloading Resources..."),
						mockCache.EXPECT().Sync(resource.Catalog{
							Items: []resource.Item{
								{Name: "some-item"},
								{Name: "cfdev-deps.tgz"},
							},
						}),
						mockUI.EXPECT().Say("Setting State..."),
						mockEnv.EXPECT().SetupState(filepath.Join(cacheDir, "cfdev-deps.tgz")),
						mockMetadataReader.EXPECT().Read(filepath.Join(stateDir, "metadata.yml")).Return(mdata.Metadata{
							Version:          "v4",
							DeploymentName:   "cf",
							ArtifactVersion:  "some-artifact-version",
							AnalyticsMessage: "",
							Services:         services,
						}, nil),

						mockToggle.EXPECT().SetProp("type", "cf"),
						mockToggle.EXPECT().SetProp("artifact", "some-artifact-version"),
						mockAnalyticsClient.EXPECT().PromptOptInIfNeeded(""),
						mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN, map[string]interface{}{
							"total memory":     uint64(222),
							"available memory": uint64(111),
						}),
						mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(10000), nil),
						mockUI.EXPECT().Say("Creating the VM..."),
						mockHypervisor.EXPECT().CreateVM(hypervisor.VM{
							Name:     "cfdev",
							CPUs:     7,
							MemoryMB: 4192,
						}),
						mockUI.EXPECT().Say("Starting VPNKit..."),
						mockVpnKit.EXPECT().Start(),
						mockVpnKit.EXPECT().Watch(localExitChan),
						mockUI.EXPECT().Say("Starting the VM..."),
						mockHypervisor.EXPECT().Start("cfdev"),
						mockUI.EXPECT().Say("Waiting for the VM..."),
						mockProvisioner.EXPECT().Ping(),
						mockProvision.EXPECT().Execute(start.Args{Cpus: 7, Mem: 0}),

						mockToggle.EXPECT().Enabled().Return(true),
						mockAnalyticsD.EXPECT().Start(),
						mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END),
					)

					Expect(startCmd.Execute(start.Args{
						Cpus: 7,
						Mem:  0,
					})).To(Succeed())
				})
			})

			Context("when the system does not have enough memory", func() {
				It("gives a warning but starts the vm anyways", func() {
					if runtime.GOOS == "darwin" {
						mockUI.EXPECT().Say("Installing cfdevd network helper...")
						mockCFDevD.EXPECT().Install()
					}

					gomock.InOrder(
						mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(111), nil),
						mockSystemProfiler.EXPECT().GetTotalMemory().Return(uint64(222), nil),
						mockHost.EXPECT().CheckRequirements(),
						mockHypervisor.EXPECT().IsRunning("cfdev").Return(false, nil),
						mockStop.EXPECT().RunE(nil, nil),
						mockEnv.EXPECT().CreateDirs(),

						mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
						mockUI.EXPECT().Say("Downloading Resources..."),
						mockCache.EXPECT().Sync(resource.Catalog{
							Items: []resource.Item{
								{Name: "some-item"},
								{Name: "cfdev-deps.tgz"},
							},
						}),
						mockUI.EXPECT().Say("Setting State..."),
						mockEnv.EXPECT().SetupState(filepath.Join(cacheDir, "cfdev-deps.tgz")),
						mockMetadataReader.EXPECT().Read(filepath.Join(stateDir, "metadata.yml")).Return(metadata, nil),

						mockToggle.EXPECT().SetProp("type", "cf"),
						mockToggle.EXPECT().SetProp("artifact", "some-artifact-version"),
						mockAnalyticsClient.EXPECT().PromptOptInIfNeeded(""),
						mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN, map[string]interface{}{
							"total memory":     uint64(222),
							"available memory": uint64(111),
						}),
						mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(1000), nil),
						mockUI.EXPECT().Say("WARNING: CF Dev requires 8765 MB of RAM to run. This machine may not have enough free RAM."),
						mockUI.EXPECT().Say("Creating the VM..."),
						mockHypervisor.EXPECT().CreateVM(hypervisor.VM{
							Name:     "cfdev",
							CPUs:     7,
							MemoryMB: 8765,
						}),
						mockUI.EXPECT().Say("Starting VPNKit..."),
						mockVpnKit.EXPECT().Start(),
						mockVpnKit.EXPECT().Watch(localExitChan),
						mockUI.EXPECT().Say("Starting the VM..."),
						mockHypervisor.EXPECT().Start("cfdev"),
						mockUI.EXPECT().Say("Waiting for the VM..."),
						mockProvisioner.EXPECT().Ping(),
						mockProvision.EXPECT().Execute(start.Args{Cpus: 7, Mem: 0}),

						mockToggle.EXPECT().Enabled().Return(true),
						mockAnalyticsD.EXPECT().Start(),
						mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END),
					)

					Expect(startCmd.Execute(start.Args{
						Cpus: 7,
						Mem:  0,
					})).To(Succeed())
				})
			})
		})

		Context("when flags are provided", func() {

			Context("and the --no-provision flag is provided", func() {
				It("starts the VM and garden but does not provision", func() {
					if runtime.GOOS == "darwin" {
						mockUI.EXPECT().Say("Installing cfdevd network helper...")
						mockCFDevD.EXPECT().Install()
					}

					gomock.InOrder(
						mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(111), nil),
						mockSystemProfiler.EXPECT().GetTotalMemory().Return(uint64(222), nil),
						mockHost.EXPECT().CheckRequirements(),
						mockHypervisor.EXPECT().IsRunning("cfdev").Return(false, nil),
						mockStop.EXPECT().RunE(nil, nil),
						mockEnv.EXPECT().CreateDirs(),

						mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
						mockUI.EXPECT().Say("Downloading Resources..."),
						mockCache.EXPECT().Sync(resource.Catalog{
							Items: []resource.Item{
								{Name: "some-item"},
								{Name: "cfdev-deps.tgz"},
							},
						}),
						mockUI.EXPECT().Say("Setting State..."),
						mockEnv.EXPECT().SetupState(filepath.Join(cacheDir, "cfdev-deps.tgz")),
						mockMetadataReader.EXPECT().Read(filepath.Join(stateDir, "metadata.yml")).Return(metadata, nil),

						mockToggle.EXPECT().SetProp("type", "cf"),
						mockToggle.EXPECT().SetProp("artifact", "some-artifact-version"),
						mockAnalyticsClient.EXPECT().PromptOptInIfNeeded(""),
						mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN, map[string]interface{}{
							"total memory":     uint64(222),
							"available memory": uint64(111),
						}),
						mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(10000), nil),
						mockUI.EXPECT().Say("WARNING: It is recommended that you run CF Dev with at least 8765 MB of RAM."),
						mockUI.EXPECT().Say("Creating the VM..."),
						mockHypervisor.EXPECT().CreateVM(hypervisor.VM{
							Name:     "cfdev",
							CPUs:     7,
							MemoryMB: 6666,
						}),
						mockUI.EXPECT().Say("Starting VPNKit..."),
						mockVpnKit.EXPECT().Start(),
						mockVpnKit.EXPECT().Watch(localExitChan),
						mockUI.EXPECT().Say("Starting the VM..."),
						mockHypervisor.EXPECT().Start("cfdev"),
						mockUI.EXPECT().Say("Waiting for the VM..."),
						mockProvisioner.EXPECT().Ping(),
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

			Context("and the requested memory > base memory", func() {
				Context("and available memory > requested memory", func() {
					Context("should start successfully", func() {
						It("starts the vm with default settings", func() {
							if runtime.GOOS == "darwin" {
								mockUI.EXPECT().Say("Installing cfdevd network helper...")
								mockCFDevD.EXPECT().Install()
							}

							gomock.InOrder(
								mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(15000), nil),
								mockSystemProfiler.EXPECT().GetTotalMemory().Return(uint64(16000), nil),
								mockHost.EXPECT().CheckRequirements(),
								mockHypervisor.EXPECT().IsRunning("cfdev").Return(false, nil),
								mockStop.EXPECT().RunE(nil, nil),
								mockEnv.EXPECT().CreateDirs(),

								mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
								mockUI.EXPECT().Say("Downloading Resources..."),
								mockCache.EXPECT().Sync(resource.Catalog{
									Items: []resource.Item{
										{Name: "some-item"},
										{Name: "cfdev-deps.tgz"},
									},
								}),
								mockUI.EXPECT().Say("Setting State..."),
								mockEnv.EXPECT().SetupState(filepath.Join(cacheDir, "cfdev-deps.tgz")),
								mockMetadataReader.EXPECT().Read(filepath.Join(stateDir, "metadata.yml")).Return(metadata, nil),

								mockToggle.EXPECT().SetProp("type", "cf"),
								mockToggle.EXPECT().SetProp("artifact", "some-artifact-version"),
								mockAnalyticsClient.EXPECT().PromptOptInIfNeeded(""),
								mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN, map[string]interface{}{
									"total memory":     uint64(16000),
									"available memory": uint64(15000),
								}),
								mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(15000), nil),
								mockUI.EXPECT().Say("Creating the VM..."),
								mockHypervisor.EXPECT().CreateVM(hypervisor.VM{
									Name:     "cfdev",
									CPUs:     7,
									MemoryMB: 10000,
								}),
								mockUI.EXPECT().Say("Starting VPNKit..."),
								mockVpnKit.EXPECT().Start(),
								mockVpnKit.EXPECT().Watch(localExitChan),
								mockUI.EXPECT().Say("Starting the VM..."),
								mockHypervisor.EXPECT().Start("cfdev"),
								mockUI.EXPECT().Say("Waiting for the VM..."),
								mockProvisioner.EXPECT().Ping(),
								mockProvision.EXPECT().Execute(start.Args{Cpus: 7, Mem: 10000}),

								mockToggle.EXPECT().Enabled().Return(true),
								mockAnalyticsD.EXPECT().Start(),
								mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END),
							)

							Expect(startCmd.Execute(start.Args{
								Cpus: 7,
								Mem:  10000,
							})).To(Succeed())
						})
					})
				})

				Context("and available mem < requested mem", func() {
					It("gives a warning and continues to start up", func() {
						if runtime.GOOS == "darwin" {
							mockUI.EXPECT().Say("Installing cfdevd network helper...")
							mockCFDevD.EXPECT().Install()
						}

						gomock.InOrder(
							mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(9000), nil),
							mockSystemProfiler.EXPECT().GetTotalMemory().Return(uint64(9500), nil),
							mockHost.EXPECT().CheckRequirements(),
							mockHypervisor.EXPECT().IsRunning("cfdev").Return(false, nil),
							mockStop.EXPECT().RunE(nil, nil),
							mockEnv.EXPECT().CreateDirs(),

							mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
							mockUI.EXPECT().Say("Downloading Resources..."),
							mockCache.EXPECT().Sync(resource.Catalog{
								Items: []resource.Item{
									{Name: "some-item"},
									{Name: "cfdev-deps.tgz"},
								},
							}),
							mockUI.EXPECT().Say("Setting State..."),
							mockEnv.EXPECT().SetupState(filepath.Join(cacheDir, "cfdev-deps.tgz")),
							mockMetadataReader.EXPECT().Read(filepath.Join(stateDir, "metadata.yml")).Return(metadata, nil),

							mockToggle.EXPECT().SetProp("type", "cf"),
							mockToggle.EXPECT().SetProp("artifact", "some-artifact-version"),
							mockAnalyticsClient.EXPECT().PromptOptInIfNeeded(""),
							mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN, map[string]interface{}{
								"total memory":     uint64(9500),
								"available memory": uint64(9000),
							}),
							mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(1200), nil),
							mockUI.EXPECT().Say("WARNING: This machine may not have enough available RAM to run with what is specified."),
							mockUI.EXPECT().Say("Creating the VM..."),
							mockHypervisor.EXPECT().CreateVM(hypervisor.VM{
								Name:     "cfdev",
								CPUs:     7,
								MemoryMB: 10000,
							}),
							mockUI.EXPECT().Say("Starting VPNKit..."),
							mockVpnKit.EXPECT().Start(),
							mockVpnKit.EXPECT().Watch(localExitChan),
							mockUI.EXPECT().Say("Starting the VM..."),
							mockHypervisor.EXPECT().Start("cfdev"),
							mockUI.EXPECT().Say("Waiting for the VM..."),
							mockProvisioner.EXPECT().Ping(),
							mockProvision.EXPECT().Execute(start.Args{Cpus: 7, Mem: 10000}),

							mockToggle.EXPECT().Enabled().Return(true),
							mockAnalyticsD.EXPECT().Start(),
							mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END),
						)

						Expect(startCmd.Execute(start.Args{
							Cpus: 7,
							Mem:  10000,
						})).To(Succeed())
					})
				})

			})

			Context("and requested memory < base memory", func() {
				Context("available memory >= requested memory", func() {
					It("starts with warning", func() {
						if runtime.GOOS == "darwin" {
							mockUI.EXPECT().Say("Installing cfdevd network helper...")
							mockCFDevD.EXPECT().Install()
						}

						gomock.InOrder(
							mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(15000), nil),
							mockSystemProfiler.EXPECT().GetTotalMemory().Return(uint64(16000), nil),
							mockHost.EXPECT().CheckRequirements(),
							mockHypervisor.EXPECT().IsRunning("cfdev").Return(false, nil),
							mockStop.EXPECT().RunE(nil, nil),
							mockEnv.EXPECT().CreateDirs(),

							mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
							mockUI.EXPECT().Say("Downloading Resources..."),
							mockCache.EXPECT().Sync(resource.Catalog{
								Items: []resource.Item{
									{Name: "some-item"},
									{Name: "cfdev-deps.tgz"},
								},
							}),
							mockUI.EXPECT().Say("Setting State..."),
							mockEnv.EXPECT().SetupState(filepath.Join(cacheDir, "cfdev-deps.tgz")),
							mockMetadataReader.EXPECT().Read(filepath.Join(stateDir, "metadata.yml")).Return(mdata.Metadata{
								Version:          "v4",
								DefaultMemory:    8765,
								DeploymentName:   "some-deployment-name",
								ArtifactVersion:  "some-other-artifact-version",
								AnalyticsMessage: "some-custom-analytics-message",
								Services:         services,
							}, nil),

							mockToggle.EXPECT().SetProp("type", "some-deployment-name"),
							mockToggle.EXPECT().SetProp("artifact", "some-other-artifact-version"),
							mockAnalyticsClient.EXPECT().PromptOptInIfNeeded("some-custom-analytics-message"),
							mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN, map[string]interface{}{
								"total memory":     uint64(16000),
								"available memory": uint64(15000),
							}),
							mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(15000), nil),
							mockUI.EXPECT().Say("WARNING: It is recommended that you run SOME-DEPLOYMENT-NAME Dev with at least 8765 MB of RAM."),
							mockUI.EXPECT().Say("Creating the VM..."),
							mockHypervisor.EXPECT().CreateVM(hypervisor.VM{
								Name:     "cfdev",
								CPUs:     7,
								MemoryMB: 6000,
							}),
							mockUI.EXPECT().Say("Starting VPNKit..."),
							mockVpnKit.EXPECT().Start(),
							mockVpnKit.EXPECT().Watch(localExitChan),
							mockUI.EXPECT().Say("Starting the VM..."),
							mockHypervisor.EXPECT().Start("cfdev"),
							mockUI.EXPECT().Say("Waiting for the VM..."),
							mockProvisioner.EXPECT().Ping(),
							mockProvision.EXPECT().Execute(start.Args{Cpus: 7, Mem: 6000}),
							mockToggle.EXPECT().Enabled().Return(true),
							mockAnalyticsD.EXPECT().Start(),
							mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END),
						)

						Expect(startCmd.Execute(start.Args{
							Cpus: 7,
							Mem:  6000,
						})).To(Succeed())
					})
				})

				Context("and available mem < requested mem", func() {
					It("gives two warnings but starts anyways", func() {
						if runtime.GOOS == "darwin" {
							mockUI.EXPECT().Say("Installing cfdevd network helper...")
							mockCFDevD.EXPECT().Install()
						}

						gomock.InOrder(
							mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(5000), nil),
							mockSystemProfiler.EXPECT().GetTotalMemory().Return(uint64(5500), nil),
							mockHost.EXPECT().CheckRequirements(),
							mockHypervisor.EXPECT().IsRunning("cfdev").Return(false, nil),
							mockStop.EXPECT().RunE(nil, nil),
							mockEnv.EXPECT().CreateDirs(),

							mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
							mockUI.EXPECT().Say("Downloading Resources..."),
							mockCache.EXPECT().Sync(resource.Catalog{
								Items: []resource.Item{
									{Name: "some-item"},
									{Name: "cfdev-deps.tgz"},
								},
							}),
							mockUI.EXPECT().Say("Setting State..."),
							mockEnv.EXPECT().SetupState(filepath.Join(cacheDir, "cfdev-deps.tgz")),
							mockMetadataReader.EXPECT().Read(filepath.Join(stateDir, "metadata.yml")).Return(metadata, nil),

							mockToggle.EXPECT().SetProp("type", "cf"),
							mockToggle.EXPECT().SetProp("artifact", "some-artifact-version"),
							mockAnalyticsClient.EXPECT().PromptOptInIfNeeded(""),
							mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN, map[string]interface{}{
								"total memory":     uint64(5500),
								"available memory": uint64(5000),
							}),
							mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(1200), nil),
							mockUI.EXPECT().Say("WARNING: It is recommended that you run CF Dev with at least 8765 MB of RAM."),
							mockUI.EXPECT().Say("WARNING: This machine may not have enough available RAM to run with what is specified."),
							mockUI.EXPECT().Say("Creating the VM..."),
							mockHypervisor.EXPECT().CreateVM(hypervisor.VM{
								Name:     "cfdev",
								CPUs:     7,
								MemoryMB: 6000,
							}),
							mockUI.EXPECT().Say("Starting VPNKit..."),
							mockVpnKit.EXPECT().Start(),
							mockVpnKit.EXPECT().Watch(localExitChan),
							mockUI.EXPECT().Say("Starting the VM..."),
							mockHypervisor.EXPECT().Start("cfdev"),
							mockUI.EXPECT().Say("Waiting for the VM..."),
							mockProvisioner.EXPECT().Ping(),
							mockProvision.EXPECT().Execute(start.Args{Cpus: 7, Mem: 6000}),

							mockToggle.EXPECT().Enabled().Return(true),
							mockAnalyticsD.EXPECT().Start(),
							mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END),
						)

						Expect(startCmd.Execute(start.Args{
							Cpus: 7,
							Mem:  6000,
						})).To(Succeed())
					})
				})
			})
		})

		Context("when the -s flag is provided", func() {
			Context("arg is all", func() {
				It("deploys all the services", func() {
					if runtime.GOOS == "darwin" {
						mockUI.EXPECT().Say("Installing cfdevd network helper...")
						mockCFDevD.EXPECT().Install()
					}

					gomock.InOrder(
						mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(111), nil),
						mockSystemProfiler.EXPECT().GetTotalMemory().Return(uint64(222), nil),
						mockHost.EXPECT().CheckRequirements(),
						mockHypervisor.EXPECT().IsRunning("cfdev").Return(false, nil),
						mockStop.EXPECT().RunE(nil, nil),
						mockEnv.EXPECT().CreateDirs(),

						mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
						mockUI.EXPECT().Say("Downloading Resources..."),
						mockCache.EXPECT().Sync(resource.Catalog{
							Items: []resource.Item{
								{Name: "some-item"},
								{Name: "cfdev-deps.tgz"},
							},
						}),
						mockUI.EXPECT().Say("Setting State..."),
						mockEnv.EXPECT().SetupState(filepath.Join(cacheDir, "cfdev-deps.tgz")),
						mockMetadataReader.EXPECT().Read(filepath.Join(stateDir, "metadata.yml")).Return(metadata, nil),

						mockToggle.EXPECT().SetProp("type", "cf"),
						mockToggle.EXPECT().SetProp("artifact", "some-artifact-version"),
						mockAnalyticsClient.EXPECT().PromptOptInIfNeeded(""),
						mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN, map[string]interface{}{
							"total memory":     uint64(222),
							"available memory": uint64(111),
						}),
						mockAnalyticsClient.EXPECT().Event(cfanalytics.SELECTED_SERVICE, map[string]interface{}{"services_requested": "all"}),
						mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(10000), nil),
						mockUI.EXPECT().Say("Creating the VM..."),
						mockHypervisor.EXPECT().CreateVM(hypervisor.VM{
							Name:     "cfdev",
							CPUs:     7,
							MemoryMB: 8765,
						}),
						mockUI.EXPECT().Say("Starting VPNKit..."),
						mockVpnKit.EXPECT().Start(),
						mockVpnKit.EXPECT().Watch(localExitChan),
						mockUI.EXPECT().Say("Starting the VM..."),
						mockHypervisor.EXPECT().Start("cfdev"),
						mockUI.EXPECT().Say("Waiting for the VM..."),
						mockProvisioner.EXPECT().Ping(),
						mockProvision.EXPECT().Execute(start.Args{Cpus: 7, Mem: 0, DeploySingleService: "all"}),

						mockToggle.EXPECT().Enabled().Return(true),
						mockAnalyticsD.EXPECT().Start(),
						mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END),
					)

					Expect(startCmd.Execute(start.Args{
						Cpus:                7,
						Mem:                 0,
						DeploySingleService: "all",
					})).To(Succeed())
				})
			})

			Context("arg is multiple services", func() {
				It("deploys all the specified services", func() {
					if runtime.GOOS == "darwin" {
						mockUI.EXPECT().Say("Installing cfdevd network helper...")
						mockCFDevD.EXPECT().Install()
					}

					gomock.InOrder(
						mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(111), nil),
						mockSystemProfiler.EXPECT().GetTotalMemory().Return(uint64(222), nil),
						mockHost.EXPECT().CheckRequirements(),
						mockHypervisor.EXPECT().IsRunning("cfdev").Return(false, nil),
						mockStop.EXPECT().RunE(nil, nil),
						mockEnv.EXPECT().CreateDirs(),

						mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
						mockUI.EXPECT().Say("Downloading Resources..."),
						mockCache.EXPECT().Sync(resource.Catalog{
							Items: []resource.Item{
								{Name: "some-item"},
								{Name: "cfdev-deps.tgz"},
							},
						}),
						mockUI.EXPECT().Say("Setting State..."),
						mockEnv.EXPECT().SetupState(filepath.Join(cacheDir, "cfdev-deps.tgz")),
						mockMetadataReader.EXPECT().Read(filepath.Join(stateDir, "metadata.yml")).Return(metadata, nil),

						mockToggle.EXPECT().SetProp("type", "cf"),
						mockToggle.EXPECT().SetProp("artifact", "some-artifact-version"),
						mockAnalyticsClient.EXPECT().PromptOptInIfNeeded(""),
						mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN, map[string]interface{}{
							"total memory":     uint64(222),
							"available memory": uint64(111),
						}),
						mockAnalyticsClient.EXPECT().Event(cfanalytics.SELECTED_SERVICE, map[string]interface{}{"services_requested": "some-service-flagname,some-other-service-flagname"}),
						mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(10000), nil),
						mockUI.EXPECT().Say("Creating the VM..."),
						mockHypervisor.EXPECT().CreateVM(hypervisor.VM{
							Name:     "cfdev",
							CPUs:     7,
							MemoryMB: 8765,
						}),
						mockUI.EXPECT().Say("Starting VPNKit..."),
						mockVpnKit.EXPECT().Start(),
						mockVpnKit.EXPECT().Watch(localExitChan),
						mockUI.EXPECT().Say("Starting the VM..."),
						mockHypervisor.EXPECT().Start("cfdev"),
						mockUI.EXPECT().Say("Waiting for the VM..."),
						mockProvisioner.EXPECT().Ping(),
						mockProvision.EXPECT().Execute(start.Args{Cpus: 7, Mem: 0, DeploySingleService: "some-service-flagname,some-other-service-flagname"}),

						mockToggle.EXPECT().Enabled().Return(true),
						mockAnalyticsD.EXPECT().Start(),
						mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END),
					)

					Expect(startCmd.Execute(start.Args{
						Cpus:                7,
						Mem:                 0,
						DeploySingleService: "some-service-flagname,some-other-service-flagname",
					})).To(Succeed())
				})
			})

			Context("one of the args is an unsupported service", func() {
				It("returns an error message and does not execute start command", func() {
					if runtime.GOOS == "darwin" {
						mockUI.EXPECT().Say("Installing cfdevd network helper...")
						mockCFDevD.EXPECT().Install()
					}

					gomock.InOrder(
						mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(111), nil),
						mockSystemProfiler.EXPECT().GetTotalMemory().Return(uint64(222), nil),
						mockHost.EXPECT().CheckRequirements(),
						mockHypervisor.EXPECT().IsRunning("cfdev").Return(false, nil),
						mockStop.EXPECT().RunE(nil, nil),
						mockEnv.EXPECT().CreateDirs(),

						mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
						mockUI.EXPECT().Say("Downloading Resources..."),
						mockCache.EXPECT().Sync(resource.Catalog{
							Items: []resource.Item{
								{Name: "some-item"},
								{Name: "cfdev-deps.tgz"},
							},
						}),
						mockUI.EXPECT().Say("Setting State..."),
						mockEnv.EXPECT().SetupState(filepath.Join(cacheDir, "cfdev-deps.tgz")),
						mockMetadataReader.EXPECT().Read(filepath.Join(stateDir, "metadata.yml")).Return(metadata, nil),

						mockToggle.EXPECT().SetProp("type", "cf"),
						mockToggle.EXPECT().SetProp("artifact", "some-artifact-version"),
						mockAnalyticsClient.EXPECT().PromptOptInIfNeeded(""),
						mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN, map[string]interface{}{
							"total memory":     uint64(222),
							"available memory": uint64(111),
						}),
					)

					Expect(startCmd.Execute(start.Args{
						Cpus:                7,
						Mem:                 6666,
						DeploySingleService: "some-service-flagname,non-existent-service",
					}).Error()).To(ContainSubstring("is not supported"))
				})
			})
		})

		Context("when the -f flag is provided with an incompatible deps tarball version", func() {
			It("returns an error message and does not execute start command", func() {
				tarballFile := filepath.Join(tmpDir, "custom.tgz")
				ioutil.WriteFile(tarballFile, []byte{}, 0644)
				metadata.Version = "v100"

				if runtime.GOOS == "darwin" {
					mockUI.EXPECT().Say("Installing cfdevd network helper...")
					mockCFDevD.EXPECT().Install()
				}

				gomock.InOrder(
					mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(111), nil),
					mockSystemProfiler.EXPECT().GetTotalMemory().Return(uint64(222), nil),
					mockHost.EXPECT().CheckRequirements(),
					mockHypervisor.EXPECT().IsRunning("cfdev").Return(false, nil),
					mockStop.EXPECT().RunE(nil, nil),
					mockEnv.EXPECT().CreateDirs(),

					mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
					mockUI.EXPECT().Say("Downloading Resources..."),
					// don't download cfdev-deps that we won't use
					mockCache.EXPECT().Sync(resource.Catalog{
						Items: []resource.Item{
							{Name: "some-item"},
						},
					}),
					mockUI.EXPECT().Say("Setting State..."),
					mockEnv.EXPECT().SetupState(tarballFile),
					mockMetadataReader.EXPECT().Read(filepath.Join(stateDir, "metadata.yml")).Return(metadata, nil),

					mockToggle.EXPECT().SetProp("type", "cf"),
					mockToggle.EXPECT().SetProp("artifact", "some-artifact-version"),
				)

				Expect(startCmd.Execute(start.Args{
					Cpus:     7,
					Mem:      6666,
					DepsPath: tarballFile,
				})).To(MatchError(ContainSubstring("custom.tgz is not compatible with CF Dev. Please use a compatible file")))
			})
		})

		Context("when the -f flag is provided with an existing filepath", func() {
			It("starts the given tarball, doesn't download cfdev-deps, adds the tarball name as an analytics property", func() {
				customTarball := filepath.Join(tmpDir, "custom.tgz")
				ioutil.WriteFile(customTarball, []byte{}, 0644)

				if runtime.GOOS == "darwin" {
					mockUI.EXPECT().Say("Installing cfdevd network helper...")
					mockCFDevD.EXPECT().Install()
				}

				gomock.InOrder(
					mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(111), nil),
					mockSystemProfiler.EXPECT().GetTotalMemory().Return(uint64(222), nil),
					mockHost.EXPECT().CheckRequirements(),
					mockHypervisor.EXPECT().IsRunning("cfdev").Return(false, nil),
					mockStop.EXPECT().RunE(nil, nil),
					mockEnv.EXPECT().CreateDirs(),

					mockHostNet.EXPECT().AddLoopbackAliases("some-bosh-director-ip", "some-cf-router-ip"),
					mockUI.EXPECT().Say("Downloading Resources..."),
					// don't download cfdev-deps that we won't use
					mockCache.EXPECT().Sync(resource.Catalog{
						Items: []resource.Item{
							{Name: "some-item"},
						},
					}),
					mockUI.EXPECT().Say("Setting State..."),
					mockEnv.EXPECT().SetupState(customTarball),
					mockMetadataReader.EXPECT().Read(filepath.Join(stateDir, "metadata.yml")).Return(metadata, nil),

					mockToggle.EXPECT().SetProp("type", "cf"),
					mockToggle.EXPECT().SetProp("artifact", "some-artifact-version"),
					mockAnalyticsClient.EXPECT().PromptOptInIfNeeded(""),
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_BEGIN, map[string]interface{}{
						"total memory":     uint64(222),
						"available memory": uint64(111),
					}),
					mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(10000), nil),
					mockUI.EXPECT().Say("WARNING: It is recommended that you run CF Dev with at least 8765 MB of RAM."),
					mockUI.EXPECT().Say("Creating the VM..."),
					mockHypervisor.EXPECT().CreateVM(hypervisor.VM{
						Name:     "cfdev",
						CPUs:     7,
						MemoryMB: 6666,
					}),
					mockUI.EXPECT().Say("Starting VPNKit..."),
					mockVpnKit.EXPECT().Start(),
					mockVpnKit.EXPECT().Watch(localExitChan),
					mockUI.EXPECT().Say("Starting the VM..."),
					mockHypervisor.EXPECT().Start("cfdev"),
					mockUI.EXPECT().Say("Waiting for the VM..."),
					mockProvisioner.EXPECT().Ping(),
					mockProvision.EXPECT().Execute(start.Args{Cpus: 7, Mem: 6666, DepsPath: customTarball}),

					mockToggle.EXPECT().Enabled().Return(true),
					mockAnalyticsD.EXPECT().Start(),
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END),
				)

				Expect(startCmd.Execute(start.Args{
					Cpus:     7,
					Mem:      6666,
					DepsPath: customTarball,
				})).To(Succeed())
			})
		})

		Context("when linuxkit is already running", func() {
			It("says cf dev is already running", func() {
				gomock.InOrder(
					mockSystemProfiler.EXPECT().GetAvailableMemory().Return(uint64(111), nil),
					mockSystemProfiler.EXPECT().GetTotalMemory().Return(uint64(222), nil),
					mockHost.EXPECT().CheckRequirements(),
					mockHypervisor.EXPECT().IsRunning("cfdev").Return(true, nil),
					mockUI.EXPECT().Say("CF Dev is already running..."),
					mockAnalyticsClient.EXPECT().Event(cfanalytics.START_END, map[string]interface{}{"alreadyrunning": true}),
				)

				Expect(startCmd.Execute(start.Args{})).To(Succeed())
			})
		})
	})
})
