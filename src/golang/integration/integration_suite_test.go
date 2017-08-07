package integration_test

import (
	"encoding/json"
	"flag"
	"fmt"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"time"

	"github.com/blang/semver"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/cloudfoundry/libbuildpack/packager"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

var bpDir string
var buildpackVersion string
var packagedBuildpack cutlass.VersionedBuildpackPackage

func init() {
	flag.StringVar(&buildpackVersion, "version", "", "version to use (builds if empty)")
	flag.BoolVar(&cutlass.Cached, "cached", true, "cached buildpack")
	flag.StringVar(&cutlass.DefaultMemory, "memory", "128M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "384M", "default disk for pushed apps")
	flag.Parse()
}

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

func PushAppAndConfirm(app *cutlass.App) {
	Expect(app.Push()).To(Succeed())
	Eventually(func() ([]string, error) { return app.InstanceStates() }, 20*time.Second).Should(Equal([]string{"RUNNING"}))
	Expect(app.ConfirmBuildpack(buildpackVersion)).To(Succeed())
}

func Restart(app *cutlass.App) {
	Expect(app.Restart()).To(Succeed())
	Eventually(func() ([]string, error) { return app.InstanceStates() }, 20*time.Second).Should(Equal([]string{"RUNNING"}))
}

func AssertNoInternetTraffic(fixtureName string) {
	It("has no traffic", func() {
		if !cutlass.Cached {
			Skip("Running uncached tests")
		}

		localVersion := fmt.Sprintf("%s.%s", buildpackVersion, cutlass.RandStringRunes(10))
		bpFile, err := packager.Package(bpDir, packager.CacheDir, localVersion, cutlass.Cached)
		Expect(err).To(BeNil())
		defer os.Remove(bpFile)

		traffic, err := cutlass.InternetTraffic(
			bpDir,
			filepath.Join("fixtures", fixtureName),
			bpFile,
			[]string{},
		)
		Expect(err).To(BeNil())
		Expect(traffic).To(BeEmpty())
	})
}

func ApiHasTask() bool {
	apiVersionString, err := cutlass.ApiVersion()
	Expect(err).To(BeNil())
	apiVersion, err := semver.Make(apiVersionString)
	Expect(err).To(BeNil())
	apiHasTask, err := semver.ParseRange(">= 2.75.0")
	Expect(err).To(BeNil())
	return apiHasTask(apiVersion)
}

var _ = SynchronizedBeforeSuite(func() []byte {
	if buildpackVersion == "" {
		packagedBuildpack, err := cutlass.PackageUniquelyVersionedBuildpack()
		Expect(err).NotTo(HaveOccurred())

		data, err := json.Marshal(packagedBuildpack)
		Expect(err).NotTo(HaveOccurred())
		return data
	}

	return []byte{}
}, func(data []byte) {
	var err error
	if len(data) > 0 {
		err = json.Unmarshal(data, &packagedBuildpack)
		Expect(err).NotTo(HaveOccurred())
		buildpackVersion = packagedBuildpack.Version
	}

	bpDir, err = cutlass.FindRoot()
	Expect(err).NotTo(HaveOccurred())

	cutlass.SeedRandom()
	cutlass.DefaultStdoutStderr = GinkgoWriter
})

var _ = SynchronizedAfterSuite(func() {
	// Run on all nodes
}, func() {
	Expect(cutlass.RemovePackagedBuildpack(packagedBuildpack)).To(Succeed())
	Expect(cutlass.DeleteOrphanedRoutes()).To(Succeed())
})

func AssertUsesProxyDuringStagingIfPresent(fixtureName string) func() {
	return func() {
		Context("with an uncached buildpack", func() {
			var bpFile string
			var proxy *httptest.Server
			BeforeEach(func() {
				var err error
				if cutlass.Cached {
					Skip("Running cached tests")
				}

				localVersion := fmt.Sprintf("%s.%s", buildpackVersion, time.Now().Format("20060102150405"))
				bpFile, err = packager.Package(bpDir, packager.CacheDir, localVersion, cutlass.Cached)
				Expect(err).To(BeNil())

				proxy, err = cutlass.NewProxy()
				Expect(err).To(BeNil())
			})
			AfterEach(func() {
				os.Remove(bpFile)
				proxy.Close()
			})

			It("uses a proxy during staging if present", func() {
				traffic, err := cutlass.InternetTraffic(
					bpDir,
					filepath.Join("fixtures", fixtureName),
					bpFile,
					[]string{"HTTP_PROXY=" + proxy.URL, "HTTPS_PROXY=" + proxy.URL},
				)
				Expect(err).To(BeNil())

				destUrl, err := url.Parse(proxy.URL)
				Expect(err).To(BeNil())

				Expect(cutlass.UniqueDestination(
					traffic, fmt.Sprintf("%s.%s", destUrl.Hostname(), destUrl.Port()),
				)).To(BeNil())
			})
		})
	}
}
