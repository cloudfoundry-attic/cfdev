package telemetry_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestTelemetry(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Telemetry Suite")
}
