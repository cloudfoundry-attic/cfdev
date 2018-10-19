package toggle_test

import (
	"code.cloudfoundry.org/cfdev/cfanalytics/toggle"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"io/ioutil"
	"os"
	"path/filepath"
)

var _ = Describe("Toggle", func() {
	var (
		tmpDir, saveFile string
	)
	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "analytics")
		Expect(err).ToNot(HaveOccurred())
		saveFile = filepath.Join(tmpDir, "somefile.txt")
	})
	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe("Analytics file exists", func() {
		Context("cf and custom are enabled", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`{"cfAnalyticsEnabled":true,"customAnalyticsEnabled":true,"props":{}}`), 0644)).To(Succeed())
			})

			It("returns enabled true and custom is true", func() {
				t := toggle.New(saveFile)

				Expect(t.CustomAnalyticsDefined()).To(BeTrue())
				Expect(t.Enabled()).To(BeTrue())
				Expect(t.IsCustom()).To(BeTrue())
			})
		})

		Context("cf and custom are disabled", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`{"cfAnalyticsEnabled":false,"customAnalyticsEnabled":false,"props":{}}`), 0644)).To(Succeed())
			})

			It("returns enabled false and custom is false", func() {
				t := toggle.New(saveFile)

				Expect(t.CustomAnalyticsDefined()).To(BeTrue())
				Expect(t.Enabled()).To(BeFalse())
				Expect(t.IsCustom()).To(BeFalse())
			})
		})

		Context("cf is enabled and custom is disabled", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`{"cfAnalyticsEnabled":true,"customAnalyticsEnabled":false,"props":{}}`), 0644)).To(Succeed())
			})

			It("returns enabled true and custom is false", func() {
				t := toggle.New(saveFile)

				Expect(t.CustomAnalyticsDefined()).To(BeFalse())
				Expect(t.Enabled()).To(BeTrue())
				Expect(t.IsCustom()).To(BeFalse())
			})
		})

		Context("cf is disabled and custom is enabled", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`{"cfAnalyticsEnabled":false,"customAnalyticsEnabled":true,"props":{}}`), 0644)).To(Succeed())
			})

			It("returns enabled true and custom is true", func() {
				t := toggle.New(saveFile)

				Expect(t.CustomAnalyticsDefined()).To(BeTrue())
				Expect(t.Enabled()).To(BeTrue())
				Expect(t.IsCustom()).To(BeTrue())
			})
		})

		Context("update customAnalyticsEnabled from false to true and save", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`{"cfAnalyticsEnabled":false,"customAnalyticsEnabled":false,"props":{}}`), 0644)).To(Succeed())
			})

			It("updates somefile.txt", func() {
				t := toggle.New(saveFile)
				Expect(t.SetCustomAnalyticsEnabled(true)).To(Succeed())
				Expect(t.IsCustom()).To(BeTrue())

				txt, err := ioutil.ReadFile(saveFile)
				Expect(err).ToNot(HaveOccurred())
				Expect(string(txt)).To(Equal(`{"cfAnalyticsEnabled":true,"customAnalyticsEnabled":true,"props":{}}`))
			})
		})

		Describe("Analytics file does NOT exist", func() {
			Context("and custom analytics are set to true", func() {
				It("returns enabled true and custom is true and defined is true", func() {
					t := toggle.New(saveFile)

					Expect(t.Defined()).To(BeFalse())
					Expect(t.CustomAnalyticsDefined()).To(BeFalse())
					Expect(t.SetCustomAnalyticsEnabled(true)).To(Succeed())
					Expect(t.Defined()).To(BeTrue())
					Expect(t.CustomAnalyticsDefined()).To(BeTrue())
					Expect(t.Enabled()).To(BeTrue())
					Expect(t.IsCustom()).To(BeTrue())
				})
			})
		})
		Describe("Set Prop", func() {
			var (
				t *toggle.Toggle
			)
			BeforeEach(func() {
				saveFile = filepath.Join(tmpDir, "somedir", "somefile.txt")
				t = toggle.New(saveFile)
			})
			It("sets props", func() {
				Expect(t.GetProps()).To(BeEquivalentTo(map[string]interface{}{}))
				Expect(t.SetProp("key", "value")).To(Succeed())
				Expect(t.GetProps()).To(BeEquivalentTo(map[string]interface{}{
					"key": "value",
				}))
				Expect(t.SetProp("other", "thing")).To(Succeed())
				Expect(t.GetProps()).To(BeEquivalentTo(map[string]interface{}{
					"key":   "value",
					"other": "thing",
				}))
			})
			It("writes props to file", func() {
				Expect(t.SetCFAnalyticsEnabled(false)).To(Succeed())
				Expect(t.SetProp("key", "value")).To(Succeed())
				Expect(t.SetProp("other", "thing")).To(Succeed())
				Expect(ioutil.ReadFile(saveFile)).To(MatchJSON(`{"cfAnalyticsEnabled":false,"customAnalyticsEnabled":false,"props": {"key": "value","other": "thing"}}`))
			})
		})
	})
})
