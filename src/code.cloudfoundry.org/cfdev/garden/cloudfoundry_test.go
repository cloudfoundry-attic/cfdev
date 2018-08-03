package garden_test

import (
	"errors"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden/gardenfakes"
	"gopkg.in/yaml.v2"
)

var _ = Describe("DeployCloudFoundry", func() {
	var (
		fakeClient       *gardenfakes.FakeClient
		err              error
		dockerRegistries []string
		gclient          *gdn.Garden
	)

	BeforeEach(func() {
		fakeClient = new(gardenfakes.FakeClient)
		fakeClient.CreateReturns(nil, errors.New("some error"))
		gclient = &gdn.Garden{Client: fakeClient}
	})

	JustBeforeEach(func() {
		err = gclient.DeployCloudFoundry(dockerRegistries)
	})

	It("creates a container", func() {
		Expect(fakeClient.CreateCallCount()).To(Equal(1))
		spec := fakeClient.CreateArgsForCall(0)

		Expect(spec).To(Equal(garden.ContainerSpec{
			Handle:     "deploy-cf",
			Privileged: true,
			Network:    "10.246.0.0/16",
			Image: garden.ImageRef{
				URI: "/var/vcap/cache/workspace.tar",
			},
			BindMounts: []garden.BindMount{
				{
					SrcPath: "/var/vcap",
					DstPath: "/var/vcap",
					Mode:    garden.BindMountModeRW,
				},
				{
					SrcPath: "/var/vcap/cache",
					DstPath: "/var/vcap/cache",
					Mode:    garden.BindMountModeRO,
				},
			},
		}))
	})

	Context("when a list of docker registries is provided", func() {
		BeforeEach(func() {
			dockerRegistries = []string{
				"host.cfdev.sh:5000",
				"host.cfdev.sh:5001",
			}
		})

		It("sets the DOCKER_REGISTRIES variable when creating the container", func() {
			spec := fakeClient.CreateArgsForCall(0)
			Expect(spec.Env).To(ContainElement(HavePrefix("DOCKER_REGISTRIES=")))
		})

		It("sets the DOCKER_REGISTRIES value to be a yaml array of registries", func() {
			spec := fakeClient.CreateArgsForCall(0)
			result := strings.SplitN(spec.Env[0], "=", 2)

			var registries []string

			err := yaml.Unmarshal([]byte(result[1]), &registries)
			Expect(err).ToNot(HaveOccurred())
			Expect(result[0]).To(Equal("DOCKER_REGISTRIES"))
			Expect(registries).To(ConsistOf("host.cfdev.sh:5000", "host.cfdev.sh:5001"))
		})
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

		It("starts to deploy cloud foundry", func() {
			Expect(fakeContainer.RunCallCount()).To(Equal(1))

			spec, io := fakeContainer.RunArgsForCall(0)
			Expect(io).To(Equal(garden.ProcessIO{}))
			Expect(spec).To(Equal(garden.ProcessSpec{
				ID:   "deploy-cf",
				Path: "/bin/bash",
				Args: []string{"/var/vcap/cache/bin/deploy-cf"},
				User: "root",
			}))
		})

		Context("when deploying cloud foundry succeeds", func() {
			BeforeEach(func() {
				process := new(gardenfakes.FakeProcess)
				process.WaitReturns(0, nil)
				fakeContainer.RunReturns(process, nil)
			})

			It("returns without an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("deletes the container", func() {
				Expect(fakeClient.DestroyCallCount()).To(Equal(1))
				handle := fakeClient.DestroyArgsForCall(0)
				Expect(handle).To(Equal("deploy-cf"))
			})
		})

		Context("when the deploy cannot start", func() {
			BeforeEach(func() {
				fakeContainer.RunReturns(nil, errors.New("unable to start process"))
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("unable to start process"))
			})
		})

		Context("when the deploy finishes with a non-zero exit code", func() {
			BeforeEach(func() {
				process := new(gardenfakes.FakeProcess)
				process.WaitReturns(23, nil)
				fakeContainer.RunReturns(process, nil)
			})

			It("returns an error", func() {
				Expect(err).To(MatchError("process exited with status 23"))
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
