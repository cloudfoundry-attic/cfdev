// +build darwin

package cmd_test

import (
	"errors"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/pkg/cfdevd/cmd"
	"code.cloudfoundry.org/cfdev/pkg/cfdevd/cmd/mocks"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ bool = Describe("UninstallCommand", func() {
	Describe("Execute", func() {
		var (
			controller       *gomock.Controller
			mockDaemonRunner *mocks.MockDaemonRunner
			uninstall        *cmd.UninstallCommand

			socketDir  string
			conn       *net.UnixConn
			recvdBytes chan uint8
		)
		BeforeEach(func() {
			controller = gomock.NewController(GinkgoT())
			mockDaemonRunner = mocks.NewMockDaemonRunner(controller)
			uninstall = &cmd.UninstallCommand{
				DaemonRunner: mockDaemonRunner,
			}
		})
		BeforeEach(func() {
			var err error
			recvdBytes = make(chan uint8)
			socketDir, err = ioutil.TempDir(os.Getenv("TMPDIR"), "socket")
			Expect(err).NotTo(HaveOccurred())
			addr := &net.UnixAddr{
				Name: filepath.Join(socketDir, "some.socket"),
			}
			ln, err := net.ListenUnix("unix", addr)
			Expect(err).NotTo(HaveOccurred())
			go func() {
				defer GinkgoRecover()
				conn, err := ln.Accept()
				Expect(err).NotTo(HaveOccurred())
				defer conn.Close()
				data := make([]byte, 1, 1)
				_, err = conn.Read(data)
				Expect(err).NotTo(HaveOccurred())
				recvdBytes <- uint8(data[0])
			}()
			conn, err = net.DialUnix("unix", nil, addr)
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() {
			os.RemoveAll(socketDir)
			conn.Close()
			controller.Finish()
		})
		It("removes the correct daemon", func() {
			mockDaemonRunner.EXPECT().RemoveDaemon("org.cloudfoundry.cfdevd")
			Expect(uninstall.Execute(conn)).To(Succeed())
		})
		It("sends 0 (success) over the communication socket", func() {
			mockDaemonRunner.EXPECT().RemoveDaemon(gomock.Any())
			Expect(uninstall.Execute(conn)).To(Succeed())
			Expect(<-recvdBytes).To(Equal(uint8(0)))
		})

		Context("RemoveDaemon fails", func() {
			BeforeEach(func() {
				mockDaemonRunner.EXPECT().RemoveDaemon(gomock.Any()).Return(errors.New("Mega Fail"))
			})
			It("returns the failure from launchd", func() {
				Expect(uninstall.Execute(conn)).To(Equal(errors.New("Mega Fail")))
			})
			It("sends 1 (failure) over the communication socket", func() {
				uninstall.Execute(conn)
				Expect(<-recvdBytes).To(Equal(uint8(1)))
			})
		})
	})
})
