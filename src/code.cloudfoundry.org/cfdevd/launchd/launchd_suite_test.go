package launchd

import (
	"math/rand"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestLaunchd(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Launchd Suite")
}

var _ = BeforeSuite(func() {
	rand.Seed(time.Now().UnixNano())
})
