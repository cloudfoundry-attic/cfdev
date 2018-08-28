// +build darwin

package privileged_test

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"os/exec"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const SOCK = "/var/tmp/cfdevd.socket"

var _ = Describe("cfdevd test", func() {
	var bin string

	BeforeSuite(func() {
		var err error
		session, err := gexec.Start(exec.Command("sudo", "--non-interactive", "launchctl", "remove", "org.cloudfoundry.cfdevd"), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(), "You may need to log sudo in")
		Expect(string(session.Out.Contents())).ShouldNot(ContainSubstring("org.cloudfoundry.cfdevd"))

		bin, err = gexec.Build("code.cloudfoundry.org/cfdev/cfdevd")
		Expect(err).NotTo(HaveOccurred())
		session, err = gexec.Start(exec.Command("sudo", "--non-interactive", bin, "install"), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))

		session, err = gexec.Start(exec.Command("sudo", "--non-interactive", "launchctl", "list"), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).Should(ContainSubstring("org.cloudfoundry.cfdevd"))

		conn, err := net.DialUnix("unix", nil, &net.UnixAddr{
			Net:  "unix",
			Name: SOCK,
		})
		Expect(err).NotTo(HaveOccurred())
		defer conn.Close()

		Expect(sendHello(conn, "VMN3T", 22, "0123456789012345678901234567890123456789")).To(Succeed())
		Expect(recvHello(conn)).To(Equal("CFD3V"))
		sendAddAlias(conn)

		session, err = gexec.Start(exec.Command("ifconfig"), GinkgoWriter, GinkgoWriter)
		Expect(err).NotTo(HaveOccurred())
		Eventually(session).Should(gexec.Exit(0))
		Expect(string(session.Out.Contents())).Should(ContainSubstring("10.245.0.2"))
		Expect(string(session.Out.Contents())).Should(ContainSubstring("10.144.0.34"))
	})

	AfterSuite(func() {
		conn, err := net.DialUnix("unix", nil, &net.UnixAddr{
			Net:  "unix",
			Name: SOCK,
		})
		Expect(err).NotTo(HaveOccurred())
		defer conn.Close()

		Expect(sendHello(conn, "VMN3T", 22, "0123456789012345678901234567890123456789")).To(Succeed())
		Expect(recvHello(conn)).To(Equal("CFD3V"))
		sendRemoveAlias(conn)

		conn, err = net.DialUnix("unix", nil, &net.UnixAddr{
			Net:  "unix",
			Name: SOCK,
		})
		Expect(err).NotTo(HaveOccurred())
		defer conn.Close()

		Expect(sendHello(conn, "VMN3T", 22, "0123456789012345678901234567890123456789")).To(Succeed())
		Expect(recvHello(conn)).To(Equal("CFD3V"))
		Expect(sendUninstall(conn)).To(Succeed())

		Eventually(func() (string, error) {
			session, err := gexec.Start(exec.Command("sudo", "--non-interactive", "launchctl", "list"), GinkgoWriter, GinkgoWriter)
			Eventually(session).Should(gexec.Exit(0))
			return string(session.Out.Contents()), err
		}).ShouldNot(ContainSubstring("org.cloudfoundry.cfdevd"))

		gexec.KillAndWait()
		gexec.CleanupBuildArtifacts()
	})

	var conn *net.UnixConn
	BeforeEach(func() {
		var err error
		conn, err = net.DialUnix("unix", nil, &net.UnixAddr{
			Net:  "unix",
			Name: SOCK,
		})
		Expect(err).NotTo(HaveOccurred())
	})
	AfterEach(func() {
		conn.Close()
	})

	It("binds ports on bosh ip", func() {
		Expect(sendHello(conn, "VMN3T", 22, "0123456789012345678901234567890123456789")).To(Succeed())
		Expect(recvHello(conn)).To(Equal("CFD3V"))
		Expect(sendBindAddr(conn, "10.245.0.2", 1777)).To(Succeed())
		ln, _, err := recvBindAddr(conn, "10.245.0.2", 1777)
		Expect(err).NotTo(HaveOccurred())
		defer ln.Close()

		msg := "Hello from test"
		go sendMessage("10.245.0.2:1777", msg)
		Expect(readFromListener(ln)).To(Equal(msg))
	})

	It("binds ports on gorouter ip", func() {
		Expect(sendHello(conn, "VMN3T", 22, "0123456789012345678901234567890123456789")).To(Succeed())
		Expect(recvHello(conn)).To(Equal("CFD3V"))
		Expect(sendBindAddr(conn, "10.144.0.34", 1888)).To(Succeed())
		ln, _, err := recvBindAddr(conn, "10.144.0.34", 1888)
		Expect(err).NotTo(HaveOccurred())
		defer ln.Close()

		msg := "Hello from test"
		go sendMessage("10.144.0.34:1888", msg)
		Expect(readFromListener(ln)).To(Equal(msg))
	})

	It("refuses to bind ports on other interfaces", func() {
		Expect(sendHello(conn, "VMN3T", 22, "0123456789012345678901234567890123456789")).To(Succeed())
		Expect(recvHello(conn)).To(Equal("CFD3V"))
		Expect(sendBindAddr(conn, "127.0.0.1", 1888)).To(Succeed())
		_, b, _ := recvBindAddr(conn, "10.245.0.2", 1888)
		Expect(b[0]).To(Equal(uint8(71)))
	})

	Context("binding to a bound port", func() {
		var prior net.Listener
		BeforeEach(func() {
			var err error
			prior, err = net.Listen("tcp", "10.245.0.2:1999")
			Expect(err).NotTo(HaveOccurred())
		})
		AfterEach(func() { prior.Close() })

		It("sends an error", func() {
			Expect(sendHello(conn, "VMN3T", 22, "0123456789012345678901234567890123456789")).To(Succeed())
			Expect(recvHello(conn)).To(Equal("CFD3V"))
			Expect(sendBindAddr(conn, "10.245.0.2", 1999)).To(Succeed())
			_, b, _ := recvBindAddr(conn, "10.245.0.2", 1999)
			Expect(b[0]).To(Equal(uint8(48)))
		})
	})
})

