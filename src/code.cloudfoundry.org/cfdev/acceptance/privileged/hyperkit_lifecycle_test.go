package privileged_test

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"time"

	"io"
	"io/ioutil"
	"net/http"
	"syscall"

	. "code.cloudfoundry.org/cfdev/acceptance"
	"code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("hyperkit lifecycle", func() {
	var (
		cfdevHome       string
		hyperkitPidPath string
	)

	BeforeEach(func() {
		Expect(HasSudoPrivilege()).To(BeTrue(), "Please run 'sudo echo hi' first")
		RemoveIPAliases(BoshDirectorIP, CFRouterIP)
		FullCleanup()

		cfdevHome = os.Getenv("CFDEV_HOME")
		if cfdevHome == "" {
			cfdevHome = filepath.Join(os.Getenv("HOME"), ".cfdev")
		}
		hyperkitPidPath = filepath.Join(cfdevHome, "state", "linuxkit", "hyperkit.pid")

		session := cf.Cf("install-plugin", pluginPath, "-f")
		Eventually(session).Should(gexec.Exit(0))
		session = cf.Cf("plugins")
		Eventually(session).Should(gexec.Exit(0))
	})

	AfterEach(func() {
		gexec.KillAndWait()
		RemoveIPAliases(BoshDirectorIP, CFRouterIP)

		session := cf.Cf("dev", "stop")
		Eventually(session).Should(gexec.Exit(0))
	})

	It("runs the entire vm lifecycle", func() {
		var session *gexec.Session
		isoPath := os.Getenv("ISO_PATH")
		if isoPath != "" {
			session = cf.Cf("dev", "start", "-f", isoPath, "-m", "8192")
		} else {
			session = cf.Cf("dev", "start")
		}
		Eventually(session, 20*time.Minute).Should(gbytes.Say("Starting VPNKit"))

		Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.vpnkit"), 10, 1).Should(BeTrue())
		Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.linuxkit"), 10, 1).Should(BeTrue())

		By("waiting for garden to listen")
		client := client.New(connection.New("tcp", "localhost:8888"))
		Eventually(client.Ping, 360).Should(Succeed())

		EventuallyWeCanTargetTheBOSHDirector()

		By("waiting for cfdev cli to exit when the deploy finished")
		Eventually(session, 3600).Should(gexec.Exit(0))

		By("waiting for cf router to listen")
		session = cf.Cf("login", "-a", "https://api.v3.pcfdev.io", "--skip-ssl-validation", "-u", "admin", "-p", "admin")
		Eventually(session).Should(gexec.Exit(0))

		By("pushing an app")
		PushAnApp()

		hyperkitPid := PidFromFile(hyperkitPidPath)

		By("deploy finished - stopping...")
		session = cf.Cf("dev", "stop")
		Eventually(session).Should(gexec.Exit(0))

		//ensure pid is not running
		Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.linuxkit"), 5, 1).Should(BeFalse())
		EventuallyProcessStops(hyperkitPid, 5)
		Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.vpnkit"), 5, 1).Should(BeFalse())
	})

	Context("Run with", func() {
		var assetUrl = "https://s3.amazonaws.com/cfdev-test-assets/test-deps.dev"
		var assetDir string

		BeforeEach(func() {
			var err error
			assetDir, err = ioutil.TempDir(os.TempDir(), "asset")
			Expect(err).ToNot(HaveOccurred())
			downloadTestAsset(assetDir, assetUrl)
		})

		AfterEach(func() {
			err := os.RemoveAll(assetDir)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Custom ISO", func() {
			session := cf.Cf("dev", "start", "-f", filepath.Join(assetDir, "test-deps.dev"))
			Eventually(session, 20*time.Minute).Should(gbytes.Say("Starting VPNKit"))

			By("settingup VPNKit dependencies")
			Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.vpnkit"), 10, 1).Should(BeTrue())
			Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.linuxkit"), 10, 1).Should(BeTrue())

			By("waiting for garden to listen")
			EventuallyShouldListenAt("http://"+GardenIP+":8888", 360)

			client := client.New(connection.New("tcp", "localhost:8888"))
			Eventually(func() (string, error) {
				return GetFile(client, "deploy-bosh", "/var/vcap/cache/test-file-one.txt")
			}).Should(Equal("testfileone\n"))

			session.Terminate()
			Eventually(session).Should(gexec.Exit())

			hyperkitPid := PidFromFile(hyperkitPidPath)

			By("deploy finished - stopping...")
			session = cf.Cf("dev", "stop")
			Eventually(session).Should(gexec.Exit(0))

			//ensure pid is not running
			Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.linuxkit"), 5, 1).Should(BeFalse())
			EventuallyProcessStops(hyperkitPid, 5)
			Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.vpnkit"), 5, 1).Should(BeFalse())
		})
	})
})

