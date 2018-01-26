package garden_test

import (
	"errors"
	"fmt"

	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden/gardenfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Fetching BOSH Configuration", func() {
	var (
		fakeClient *gardenfakes.FakeClient
		boshConfig gdn.BOSHConfiguration
		err        error
	)

	BeforeEach(func() {
		fakeClient = new(gardenfakes.FakeClient)
		fakeClient.CreateReturns(nil, errors.New("some error"))
	})

	JustBeforeEach(func() {
		boshConfig, err = gdn.FetchBOSHConfig(fakeClient)
	})

	It("creates a container", func() {
		Expect(fakeClient.CreateCallCount()).To(Equal(1))
		spec := fakeClient.CreateArgsForCall(0)

		Expect(spec).To(Equal(garden.ContainerSpec{
			Handle:     "fetch-bosh-config",
			Privileged: true,
			Network:    "10.246.0.0/16",
			Image: garden.ImageRef{
				URI: "/var/vcap/cache/workspace.tar",
			},
			BindMounts: []garden.BindMount{
				{
					SrcPath: "/var/vcap/director",
					DstPath: "/var/vcap/director",
					Mode:    garden.BindMountModeRW,
				},
			},
		}))
	})

	Context("creating the container succeeds", func() {
		var (
			fakeContainer *gardenfakes.FakeContainer
		)

		BeforeEach(func() {
			fakeContainer = new(gardenfakes.FakeContainer)
			fakeContainer.RunReturns(nil, errors.New("some error"))
			fakeClient.CreateReturns(fakeContainer, nil)
		})

		It("starts to fetch the bosh config", func() {
			Expect(fakeContainer.RunCallCount()).To(Equal(1))

			spec, _ := fakeContainer.RunArgsForCall(0)
			Expect(spec).To(Equal(garden.ProcessSpec{
				Path: "cat",
				Args: []string{"/var/vcap/director/creds.yml"},
				User: "root",
			}))
		})

		Context("when fetching the bosh config succeeds", func() {
			BeforeEach(func() {
				fakeContainer.RunStub = successfulRunStub
			})

			It("returns without an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("returns the configuration", func() {
				Expect(boshConfig).Should(Equal(gdn.BOSHConfiguration{
					AdminUsername:   "admin",
					AdminPassword:   "admin-password",
					CACertificate:   "ca-certificate",
					DirectorAddress: "10.245.0.2",

					GatewayHost:       "10.245.0.2",
					GatewayPrivateKey: "ssh-private-key",
					GatewayUsername:   "jumpbox",
				}))
			})

			It("deletes the container", func() {
				Expect(fakeClient.DestroyCallCount()).To(Equal(1))
				handle := fakeClient.DestroyArgsForCall(0)
				Expect(handle).To(Equal("fetch-bosh-config"))
			})
		})

		Context("when fetching does not start", func() {
			BeforeEach(func() {
				fakeContainer.RunReturns(nil, errors.New("unable to start process"))
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("unable to start process"))
			})
		})

		Context("when fetching finishes with a non-zero exit code", func() {
			BeforeEach(func() {
				process := new(gardenfakes.FakeProcess)
				process.WaitReturns(23, nil)
				fakeContainer.RunReturns(process, nil)
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("process exited with status 23"))
			})
		})

		Context("when fetching finishes with invalid yaml", func() {
			BeforeEach(func() {
				fakeContainer.RunStub = invalidYAMLRunStub
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when fetching finishes with missing config values", func() {
			BeforeEach(func() {
				fakeContainer.RunStub = missingConfigRunStub
			})

			It("returns an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

		Context("when we cannot determine the state of the deploy", func() {
			BeforeEach(func() {
				process := new(gardenfakes.FakeProcess)
				process.WaitReturns(-10, errors.New("connection to garden lost"))
				fakeContainer.RunReturns(process, nil)
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("connection to garden lost"))
			})
		})
	})

	Context("creating the container fails", func() {
		BeforeEach(func() {
			fakeClient.CreateReturns(nil, errors.New("unable to create container"))
		})

		It("forwards the error", func() {
			Expect(err).To(MatchError("unable to create container"))
		})
	})
})
var successfulRunStub = func(spec garden.ProcessSpec, io garden.ProcessIO) (garden.Process, error) {
	const config = `
director_ssl:
  ca: ca-certificate
jumpbox_ssh:
  private_key: ssh-private-key
admin_password: admin-password
`

	fmt.Fprint(io.Stdout, config)
	process := new(gardenfakes.FakeProcess)
	process.WaitReturns(0, nil)
	return process, nil

}

var invalidYAMLRunStub = func(spec garden.ProcessSpec, io garden.ProcessIO) (garden.Process, error) {
	const config = `| invalid yaml`

	fmt.Fprint(io.Stdout, config)
	process := new(gardenfakes.FakeProcess)
	process.WaitReturns(0, nil)
	return process, nil

}

var missingConfigRunStub = func(spec garden.ProcessSpec, io garden.ProcessIO) (garden.Process, error) {
	fmt.Fprint(io.Stdout, "")
	process := new(gardenfakes.FakeProcess)
	process.WaitReturns(0, nil)
	return process, nil

}
