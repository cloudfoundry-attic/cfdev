package toggle_test

import (
	"io/ioutil"
	"os"
	"path/filepath"

	"code.cloudfoundry.org/cfdev/cfanalytics/toggle"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Toggle", func() {
	var (
		tmpDir string
	)
	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "analytics.")
		Expect(err).ToNot(HaveOccurred())
	})
	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe("Defined", func() {
		Context("save file exists", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "somefile.txt"), []byte("something"), 0644)).To(Succeed())
			})
			It("returns true", func() {
				t := toggle.New(filepath.Join(tmpDir, "somefile.txt"), "true", "false")
				Expect(t.Defined()).To(BeTrue())
			})
		})
		Context("save file does NOT exist", func() {
			It("returns true", func() {
				t := toggle.New(filepath.Join(tmpDir, "somefile.txt"), "true", "false")
				Expect(t.Defined()).To(BeFalse())
			})
		})
	})

	Describe("Get", func() {
		Context("save file does NOT exist", func() {
			It("returns false", func() {
				t := toggle.New(filepath.Join(tmpDir, "somefile.txt"), "true", "false")
				Expect(t.Get()).To(BeFalse())
			})
		})
		Context("save file exists with value equal to trueVal", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "somefile.txt"), []byte("istrue"), 0644)).To(Succeed())
			})
			It("returns true", func() {
				t := toggle.New(filepath.Join(tmpDir, "somefile.txt"), "istrue", "isfalse")
				Expect(t.Get()).To(BeTrue())
			})
		})
		Context("save file exists with value equal to falseVal", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(filepath.Join(tmpDir, "somefile.txt"), []byte("isfalse"), 0644)).To(Succeed())
			})
			It("returns true", func() {
				t := toggle.New(filepath.Join(tmpDir, "somefile.txt"), "istrue", "isfalse")
				Expect(t.Get()).To(BeFalse())
			})
		})
	})

	Describe("Set", func() {
		var (
			t        *toggle.Toggle
			savePath string
		)
		BeforeEach(func() {
			savePath = filepath.Join(tmpDir, "somedir", "somefile.txt")
			t = toggle.New(savePath, "isTrue", "isFalse")
		})
		It("sets val for get", func() {
			Expect(t.Set(true)).To(Succeed())
			Expect(t.Get()).To(BeTrue())

			Expect(t.Set(false)).To(Succeed())
			Expect(t.Get()).To(BeFalse())
		})
		It("writes trueVal to file", func() {
			Expect(t.Set(true)).To(Succeed())
			Expect(ioutil.ReadFile(savePath)).To(Equal([]byte("isTrue")))
		})
		It("writes falseVal to file", func() {
			Expect(t.Set(false)).To(Succeed())
			Expect(ioutil.ReadFile(savePath)).To(Equal([]byte("isFalse")))
		})
	})
})
