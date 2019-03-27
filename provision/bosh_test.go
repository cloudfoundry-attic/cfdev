package provision_test

import (
	"code.cloudfoundry.org/cfdev/provision"
	"code.cloudfoundry.org/cfdev/provision/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"time"
)

var _ = Describe("Bosh", func() {
	Describe("GetVMProgress", func() {
		var (
			b              provision.Bosh
			mockController *gomock.Controller
			mockRunner     *mocks.MockRunner
		)

		BeforeEach(func() {
			mockController = gomock.NewController(GinkgoT())
			mockRunner = mocks.NewMockRunner(mockController)

			b = provision.Bosh{
				Runner: mockRunner,
			}
		})

		AfterEach(func() {
			mockController.Finish()
		})

		Context("when the task is an errand", func() {
			It("returns the 'running errand' state", func() {
				result := b.GetVMProgress(time.Now(), "some-deployment", true)

				Expect(result.State).To(Equal(provision.RunningErrand))
			})
		})

		Context("when retrieving the vm progress fails", func() {
			It("returns the 'preparing' state", func() {
				mockRunner.EXPECT().Output(gomock.Any()).Return(nil, errors.New(""))

				result := b.GetVMProgress(time.Now(), "some-deployment", false)

				Expect(result.State).To(Equal(provision.Preparing))
			})
		})

		Context("when there are no VM instance results returned", func() {
			It("returns the 'preparing' state", func() {
				output := `
{
    "Tables": [
    ]
}
`

				mockRunner.EXPECT().Output(gomock.Any()).Return([]byte(output), nil)

				result := b.GetVMProgress(time.Now(), "some-deployment", false)

				Expect(result.State).To(Equal(provision.Preparing))
			})
		})

		Context("when there are no VM instance processes results returned", func() {
			It("returns the 'preparing' state", func() {
				output := `
{
    "Tables": [
        {
            "Content": "instances",
            "Header": {
                "az": "AZ",
                "instance": "Instance",
                "ips": "IPs",
                "process": "Process",
                "process_state": "Process State"
            },
            "Rows": [
            ],
            "Notes": null
        }
    ]
}
`

				mockRunner.EXPECT().Output(gomock.Any()).Return([]byte(output), nil)

				result := b.GetVMProgress(time.Now(), "some-deployment", false)

				Expect(result.State).To(Equal(provision.Preparing))
			})
		})

		Context("when vm instances returned without processes", func() {
			It("returns the 'deploying' state with the appropriate counts", func() {
				output := `
{
    "Tables": [
        {
            "Content": "instances",
            "Header": {
                "az": "AZ",
                "instance": "Instance",
                "ips": "IPs",
                "process": "Process",
                "process_state": "Process State"
            },
            "Rows": [
                {
                    "az": "null",
                    "instance": "instance/0",
                    "ips": "10.0.0.1",
                    "process": "",
                    "process_state": "running"
                },
                {
                    "az": "null",
                    "instance": "instance/0",
                    "ips": "10.0.0.1",
                    "process": "",
                    "process_state": "running"
                },
                {
                    "az": "null",
                    "instance": "instance/1",
                    "ips": "10.0.0.2",
                    "process": "",
                    "process_state": "running"
                }
            ],
            "Notes": null
        }
    ]
}
`

				mockRunner.EXPECT().Output(gomock.Any()).Return([]byte(output), nil)

				result := b.GetVMProgress(time.Now(), "some-deployment", false)

				Expect(result.State).To(Equal(provision.Deploying))
				Expect(result.Total).To(Equal(2))
				Expect(result.Done).To(Equal(0))
			})
		})

		Context("when vm instances are returned with processes not yet completed", func() {
			It("returns the 'deploying' state with the appropriate counts", func() {
				output := `
{
    "Tables": [
        {
            "Content": "instances",
            "Header": {
                "az": "AZ",
                "instance": "Instance",
                "ips": "IPs",
                "process": "Process",
                "process_state": "Process State"
            },
            "Rows": [
                {
                    "az": "null",
                    "instance": "instance/0",
                    "ips": "10.0.0.1",
                    "process": "process-0",
                    "process_state": "starting"
                },
                {
                    "az": "null",
                    "instance": "instance/0",
                    "ips": "10.0.0.1",
                    "process": "process-1",
                    "process_state": "running"
                },
                {
                    "az": "null",
                    "instance": "instance/1",
                    "ips": "10.0.0.2",
                    "process": "",
                    "process_state": "running"
                }
            ],
            "Notes": null
        }
    ]
}
`

				mockRunner.EXPECT().Output(gomock.Any()).Return([]byte(output), nil)

				result := b.GetVMProgress(time.Now(), "some-deployment", false)

				Expect(result.State).To(Equal(provision.Deploying))
				Expect(result.Total).To(Equal(2))
				Expect(result.Done).To(Equal(0))
			})
		})

		Context("when vm instances are returned with processes completed", func() {
			It("returns the 'deploying' state with the appropriate counts", func() {
				output := `
{
    "Tables": [
        {
            "Content": "instances",
            "Header": {
                "az": "AZ",
                "instance": "Instance",
                "ips": "IPs",
                "process": "Process",
                "process_state": "Process State"
            },
            "Rows": [
                {
                    "az": "null",
                    "instance": "instance/0",
                    "ips": "10.0.0.1",
                    "process": "process-0",
                    "process_state": "running"
                },
                {
                    "az": "null",
                    "instance": "instance/0",
                    "ips": "10.0.0.1",
                    "process": "process-1",
                    "process_state": "running"
                },
                {
                    "az": "null",
                    "instance": "instance/1",
                    "ips": "10.0.0.2",
                    "process": "",
                    "process_state": "running"
                }
            ],
            "Notes": null
        }
    ]
}
`

				mockRunner.EXPECT().Output(gomock.Any()).Return([]byte(output), nil)

				result := b.GetVMProgress(time.Now(), "some-deployment", false)

				Expect(result.State).To(Equal(provision.Deploying))
				Expect(result.Total).To(Equal(2))
				Expect(result.Done).To(Equal(1))
			})
		})
	})
})
