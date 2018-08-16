package garden_test

import (
	"bytes"
	gdn "code.cloudfoundry.org/cfdev/garden"
	"code.cloudfoundry.org/garden"
	"code.cloudfoundry.org/garden/gardenfakes"
	"errors"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

var _ = Describe("Logs", func() {
	var (
		fakeClient     *gardenfakes.FakeClient
		err            error
		destinationDir string
	)

	BeforeEach(func() {
		fakeClient = new(gardenfakes.FakeClient)
		fakeClient.CreateReturns(nil, errors.New("some error"))
	})

	AfterEach(func() {
		os.RemoveAll(destinationDir)
	})

	JustBeforeEach(func() {
		err = gdn.FetchLogs(fakeClient, destinationDir)
	})

	It("creates a container", func() {
		Expect(fakeClient.CreateCallCount()).To(Equal(1))
		spec := fakeClient.CreateArgsForCall(0)

		Expect(spec).To(Equal(garden.ContainerSpec{
			Handle:     "fetch-logs",
			Privileged: true,
			Network:    "10.246.0.0/16",
			Image: garden.ImageRef{
				URI: "/var/vcap/cache/workspace.tar",
			},
			BindMounts: []garden.BindMount{
				{
					SrcPath: "/var/vcap",
					DstPath: "/var/vcap",
					Mode:    garden.BindMountModeRO,
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
			fakeClient.CreateReturns(fakeContainer, nil)

			var err error
			destinationDir, err = ioutil.TempDir("", "cfdev-test-")
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(fakeClient.DestroyCallCount()).To(Equal(1))
			Expect(fakeClient.DestroyArgsForCall(0)).To(Equal("fetch-logs"))
		})

		Context("retrieving the logs succeeds", func() {
			BeforeEach(func() {

				fakeContainer.StreamOutReturns(newFakeReadCloser("some-tar-file"), nil)
			})

			It("puts the logs onto the file system", func() {
				Expect(fakeContainer.StreamOutCallCount()).To(Equal(1))

				arg := fakeContainer.StreamOutArgsForCall(0)
				Expect(arg).To(Equal(garden.StreamOutSpec{
					Path: "/var/vcap/logs",
				}))

				Expect(filepath.Join(destinationDir, "cfdev-logs.tgz")).To(BeAnExistingFile())
			})
		})

		Context("retrieving the logs succeeds but destination dir does not exist", func() {
			BeforeEach(func() {
				destinationDir = filepath.Join(destinationDir, "banana")

				fakeContainer.StreamOutReturns(newFakeReadCloser("some-tar-file"), nil)
			})

			It("creates the dir before putting the logs onto the file system", func() {
				Expect(fakeContainer.StreamOutCallCount()).To(Equal(1))

				arg := fakeContainer.StreamOutArgsForCall(0)
				Expect(arg).To(Equal(garden.StreamOutSpec{
					Path: "/var/vcap/logs",
				}))

				Expect(filepath.Join(destinationDir, "cfdev-logs.tgz")).To(BeAnExistingFile())
			})
		})

		Context("when the stream out invocation fails", func() {
			BeforeEach(func() {
				fakeContainer.StreamOutReturns(nil, errors.New("some-stream-out-error"))
			})

			It("returns the error", func() {
				Expect(err).To(MatchError("some-stream-out-error"))
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

func newFakeReadCloser(contents string) *fakeReadCloser {
	return &fakeReadCloser{
		bytes.NewBufferString(contents),
	}
}

type fakeReadCloser struct {
	io.Reader
}

func (fakeReadCloser) Close() error { return nil }
