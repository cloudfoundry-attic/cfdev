package client_test

import (
	"encoding/binary"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cfdevd/client"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CFDevD Client", func() {
	var (
		tmpDir     string
		socketPath string
		subject    *client.Client
	)
	BeforeEach(func() {
		tmpDir, _ = ioutil.TempDir("", "cfdevd.")
		socketPath = filepath.Join(tmpDir, "cfdevd.socket")
		subject = client.New("TEST1", socketPath)
	})
	AfterEach(func() { os.RemoveAll(tmpDir) })

	Context("cfdevd socket exists", func() {
		var instructions chan byte
		var uninstallErrorCode int
		BeforeEach(func() {
			instructions = make(chan byte, 1)
			ln, err := net.Listen("unix", socketPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(socketPath).To(BeAnExistingFile())
			go func() {
				conn, err := ln.Accept()
				Expect(err).NotTo(HaveOccurred())
				handshake := make([]byte, 49, 49)
				binary.Read(conn, binary.LittleEndian, handshake)
				binary.Write(conn, binary.LittleEndian, handshake)
				instruction := make([]byte, 1, 1)
				binary.Read(conn, binary.LittleEndian, instruction)
				instructions <- instruction[0]
				if uninstallErrorCode == -1 {
					conn.Close()
				} else {
					binary.Write(conn, binary.LittleEndian, []byte{byte(uninstallErrorCode)})
				}
			}()
		})
		It("succeeds and sends the uninstall command to cfdevd", func() {
			uninstallErrorCode = 0
			serverName, err := subject.Uninstall()
			Expect(err).ToNot(HaveOccurred())

			Eventually(instructions).Should(Receive(Equal(byte(1))))
			Expect(serverName).To(Equal("TEST1"))
		})
		Context("cfdevd stops after receiving uninstall command, thus closes connection before writing success code", func() {
			It("succeeds", func() {
				uninstallErrorCode = -1
				_, err := subject.Uninstall()
				Expect(err).ToNot(HaveOccurred())

				Eventually(instructions).Should(Receive(Equal(byte(1))))
			})
		})
		Context("cfdevd returns error to uninstall", func() {
			It("returns the error", func() {
				uninstallErrorCode = 1
				_, err := subject.Uninstall()
				Expect(err).To(MatchError("failed to uninstall cfdevd: errorcode: 1"))
			})
		})
	})
	Context("cfdevd socket is specified but does not exist", func() {
		It("succeeds", func() {
			_, err := subject.Uninstall()
			Expect(err).ToNot(HaveOccurred())
		})
	})
})
