package acceptance

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"syscall"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

const (
	GardenIP       = "localhost"
	BoshDirectorIP = "10.144.0.4"
	CFRouterIP     = "10.144.0.34"
)

func SetupDependencies(cacheDir string) {
	gopaths := strings.Split(os.Getenv("GOPATH"), ":")

	assets := []string{
		"cfdev-efi.iso",
		"cf-deps.iso",
		"vpnkit",
		"hyperkit",
		"linuxkit",
		"UEFI.fd",
		"qcow-tool",
	}

	if runtime.GOOS == "windows" {
		assets = []string{
			"cfdev-efi.iso",
			"cf-deps.iso",
			"vpnkit.exe",
			"winsw.exe",
		}
	}

	err := os.MkdirAll(cacheDir, 0777)
	Expect(err).ToNot(HaveOccurred())

	for _, asset := range assets {
		target := filepath.Join(cacheDir, asset)

		goPath := gopaths[0]
		if runtime.GOOS == "windows" {
			goPath = os.Getenv("GOPATH")
		}

		for _, origin := range []string{filepath.Join(goPath, "output", asset), filepath.Join(goPath, "linuxkit", asset), filepath.Join(GetCfdevHome(), "cache", asset)} {

			if exists, _ := FileExists(origin); exists {
				Expect(os.Symlink(origin, target)).To(Succeed())
				break
			}
		}
		Expect(target).To(BeAnExistingFile())
	}
}

func HttpServerIsListeningAt(url string) error {
	client := http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}

	resp, err := client.Get(url)

	if resp != nil {
		resp.Body.Close()
	}

	return err
}

func EventuallyProcessStops(pid int, timeoutSec int) {
	EventuallyWithOffset(1, func() (bool, error) {
		return processIsRunning(pid)
	}, timeoutSec).Should(BeFalse())
}

func processIsRunning(pid int) (bool, error) {
	proc, err := os.FindProcess(pid)
	if err != nil {
		return false, err
	}

	if err := proc.Signal(syscall.Signal(0)); err != nil {
		return false, nil
	}

	return true, nil
}

func IsLaunchdRunning(label string) func() (bool, error) {
	return func() (bool, error) {
		if runtime.GOOS == "darwin" {
			txt, err := exec.Command("launchctl", "list", label).CombinedOutput()
			if err != nil {
				if strings.Contains(string(txt), "Could not find service") {
					return false, nil
				}
				return false, err
			}
			re := regexp.MustCompile(`^\s*"PID"\s*=`)
			for _, line := range strings.Split(string(txt), "\n") {
				if re.MatchString(line) {
					return true, nil
				}
			}
			return false, nil
		} else {
			cmd := exec.Command("powershell.exe", "-Command", fmt.Sprintf("Get-Service | Where-Object {$_.Name -eq \"%s\"}", label))
			output, err := cmd.Output()
			if err != nil {
				return false, err
			}

			if strings.Contains(string(output), label) {
				return true, nil
			}

			return false, nil
		}
	}
}

func PidFromFile(pidFile string) int {
	pidBytes, _ := ioutil.ReadFile(pidFile)
	pid, _ := strconv.ParseInt(string(pidBytes), 10, 64)
	return int(pid)
}

func FileExists(file string) (bool, error) {
	_, err := os.Stat(file)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}

		return false, err
	}

	return true, nil
}

func GetCfdevHome() string {
	cfdevHome := os.Getenv("CFDEV_HOME")
	if cfdevHome != "" {
		return cfdevHome
	}

	if runtime.GOOS == "windows" {
		return filepath.Join(os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"), ".cfdev")
	} else {
		return filepath.Join(os.Getenv("HOME"), ".cfdev")
	}
}

func GetCfPluginPath() string {
	if runtime.GOOS == "windows" {
		return "cf"
	} else {
		return "/usr/local/bin/cf"
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

func IsWindows() bool {
	return runtime.GOOS == "windows"
}

func HasSudoPrivilege() bool {
	if IsWindows() {
		return true
	}

	cmd := exec.Command("sh", "-c", "sudo -n true")
	session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit())

	if session.ExitCode() == 0 {
		return true
	}
	return false
}
