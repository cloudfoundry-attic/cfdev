package cfanalytics_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestCfanalytics(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cfanalytics Suite")
}
