package brats_test

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/Masterminds/semver"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/bratshelper"
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = func() bool {
	testing.Init()
	return true
}()

func init() {
	flag.StringVar(&cutlass.DefaultMemory, "memory", "128M", "default memory for pushed apps")
	flag.StringVar(&cutlass.DefaultDisk, "disk", "256M", "default disk for pushed apps")
	flag.Parse()
}

var _ = SynchronizedBeforeSuite(func() []byte {
	// Run once
	return bratshelper.InitBpData(os.Getenv("CF_STACK"), true).Marshal()
}, func(data []byte) {
	// Run on all nodes
	bratshelper.Data.Unmarshal(data)

	Expect(cutlass.CopyCfHome()).To(Succeed())
	cutlass.SeedRandom()
	cutlass.DefaultStdoutStderr = GinkgoWriter
})

var _ = SynchronizedAfterSuite(func() {
	// Run on all nodes
}, func() {
	// Run once
	Expect(cutlass.DeleteOrphanedRoutes()).To(Succeed())
	Expect(cutlass.DeleteBuildpack(strings.Replace(bratshelper.Data.Cached, "_buildpack", "", 1))).To(Succeed())
	Expect(cutlass.DeleteBuildpack(strings.Replace(bratshelper.Data.Uncached, "_buildpack", "", 1))).To(Succeed())
	Expect(os.Remove(bratshelper.Data.CachedFile)).To(Succeed())
})

func TestBrats(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Brats Suite")
}

func CopyBrats(version string) *cutlass.App {
	dir, err := cutlass.CopyFixture(filepath.Join(bratshelper.Data.BpDir, "fixtures", "brats"))
	Expect(err).ToNot(HaveOccurred())

	if version == "" {
		manifest, err := libbuildpack.NewManifest(bratshelper.Data.BpDir, nil, time.Now())
		Expect(err).ToNot(HaveOccurred())
		dep, err := manifest.DefaultVersion("go")
		Expect(err).ToNot(HaveOccurred())
		version = dep.Version
	}

	data, err := ioutil.ReadFile(filepath.Join(dir, "Godeps", "Godeps.json"))
	Expect(err).ToNot(HaveOccurred())
	data = bytes.Replace(data, []byte("<%= version %>"), []byte(version), -1)
	Expect(ioutil.WriteFile(filepath.Join(dir, "Godeps", "Godeps.json"), data, 0644)).To(Succeed())

	return cutlass.New(dir)
}

func GetOldestVersion(dep, bpDir string) string {
	manifest, err := libbuildpack.NewManifest(bpDir, nil, time.Now())
	Expect(err).ToNot(HaveOccurred())
	deps := manifest.AllDependencyVersions(dep)

	if len(deps) == 0 {
		Fail(fmt.Sprintf("No dependencies found in manifest.yml for %s", dep))
	}

	sort.Slice(deps, func(i, j int) bool {
		s1, err := semver.NewVersion(deps[i])
		Expect(err).NotTo(HaveOccurred())
		s2, err := semver.NewVersion(deps[i])
		Expect(err).NotTo(HaveOccurred())
		return s1.LessThan(s2)
	})

	return deps[0]
}

func PushApp(app *cutlass.App) {
	Expect(app.Push()).To(Succeed())
	Eventually(app.InstanceStates, 20*time.Second).Should(Equal([]string{"RUNNING"}))
}
