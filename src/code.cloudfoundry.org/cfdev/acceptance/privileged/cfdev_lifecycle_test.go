package privileged_test

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"

	"time"

	"io"
	"io/ioutil"
	"net/http"

	. "code.cloudfoundry.org/cfdev/acceptance"
	"code.cloudfoundry.org/garden/client"
	"code.cloudfoundry.org/garden/client/connection"
	"github.com/cloudfoundry-incubator/cf-test-helpers/cf"
	"github.com/onsi/gomega/gbytes"
)

var _ = Describe("cfdev lifecycle", func() {

	var (
		startSession *gexec.Session
	)
	BeforeEach(func() {
		pluginPath = os.Getenv("CFDEV_PLUGIN_PATH")
		if pluginPath == "" {
			Fail("please provide CFDEV_PLUGIN_PATH (use ./generate-plugin.sh)")
		}
		os.Unsetenv("BOSH_ALL_PROXY")

		cfdevHome = os.Getenv("CFDEV_HOME")
		if cfdevHome == "" {
			cfdevHome = filepath.Join(os.Getenv("HOME"), ".cfdev")
		}
		hyperkitPidPath = filepath.Join(cfdevHome, "state", "linuxkit", "hyperkit.pid")

		fmt.Println("PLUGIN PATH: " + pluginPath)
		session := cf.Cf("install-plugin", pluginPath, "-f")
		Eventually(session).Should(gexec.Exit(0))
	})

	AfterEach(func() {
		session := cf.Cf("uninstall-plugin", "cfdev")
		Eventually(session).Should(gexec.Exit(0))
	})

	Context("starting the default cf dev file", func() {
		BeforeEach(func() {
			isoPath := os.Getenv("ISO_PATH")
			if isoPath != "" {
				startSession = cf.Cf("dev", "start", "-f", isoPath, "-m", "8192")
			} else {
				startSession = cf.Cf("dev", "start")
			}
		})

		AfterEach(func() {
			if os.Getenv("CFDEV_FETCH_LOGS") == "true" {
				var logsSession *gexec.Session

				if dir := os.Getenv("CFDEV_LOG_DIR"); dir != "" {
					logsSession = cf.Cf("dev", "logs", "--dir", dir)
				} else {
					logsSession = cf.Cf("dev", "logs")
				}

				Eventually(logsSession).Should(gexec.Exit())
			}

			startSession.Terminate()
			Eventually(startSession).Should(gexec.Exit())

			By("deploy finished - stopping...")
			stopSession := cf.Cf("dev", "stop")
			Eventually(stopSession, 30*time.Second).Should(gexec.Exit(0))

			//ensure pid is not running
			if IsWindows() {
				Expect(doesVMExist()).To(BeFalse())
			} else {
				Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.linuxkit"), 5, 1).Should(BeFalse())

				hyperkitPid := PidFromFile(hyperkitPidPath)
				EventuallyProcessStops(hyperkitPid, 5)
			}

			Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.vpnkit"), 5, 1).Should(BeFalse())

			gexec.KillAndWait()
			RemoveIPAliases(BoshDirectorIP, CFRouterIP)
		})

		It("runs the entire vm lifecycle", func() {
			Eventually(startSession, 1*time.Hour).Should(gbytes.Say("Starting VPNKit"))

			Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.vpnkit"), 30, 1).Should(BeTrue())

			if !IsWindows() {
				Eventually(IsLaunchdRunning("org.cloudfoundry.cfdev.linuxkit"), 10, 1).Should(BeTrue())
			}

			By("waiting for garden to listen")
			client := client.New(connection.New("tcp", "localhost:8888"))
			Eventually(client.Ping, 360).Should(Succeed())

			EventuallyWeCanTargetTheBOSHDirector()

			By("waiting for cfdev cli to exit when the deploy finished")
			Eventually(startSession, 3*time.Hour).Should(gexec.Exit(0))

			By("waiting for cf router to listen")
			loginSession := cf.Cf("login", "-a", "https://api.dev.cfdev.sh", "--skip-ssl-validation", "-u", "admin", "-p", "admin", "-o", "cfdev-org", "-s", "cfdev-space")
			Eventually(loginSession).Should(gexec.Exit(0))

			By("pushing an app")
			PushAnApp()
		})
	})
})

