package sanity

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestContext(t *testing.T) {

	RegisterFailHandler(Fail)

	suiteConfig, _ := GinkgoConfiguration()

	RunSpecs(t, "Test on sanity", suiteConfig)
}

var _ = Describe("CSI Driver", func() {
	Context("NFS Sanity Test", testNfsSanity)
})
