package deploy_service

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestDeployService(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Cmd Deploy Service Suite")
}
