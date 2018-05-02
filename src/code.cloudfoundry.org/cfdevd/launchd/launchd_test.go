package launchd_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/cfdevd/launchd"
	"code.cloudfoundry.org/cfdevd/launchd/models"
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

		It("should write the plist and load the daemon", func() {
			executableToInstall := filepath.Join(binDir, "some-executable")
			spec := models.DaemonSpec{
				Label:            "org.some-org.some-daemon-name",
				Program:          executableToInstall,
				ProgramArguments: []string{executableToInstall, "some-arg"},
				RunAtLoad:        true,
			}

			Expect(lnchd.AddDaemon(spec)).To(Succeed())

			Expect(plistPath).To(BeAnExistingFile())
			Expect(ioutil.ReadFile(plistPath)).To(MatchXML(fmt.Sprintf(
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
`, executableToInstall, executableToInstall)))
			plistFileInfo, err := os.Stat(plistPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(plistFileInfo.Mode()).To(BeEquivalentTo(0644))

			Expect(loadedDaemons()).Should(ContainSubstring("org.some-org.some-daemon-name"))
		})

		It("sets unix socket listeners on plist", func() {
			executableToInstall := filepath.Join(binDir, "some-executable")
			spec := models.DaemonSpec{
				Label:            "org.some-org.some-daemon-name",
				Program:          executableToInstall,
				ProgramArguments: []string{executableToInstall, "some-arg"},
				RunAtLoad:        true,
				Sockets: map[string]string{
					"CoolSocket": "/var/tmp/my.cool.socket",
				},
			}

			Expect(lnchd.AddDaemon(spec)).To(Succeed())

			Expect(ioutil.ReadFile(plistPath)).To(MatchXML(fmt.Sprintf(
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
`, executableToInstall, executableToInstall)))
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
			Expect(lnchd.RemoveDaemon("org.some-org.some-daemon-to-remove")).To(Succeed())
			Expect(loadedDaemons()).ShouldNot(ContainSubstring("org.some-org.some-daemon-to-remove"))
			Expect(plistPath).NotTo(BeAnExistingFile())
		})
	})

	Describe("IsRunning", func() {
		var tmpDir string
		var lnchd launchd.Launchd
		BeforeEach(func() {
			tmpDir, _ = ioutil.TempDir("", "plist")
			lnchd = launchd.Launchd{
				PListDir: tmpDir,
			}
		})
		AfterEach(func() { Expect(os.RemoveAll(tmpDir)).To(Succeed()) })

		Context("label not loaded", func() {
			It("returns false", func() {
				Expect(lnchd.IsRunning("some-service-that-doesnt-exist")).To(BeFalse())
			})
		})

		Context("label has been loaded", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "exe"), []byte("#!/usr/bin/env bash\nsleep 60"), 0700)).To(Succeed())
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "some.plist"), []byte(fmt.Sprintf(`
	<?xml version="1.0" encoding="UTF-8"?>
	<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
	<plist version="1.0">
	<dict>
		<key>Label</key>
		<string>org.some-org.some-daemon</string>
		<key>RunAtLoad</key>
		<false/>
		<key>Program</key>
		<string>%s/exe</string>
		<key>ProgramArguments</key>
		<array>
			<string>%s/exe</string>
		</array>
	</dict>
	</plist>`, tmpDir, tmpDir)), 0644)).To(Succeed())
				lnchd = launchd.Launchd{
					PListDir: tmpDir,
				}
				Expect(exec.Command("launchctl", "load", filepath.Join(tmpDir, "some.plist")).Run()).To(Succeed())
				Expect(loadedDaemons()).Should(ContainSubstring("org.some-org.some-daemon"))
			})
			AfterEach(func() {
				exec.Command("launchctl", "remove", "org.some-org.some-daemon").Output()
			})
			Context("but not started", func() {
				It("returns false", func() {
					Expect(lnchd.IsRunning("org.some-org.some-daemon")).To(BeFalse())
				})
			})
			Context("and started", func() {
				BeforeEach(func() {
					Expect(exec.Command("launchctl", "start", "org.some-org.some-daemon").Run()).To(Succeed())
				})
				It("returns true", func() {
					Expect(lnchd.IsRunning("org.some-org.some-daemon")).To(BeTrue())
				})
			})
		})
	})
})

func loadedDaemons() string {
	session, err := gexec.Start(exec.Command("launchctl", "list"), GinkgoWriter, GinkgoWriter)
	Expect(err).NotTo(HaveOccurred())
	Eventually(session).Should(gexec.Exit(0))
	return string(session.Out.Contents())
}