func EventuallyWeCanTargetTheBOSHDirector() {
	By("waiting for bosh to listen")
	EventuallyShouldListenAt("https://"+BoshDirectorIP+":25555", 480)

	// Even though the test below is very similar this fails fast when `bosh env`
	// command is broken

	session := cf.Cf("dev", "bosh", "env")
	Eventually(session).Should(gexec.Exit(0))

	// This test is more representative of how `bosh env` should be invoked
	w := gexec.NewPrefixedWriter("[bosh env] ", GinkgoWriter)
	boshEnv := func() *gexec.Session {
		boshCmd := exec.Command("/bin/sh",
			"-e",
			"-c", fmt.Sprintf(`eval "$(cf dev bosh env)" && bosh env`))

		session, err := gexec.Start(boshCmd, w, w)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit())
		return session
	}

	Eventually(boshEnv, time.Minute, 10*time.Second).Should(gexec.Exit(0))
}

func RemoveIPAliases(aliases ...string) {
	for _, alias := range aliases {
		cmd := exec.Command("sudo", "-n", "ifconfig", "lo0", "inet", alias+"/32", "remove")
		writer := gexec.NewPrefixedWriter("[ifconfig] ", GinkgoWriter)
		session, err := gexec.Start(cmd, writer, writer)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit())
	}
}

func downloadTestAsset(targetDir string, resourceUrl string) error {
	out, err := os.Create(filepath.Join(targetDir, "test-deps.dev"))
	if err != nil {
		return err
	}
	defer out.Close()

	resp, err := http.Get(resourceUrl)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func FullCleanup() {
	out, err := exec.Command("ps", "aux").Output()
	Expect(err).NotTo(HaveOccurred())
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "linuxkit") || strings.Contains(line, "hyperkit") || strings.Contains(line, "vpnkit") {
			cols := strings.Fields(line)
			pid, err := strconv.Atoi(cols[1])
			if err == nil && pid > 0 {
				syscall.Kill(pid, syscall.SIGKILL)
			}
		}
	}
	out, err = exec.Command("ps", "aux").Output()
	Expect(err).NotTo(HaveOccurred())
	for _, line := range strings.Split(string(out), "\n") {
		if strings.Contains(line, "linuxkit") || strings.Contains(line, "hyperkit") || strings.Contains(line, "vpnkit") {
			fmt.Printf("WARNING: one of the 'kits' processes are was still running: %s", line)
		}
	}
}

func PushAnApp() {
	server, port := fakeTcpServer()
	defer server.Close()

	tmpDir, _ := ioutil.TempDir("", "cf-test-app-")
	defer os.RemoveAll(tmpDir)
	Expect(ioutil.WriteFile(filepath.Join(tmpDir, "app"), []byte(`#!/usr/bin/env ruby
require 'webrick'
require 'open-uri'
require 'socket'
server = WEBrick::HTTPServer.new :Port => ENV['PORT']
server.mount_proc '/' do |request, response|
	response.body = 'Hello, world!'
end
server.mount_proc '/external' do |request, response|
	response.body = open('http://example.com').read
end
server.mount_proc '/host' do |request, response|
	# response.body = %x(host host.pcfdev.io)
	response.body = TCPSocket.new('host.pcfdev.io', `+strconv.Itoa(port)+`).gets
end
trap 'INT' do server.shutdown end
server.start
`), 0755)).To(Succeed())

	session := cf.Cf("push", "cf-test-app", "-p", tmpDir, "-b", "binary_buildpack", "-c", "./app")
	Eventually(session, 60).Should(gexec.Exit(0))

	Expect(httpGet("http://cf-test-app.v3.pcfdev.io")).To(Equal("Hello, world!"))
	Expect(httpGet("http://cf-test-app.v3.pcfdev.io/external")).To(ContainSubstring("Example Domain"))

	// TODO enable below once host.pcfdev.io works again
	// Expect(httpGet("http://cf-test-app.v3.pcfdev.io/host")).To(Equal("Text From Test Code"))
}

func fakeTcpServer() (net.Listener, int) {
	server, err := net.Listen("tcp", "0:0")
	Expect(err).NotTo(HaveOccurred())
	go func() {
		for {
			conn, err := server.Accept()
			if err == nil {
				conn.Write([]byte("Text From Test Code"))
				conn.Close()
			}
		}
	}()
	return server, server.Addr().(*net.TCPAddr).Port
}

func httpGet(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	b, err := ioutil.ReadAll(resp.Body)
	return string(b), err
}
