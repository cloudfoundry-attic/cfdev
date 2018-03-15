package launchd_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/cfdevd/launchd"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
)

var _ = Describe("launchd", func() {
	Describe("AddDaemon", func() {
		var plistDir string
		var plistPath string
		var binDir string
		var lnchd launchd.Launchd

		BeforeEach(func() {
			plistDir, _ = ioutil.TempDir("", "plist")
			plistPath = filepath.Join(plistDir, "/org.some-org.some-daemon-name.plist")
			binDir, _ = ioutil.TempDir("", "bin")
			lnchd = launchd.Launchd{
				PListDir: plistDir,
			}
			ioutil.WriteFile(filepath.Join(binDir, "some-executable"), []byte(`some-content`), 0777)
			Expect(loadedDaemons()).ShouldNot(ContainSubstring("org.some-org.some-daemon-name"))
		})

		AfterEach(func() {
			exec.Command("launchctl", "unload", plistPath).Run()
			Expect(loadedDaemons()).ShouldNot(ContainSubstring("org.some-org.some-daemon-name"))
			Expect(os.RemoveAll(plistDir)).To(Succeed())
			Expect(os.RemoveAll(binDir)).To(Succeed())
		})

		It("should write the plist, install the binary, and load the daemon", func() {
			installationPath := filepath.Join(binDir, "org.some-org.some-daemon-executable")
			spec := launchd.DaemonSpec{
				Label:            "org.some-org.some-daemon-name",
				Program:          installationPath,
				ProgramArguments: []string{installationPath, "some-arg"},
				RunAtLoad:        true,
			}

			executableToInstall := filepath.Join(binDir, "some-executable")
			Expect(lnchd.AddDaemon(spec, executableToInstall)).To(Succeed())

			Expect(plistPath).To(BeAnExistingFile())
			plistFile, err := os.Open(plistPath)
			Expect(err).NotTo(HaveOccurred())
			plistData, err := ioutil.ReadAll(plistFile)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(plistData)).To(MatchXML(fmt.Sprintf(
				`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>org.some-org.some-daemon-name</string>
  <key>Program</key>
  <string>%s</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>some-arg</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
</dict>
</plist>
`, installationPath, installationPath)))
			plistFileInfo, err := plistFile.Stat()
			Expect(err).ToNot(HaveOccurred())
			var expectedPlistMode os.FileMode = 0644
			Expect(plistFileInfo.Mode()).To(Equal(expectedPlistMode))

			Expect(installationPath).To(BeAnExistingFile())
			installedBinary, err := os.Open(installationPath)
			Expect(err).NotTo(HaveOccurred())
			binFileInfo, err := installedBinary.Stat()
			var expectedBinMode os.FileMode = 0744
			Expect(binFileInfo.Mode()).To(Equal(expectedBinMode))
			contents, err := ioutil.ReadAll(installedBinary)
			Expect(string(contents)).To(Equal("some-content"))
			Expect(loadedDaemons()).Should(ContainSubstring("org.some-org.some-daemon-name"))
		})

		It("sets unix socket listeners on plist", func() {
			installationPath := filepath.Join(binDir, "org.some-org.some-daemon-executable")
			spec := launchd.DaemonSpec{
				Label:            "org.some-org.some-daemon-name",
				Program:          installationPath,
				ProgramArguments: []string{installationPath, "some-arg"},
				RunAtLoad:        true,
				Sockets: map[string]string{
					"CoolSocket": "/var/tmp/my.cool.socket",
				},
			}

			executableToInstall := filepath.Join(binDir, "some-executable")
			Expect(lnchd.AddDaemon(spec, executableToInstall)).To(Succeed())

			plistData, err := ioutil.ReadFile(plistPath)
			Expect(err).NotTo(HaveOccurred())
			Expect(string(plistData)).To(MatchXML(fmt.Sprintf(
				`<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>org.some-org.some-daemon-name</string>
  <key>Program</key>
  <string>%s</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>some-arg</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
  <key>Sockets</key>
  <dict>
    <key>CoolSocket</key>
    <dict>
      <key>SockPathMode</key>
      <integer>438</integer>
      <key>SockPathName</key>
      <string>/var/tmp/my.cool.socket</string>
    </dict>
  </dict>
</dict>
</plist>
`, installationPath, installationPath)))
		})
	})

	Describe("RemoveDaemon", func() {
		var (
			plistDir  string
			binDir    string
			plistPath string
			binPath   string
			lnchd     launchd.Launchd
		)

		BeforeEach(func() {
			plistDir, _ = ioutil.TempDir("", "plist")
			binDir, _ = ioutil.TempDir("", "bin")
			plistPath = filepath.Join(plistDir, "org.some-org.some-daemon-to-remove.plist")
			binPath = filepath.Join(binDir, "some-bin-to-remove")
			Expect(ioutil.WriteFile(binPath, []byte("#!/bin bash echo hi"), 0700)).To(Succeed())
			Expect(ioutil.WriteFile(plistPath, []byte(fmt.Sprintf(`
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>org.some-org.some-daemon-to-remove</string>
  <key>Program</key>
  <string>%s</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
  </array>
</dict>
</plist>`, binPath)), 0644)).To(Succeed())
			lnchd = launchd.Launchd{
				PListDir: plistDir,
			}
			Expect(exec.Command("launchctl", "load", plistPath).Run()).To(Succeed())
			Expect(loadedDaemons()).Should(ContainSubstring("org.some-org.some-daemon-to-remove"))
		})

		AfterEach(func() {
			Expect(os.RemoveAll(plistDir)).To(Succeed())
			Expect(os.RemoveAll(binDir)).To(Succeed())
		})

		It("should unload the daemon and remove the files", func() {
			spec := launchd.DaemonSpec{
				Label:            "org.some-org.some-daemon-to-remove",
				Program:          binPath,
				ProgramArguments: []string{binPath},
				RunAtLoad:        true,
			}

			Expect(lnchd.RemoveDaemon(spec)).To(Succeed())
			Expect(loadedDaemons()).ShouldNot(ContainSubstring("org.some-org.some-daemon-to-remove"))
			Expect(plistPath).NotTo(BeAnExistingFile())
			Expect(binPath).NotTo(BeAnExistingFile())
		})
	})
})

func loadedDaemons() string {
	session, err := gexec.Start(exec.Command("launchctl", "list"), GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
	return string(session.Out.Contents())
}
