package integration_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testOffline(t *testing.T, context spec.G, it spec.S) {
	AssertNoInternetTraffic(t, context, it, filepath.Join(settings.FixturesPath, "default", "install_package_spec", "vendored"))
	AssertNoInternetTraffic(t, context, it, filepath.Join(settings.FixturesPath, "default", "ldflags"))
	AssertNoInternetTraffic(t, context, it, filepath.Join(settings.FixturesPath, "default", "no_procfile"))
	AssertNoInternetTraffic(t, context, it, filepath.Join(settings.FixturesPath, "default", "vendored"))
	AssertNoInternetTraffic(t, context, it, filepath.Join(settings.FixturesPath, "glide", "vendored"))
	AssertNoInternetTraffic(t, context, it, filepath.Join(settings.FixturesPath, "godep", "vendored"))
	AssertNoInternetTraffic(t, context, it, filepath.Join(settings.FixturesPath, "mod", "install_package_spec", "vendored"))
}

func AssertNoInternetTraffic(t *testing.T, context spec.G, it spec.S, fixture string) {
	var Expect = NewWithT(t).Expect

	context("when offline", func() {
		it("builds and runs the app", func() {
			root, err := cutlass.FindRoot()
			Expect(err).NotTo(HaveOccurred())

			bpFile := filepath.Join(root, settings.Buildpack.Version+"tmp")
			cmd := exec.Command("cp", settings.Buildpack.Path, bpFile)
			Expect(cmd.Run()).To(Succeed())
			defer os.Remove(bpFile)

			traffic, _, _, err := cutlass.InternetTraffic(fixture, bpFile, nil)
			Expect(err).NotTo(HaveOccurred())
			Expect(traffic).To(BeEmpty())
		})
	})
}