func sendAddAlias(conn *net.UnixConn) {
	var instruction uint8 = 3
	_, err := conn.Write([]byte{instruction})
	Expect(err).NotTo(HaveOccurred())

	bytes := make([]byte, 1, 1)
	_, err = io.ReadFull(conn, bytes)
	Expect(err).NotTo(HaveOccurred())
	Expect(bytes[0]).To(BeZero())
}

func sendRemoveAlias(conn *net.UnixConn) {
	var instruction uint8 = 2
	_, err := conn.Write([]byte{instruction})
	Expect(err).NotTo(HaveOccurred())

	bytes := make([]byte, 1, 1)
	_, err = io.ReadFull(conn, bytes)
	Expect(err).NotTo(HaveOccurred())
	Expect(bytes[0]).To(BeZero())
}

func sendUninstall(conn *net.UnixConn) error {
	var instruction uint8 = 1
	_, err := conn.Write([]byte{instruction})
	return err
}

func sendHello(conn *net.UnixConn, id string, version uint32, sha1 string) error {
	if _, err := conn.Write([]byte(id)); err != nil {
		return err
	}
	if err := binary.Write(conn, binary.LittleEndian, version); err != nil {
		return err
	}
	_, err := conn.Write([]byte(sha1))
	return err
}

func recvHello(conn *net.UnixConn) (string, error) {
	bytes := make([]byte, 49, 49)
	if n, err := io.ReadFull(conn, bytes); err != nil {
		return "", err
	} else if n != 49 {
		return "", fmt.Errorf("Expected to read 49 bytes, read %d", n)
	}
	return string(bytes[0:5]), nil
}

func sendBindAddr(conn *net.UnixConn, ip string, port uint16) error {
	var instruction uint8 = 6
	conn.Write([]byte{instruction})
	b := []byte(net.ParseIP(ip).To4())
	conn.Write(append([]byte{}, b[3], b[2], b[1], b[0]))
	binary.Write(conn, binary.LittleEndian, port)
	_, err := conn.Write([]byte{0x0})
	return err
}

func recvBindAddr(conn *net.UnixConn, ip string, port uint16) (net.Listener, []byte, error) {
	b := make([]byte, 8, 8)
	oob := make([]byte, 16, 16)
	if _, _, _, _, err := conn.ReadMsgUnix(b, oob); err != nil {
		return nil, b, err
	}
	if b[0] != 0 {
		return nil, b, fmt.Errorf("Look at b: %d, %+v", b[0], b)
	}
	scms, err := syscall.ParseSocketControlMessage(oob)
	if err != nil {
		return nil, b, err
	}
	fds, err := syscall.ParseUnixRights(&scms[0])
	if err != nil {
		return nil, b, err
	}
	syscall.Listen(fds[0], 65536)
	defer syscall.Close(fds[0])
	file := os.NewFile(uintptr(fds[0]), fmt.Sprintf("tcp:%s:%d", ip, port))
	defer file.Close()
	ln, err := net.FileListener(file)
	return ln, b, err
}

func sendMessage(address string, mesg string) {
	defer GinkgoRecover()
	wconn, err := net.Dial("tcp", address)
	Expect(err).NotTo(HaveOccurred())
	defer wconn.Close()
	wconn.Write([]byte(mesg))
}

func readFromListener(ln net.Listener) (string, error) {
	conn, err := ln.Accept()
	if err != nil {
		return "", err
	}
	defer conn.Close()
	received := make([]byte, 15, 15)
	_, err = conn.Read(received)
	if err != nil {
		return "", err
	}
	return string(received), nil
}