func EventuallyWeCanTargetTheBOSHDirector() {
	By("waiting for bosh to listen")
	EventuallyShouldListenAt("https://"+BoshDirectorIP+":25555", 480)

	// Even though the test below is very similar this fails fast when `bosh env`
	// command is broken

	session := cf.Cf("dev", "bosh", "env")
	Eventually(session, 120, 1).Should(gexec.Exit(0))

	if !IsWindows() {
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
}

func RemoveIPAliases(aliases ...string) {
	if IsWindows() {
		return
	}

	for _, alias := range aliases {
		cmd := exec.Command("sudo", "-n", "ifconfig", "lo0", "inet", alias+"/32", "remove")
		writer := gexec.NewPrefixedWriter("[ifconfig] ", GinkgoWriter)
		session, err := gexec.Start(cmd, writer, writer)
		Expect(err).ToNot(HaveOccurred())
		Eventually(session).Should(gexec.Exit())
	}
}

func downloadTestAsset(targetDir string, resourceUrl string) error {
	out, err := os.Create(filepath.Join(targetDir, "test-deps.iso"))
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

func PushAnApp() {
	server, port := fakeTcpServer()
	defer server.Close()

	Eventually(cf.Cf("push", "cf-test-app", "--no-start", "-p", "./fixture", "-b", "ruby_buildpack"), 120).Should(gexec.Exit(0))
	Eventually(cf.Cf("set-env", "cf-test-app", "HOST_SERVER_PORT", strconv.Itoa(port)), 120).Should(gexec.Exit(0))
	Eventually(cf.Cf("create-service", "p-mysql", "10mb", "mydb"), 120).Should(gexec.Exit(0))
	Eventually(cf.Cf("bind-service", "cf-test-app", "mydb"), 120).Should(gexec.Exit(0))
	Eventually(cf.Cf("start", "cf-test-app"), 120).Should(gexec.Exit(0))

	Expect(httpGet("http://cf-test-app.dev.cfdev.sh")).To(Equal("Hello, world!"))
	Expect(httpGet("http://cf-test-app.dev.cfdev.sh/external")).To(ContainSubstring("Example Domain"))
	Expect(httpGet("http://cf-test-app.dev.cfdev.sh/host")).To(Equal("Text From Test Code"))
	Expect(httpGet("http://cf-test-app.dev.cfdev.sh/mysql")).To(ContainSubstring("innodb"))

	Eventually(cf.Cf("create-shared-domain", "tcp.dev.cfdev.sh", "--router-group", "default-tcp"), 10).Should(gexec.Exit(0))
	Eventually(cf.Cf("create-route", "cfdev-space", "tcp.dev.cfdev.sh", "--port", "1030"), 10).Should(gexec.Exit(0))
	Eventually(cf.Cf("map-route", "cf-test-app", "tcp.dev.cfdev.sh", "--port", "1030"), 10).Should(gexec.Exit(0))

	Eventually(func() (string, error) { return httpGet("http://tcp.dev.cfdev.sh:1030") }, 10).Should(Equal("Hello, world!"))
}

func fakeTcpServer() (net.Listener, int) {
	server, err := net.Listen("tcp", "localhost:0")
	Expect(err).NotTo(HaveOccurred())
	go func() {
		for {
			conn, err := server.Accept()
			if err != nil {
				continue
			}
			_, err = conn.Write([]byte("Text From Test Code"))
			Expect(err).NotTo(HaveOccurred())
			conn.Close()
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

func doesVMExist() bool {
	cmd := exec.Command("powershell.exe", "-Command", "(Get-VM -Name cfdev).name")
	output, err := cmd.Output()
	if err != nil {
		return false
	}

	return string(output) == "cfdev"
}
