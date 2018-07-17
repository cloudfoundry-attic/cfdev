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
		tmpDir, saveFile string
	)
	BeforeEach(func() {
		var err error
		tmpDir, err = ioutil.TempDir("", "analytics.")
		Expect(err).ToNot(HaveOccurred())
		saveFile = filepath.Join(tmpDir, "somefile.txt")
	})
	AfterEach(func() {
		os.RemoveAll(tmpDir)
	})

	Describe("Defined", func() {
		Context("save file exists with json and key enabled", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`{"enabled":null}`), 0644)).To(Succeed())
			})
			It("returns true", func() {
				t := toggle.New(saveFile)
				Expect(t.Defined()).To(BeTrue())
			})
		})
		Context("save file exists with json but NOT key enabled", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`{"something":null}`), 0644)).To(Succeed())
			})
			It("returns false", func() {
				t := toggle.New(saveFile)
				Expect(t.Defined()).To(BeFalse())
			})
		})
		Context("save file exists with deprecatedTrueVal", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`optin`), 0644)).To(Succeed())
			})
			It("returns true", func() {
				t := toggle.New(saveFile)
				Expect(t.Defined()).To(BeTrue())
			})
		})
		Context("save file exists with deprecatedFalseVal", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`optout`), 0644)).To(Succeed())
			})
			It("returns true", func() {
				t := toggle.New(saveFile)
				Expect(t.Defined()).To(BeTrue())
			})
		})
		Context("save file does NOT exist", func() {
			It("returns false", func() {
				t := toggle.New(saveFile)
				Expect(t.Defined()).To(BeFalse())
			})
		})
	})

	Describe("Get", func() {
		Context("save file exists with json and key enabled == true", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`{"enabled":true}`), 0644)).To(Succeed())
			})
			It("returns true", func() {
				t := toggle.New(saveFile)
				Expect(t.Get()).To(BeTrue())
			})
		})
		Context("save file exists with json and key enabled == false", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`{"enabled":false}`), 0644)).To(Succeed())
			})
			It("returns false", func() {
				t := toggle.New(saveFile)
				Expect(t.Get()).To(BeFalse())
			})
		})
		Context("save file exists with json and key enabled == 'a non bool'", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`{"enabled":"something"}`), 0644)).To(Succeed())
			})
			It("returns false", func() {
				t := toggle.New(saveFile)
				Expect(t.Get()).To(BeFalse())
			})
		})
		Context("save file exists with deprecatedTrueVal", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`optin`), 0644)).To(Succeed())
			})
			It("returns true", func() {
				t := toggle.New(saveFile)
				Expect(t.Get()).To(BeTrue())
			})
		})
		Context("save file exists with deprecatedFalseVal", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`optout`), 0644)).To(Succeed())
			})
			It("returns false", func() {
				t := toggle.New(saveFile)
				Expect(t.Get()).To(BeFalse())
			})
		})
		Context("save file does NOT exist", func() {
			It("returns false", func() {
				t := toggle.New(saveFile)
				Expect(t.Get()).To(BeFalse())
			})
		})
	})

	Describe("GetProps", func() {
		Context("file missing", func() {
			It("is empty", func() {
				t := toggle.New(saveFile)
				Expect(t.GetProps()).To(BeEmpty())
			})
		})
		Context("file has props", func() {
			BeforeEach(func() {
				Expect(ioutil.WriteFile(saveFile, []byte(`{"props":{"key":"value","other":"thing"}}`), 0644)).To(Succeed())
			})
			It("returns data from file", func() {
				t := toggle.New(saveFile)
				Expect(t.GetProps()).To(BeEquivalentTo(map[string]interface{}{
					"key":   "value",
					"other": "thing",
				}))
			})
		})
	})

	Describe("Set", func() {
		var (
			t *toggle.Toggle
		)
		BeforeEach(func() {
			saveFile = filepath.Join(tmpDir, "somedir", "somefile.txt")
			t = toggle.New(saveFile)
		})
		It("sets val for get", func() {
			Expect(t.Set(true)).To(Succeed())
			Expect(t.Get()).To(BeTrue())

			Expect(t.Set(false)).To(Succeed())
			Expect(t.Get()).To(BeFalse())
		})
		It("writes trueVal to file", func() {
			Expect(t.Set(true)).To(Succeed())
			Expect(ioutil.ReadFile(saveFile)).To(MatchJSON(`{
				"enabled": true,
				"props": {}
			}`))
		})
		It("writes falseVal to file", func() {
			Expect(t.Set(false)).To(Succeed())
			Expect(ioutil.ReadFile(saveFile)).To(MatchJSON(`{
				"enabled": false,
				"props": {}
			}`))
		})
	})

	Describe("SetProp", func() {
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
			Expect(t.Set(false)).To(Succeed())
			Expect(t.SetProp("key", "value")).To(Succeed())
			Expect(t.SetProp("other", "thing")).To(Succeed())

			Expect(ioutil.ReadFile(saveFile)).To(MatchJSON(`{
				"enabled": false,
				"props": {
					"key": "value",
					"other": "thing"
				}
			}`))
		})
		Context("optin was not defined", func() {
			It("stores prop but not enabled flag", func() {
				Expect(t.SetProp("key", "value")).To(Succeed())
				Expect(ioutil.ReadFile(saveFile)).To(MatchJSON(`{
					"props": {
						"key": "value"
					}
				}`))
			})
		})
	})
})
