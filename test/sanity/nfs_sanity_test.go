package sanity

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/chenliu1993/simple-csi-driver/internal/nfs"
	"gopkg.in/yaml.v2"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/kubernetes-csi/csi-test/v5/pkg/sanity"
)

var (
	nodeName  = "test-sanity-node"
	nfsdriver = "nfsplugin.csi.cliufreever.com"

	baseDir = "nfsshare"
	// subDir   string
	endpoint = "unix:///tmp/csi.sock"

	testNfsServer        = "127.0.0.1:2049"
	testMountPermissions = "0644"

	testVolumeParametersFile = "test_nfsvolume_parameters.yaml"

	serverKey          = "server"
	basedirKey         = "basedir"
	mountPermissionKey = "mountPermission"
	config             sanity.TestConfig
	stopCh             chan os.Signal
)

func testNfsSanity() {
	BeforeEach(func() {
		var err error

		stopCh = make(chan os.Signal, 1)

		// For testing Nfs plugin
		paramsFileContent := map[string]string{
			serverKey:          testNfsServer,
			basedirKey:         baseDir,
			mountPermissionKey: testMountPermissions,
		}

		yamlContent, err := yaml.Marshal(paramsFileContent)
		Expect(err).NotTo(HaveOccurred())

		err = os.WriteFile(testVolumeParametersFile, yamlContent, 0644)
		Expect(err).NotTo(HaveOccurred())

		config = sanity.NewTestConfig()
		config.TestVolumeParametersFile = testVolumeParametersFile
		config.Address = endpoint

		signal.Notify(stopCh, syscall.SIGTERM)
		go func() {
			nfsDriver := nfs.NewNFSDriver(nfsdriver, endpoint, nodeName, stopCh)
			nfsDriver.Run()
		}()
		time.Sleep(500 * time.Millisecond)
	})
	AfterEach(func() {
		// Clean up
		// err := os.RemoveAll(baseDir)
		// Expect(err).NotTo(HaveOccurred())

		// err = os.RemoveAll(subDir)
		// Expect(err).NotTo(HaveOccurred())

		err := os.RemoveAll(endpoint)
		Expect(err).NotTo(HaveOccurred())

		err = os.Remove(testVolumeParametersFile)
		Expect(err).NotTo(HaveOccurred())

		err = os.RemoveAll(config.TargetPath)
		Expect(err).NotTo(HaveOccurred())

		err = os.RemoveAll(config.StagingPath)
		Expect(err).NotTo(HaveOccurred())

		// stop the drivers
		stopCh <- syscall.SIGTERM
	})

	Context("NFS Sanity Test", func() {
		// Do the default sanity test
		sanity.GinkgoTest(&config)
	})
}
