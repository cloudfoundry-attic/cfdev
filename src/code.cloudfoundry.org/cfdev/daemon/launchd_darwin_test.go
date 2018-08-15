package daemon_test

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/daemon"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/gexec"
	"time"
)

var _ = Describe("Launchd", func() {
	var plistDir string
	var plistPath string
	var label string
	var lnchd daemon.Launchd

	BeforeEach(func() {
		rand.Seed(time.Now().UTC().UnixNano())
		label = randomDaemonName()
		plistDir, _ = ioutil.TempDir("", "plist")
		lnchd = daemon.Launchd{
			PListDir: plistDir,
		}
		plistPath = filepath.Join(plistDir, label+".plist")
	})

	Describe("AddDaemon", func() {
		var binDir string

		BeforeEach(func() {
			binDir, _ = ioutil.TempDir("", "bin")
			ioutil.WriteFile(filepath.Join(binDir, "some-executable"), []byte(`some-content`), 0777)
			Eventually(loadedDaemons).ShouldNot(ContainSubstring(label))
		})

		AfterEach(func() {
			exec.Command("launchctl", "remove", label).Run()
			Eventually(loadedDaemons).ShouldNot(ContainSubstring(label))
			Expect(os.RemoveAll(plistDir)).To(Succeed())
			Expect(os.RemoveAll(binDir)).To(Succeed())
		})

		It("should write the plist and load the daemon", func() {
			executableToInstall := filepath.Join(binDir, "some-executable")
			spec := daemon.DaemonSpec{
				Label:            label,
				Program:          executableToInstall,
				SessionType:      "Background",
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
  <string>%s</string>
  <key>Program</key>
  <string>%s</string>
  <key>LimitLoadToSessionType</key>
  <string>Background</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
    <string>some-arg</string>
  </array>
  <key>RunAtLoad</key>
  <true/>
</dict>
</plist>
`, label, executableToInstall, executableToInstall)))
			plistFileInfo, err := os.Stat(plistPath)
			Expect(err).ToNot(HaveOccurred())
			Expect(plistFileInfo.Mode()).To(BeEquivalentTo(0644))

			Eventually(loadedDaemons).Should(ContainSubstring(label))
		})

		It("sets unix socket listeners on plist", func() {
			executableToInstall := filepath.Join(binDir, "some-executable")
			spec := daemon.DaemonSpec{
				Label:            label,
				Program:          executableToInstall,
				SessionType:      "Background",
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
  <string>%s</string>
  <key>Program</key>
  <string>%s</string>
  <key>LimitLoadToSessionType</key>
  <string>Background</string>
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
`, label, executableToInstall, executableToInstall)))
		})
	})

	Describe("RemoveDaemon", func() {
		var (
			binDir  string
			binPath string
		)

		BeforeEach(func() {
			binDir, _ = ioutil.TempDir("", "bin")
			binPath = filepath.Join(binDir, "some-bin-to-remove")
		})

		AfterEach(func() {
			Expect(os.RemoveAll(plistDir)).To(Succeed())
			Expect(os.RemoveAll(binDir)).To(Succeed())
		})

		Context("daemon is loaded and file exists", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(binPath, []byte("#!/bin bash echo hi"), 0700)).To(Succeed())
				Expect(ioutil.WriteFile(plistPath, []byte(fmt.Sprintf(`
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>%s</string>
  <key>Program</key>
  <string>%s</string>
  <key>LimitLoadToSessionType</key>
  <string>Background</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
  </array>
</dict>
</plist>`, label, binPath)), 0644)).To(Succeed())
				lnchd = daemon.Launchd{
					PListDir: plistDir,
				}

				cmd := exec.Command("launchctl", "load", "-S", "Background", "-F", plistPath)
				session, err := gexec.Start(cmd, GinkgoWriter, GinkgoWriter)
				Expect(err).NotTo(HaveOccurred())
				Eventually(session).Should(gexec.Exit(0))
				Eventually(loadedDaemons).Should(ContainSubstring(label))
			})

			It("should unload the daemon and remove the files", func() {
				Expect(lnchd.RemoveDaemon(label)).To(Succeed())
				Eventually(loadedDaemons).ShouldNot(ContainSubstring(label))
				Expect(plistPath).NotTo(BeAnExistingFile())
			})
		})

		Context("daemon is loaded and file does not exists", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(binPath, []byte("#!/bin bash echo hi"), 0700)).To(Succeed())
				Expect(ioutil.WriteFile(plistPath, []byte(fmt.Sprintf(`
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>%s</string>
  <key>Program</key>
  <string>%s</string>
  <key>LimitLoadToSessionType</key>
  <string>Background</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
  </array>
</dict>
</plist>`, label, binPath)), 0644)).To(Succeed())
				lnchd = daemon.Launchd{
					PListDir: plistDir,
				}
				Expect(exec.Command("launchctl", "load", "-S", "Background", "-F", plistPath).Run()).To(Succeed())
				Eventually(loadedDaemons).Should(ContainSubstring(label))
				Expect(os.RemoveAll(plistDir)).To(Succeed())
			})
			It("unloads the daemon", func() {
				Expect(lnchd.RemoveDaemon(label)).To(Succeed())
				Eventually(loadedDaemons).ShouldNot(ContainSubstring(label))
			})
		})

		Context("daemon is not loaded and file exists", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(binPath, []byte("#!/bin bash echo hi"), 0700)).To(Succeed())
				Expect(ioutil.WriteFile(plistPath, []byte(fmt.Sprintf(`
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
  <key>Label</key>
  <string>%s</string>
  <key>Program</key>
  <string>%s</string>
  <key>LimitLoadToSessionType</key>
  <string>Background</string>
  <key>ProgramArguments</key>
  <array>
    <string>%s</string>
  </array>
</dict>
</plist>`, label, binPath)), 0644)).To(Succeed())
				lnchd = daemon.Launchd{
					PListDir: plistDir,
				}
				Eventually(loadedDaemons).ShouldNot(ContainSubstring(label))
			})
			It("removes the file", func() {
				Expect(lnchd.RemoveDaemon(label)).To(Succeed())
				Expect(plistPath).NotTo(BeAnExistingFile())
			})
		})

		Context("daemon is not loaded and file does not exist", func() {
			It("succeeds", func() {
				Expect(lnchd.RemoveDaemon(label)).To(Succeed())
			})
		})
	})

	Describe("IsRunning", func() {
		var tmpDir string
		var lnchd daemon.Launchd
		BeforeEach(func() {
			tmpDir, _ = ioutil.TempDir("", "plist")
			lnchd = daemon.Launchd{
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
    <string>%s</string>
    <key>RunAtLoad</key>
    <false/>
    <key>Program</key>
    <string>%s/exe</string>
    <key>LimitLoadToSessionType</key>
    <string>Background</string>
    <key>ProgramArguments</key>
    <array>
      <string>%s/exe</string>
    </array>
  </dict>
  </plist>`, label, tmpDir, tmpDir)), 0644)).To(Succeed())
				lnchd = daemon.Launchd{
					PListDir: tmpDir,
				}
				Expect(exec.Command("launchctl", "load", "-S", "Background", "-F", filepath.Join(tmpDir, "some.plist")).Run()).To(Succeed())
				Eventually(loadedDaemons).Should(ContainSubstring(label))
			})
			AfterEach(func() {
				exec.Command("launchctl", "remove", label).Output()
			})
			Context("but not started", func() {
				It("returns false", func() {
					Expect(lnchd.IsRunning(label)).To(BeFalse())
				})
			})
			Context("and started", func() {
				BeforeEach(func() {
					Expect(exec.Command("launchctl", "start", label).Run()).To(Succeed())
				})
				It("returns true", func() {
					Expect(lnchd.IsRunning(label)).To(BeTrue())
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

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

func randomDaemonName() string {
	b := make([]rune, 10)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return "some-daemon" + string(b)
}
