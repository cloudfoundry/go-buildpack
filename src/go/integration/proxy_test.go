package integration_test

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testProxy(t *testing.T, context spec.G, it spec.S) {
	AssertUsesProxyDuringStagingIfPresent(t, context, it, filepath.Join(settings.FixturesPath, "glide", "simple"))
	AssertUsesProxyDuringStagingIfPresent(t, context, it, filepath.Join(settings.FixturesPath, "glide", "vendored"))
	AssertUsesProxyDuringStagingIfPresent(t, context, it, filepath.Join(settings.FixturesPath, "godep", "vendored"))
}

func AssertUsesProxyDuringStagingIfPresent(t *testing.T, context spec.G, it spec.S, fixture string) {
	var Expect = NewWithT(t).Expect

	context("when an HTTP proxy is specified", func() {
		it("uses that proxy", func() {
			proxy, err := cutlass.NewProxy()
			Expect(err).NotTo(HaveOccurred())
			defer proxy.Close()

			root, err := cutlass.FindRoot()
			Expect(err).NotTo(HaveOccurred())

			bpFile := filepath.Join(root, settings.Buildpack.Version+"tmp")
			cmd := exec.Command("cp", settings.Buildpack.Path, bpFile)
			Expect(cmd.Run()).To(Succeed())
			defer os.Remove(bpFile)

			traffic, _, _, err := cutlass.InternetTraffic(fixture, bpFile, []string{
				"HTTP_PROXY=" + proxy.URL,
				"HTTPS_PROXY=" + proxy.URL,
			})
			Expect(err).NotTo(HaveOccurred())

			destURL, err := url.Parse(proxy.URL)
			Expect(err).NotTo(HaveOccurred())

			Expect(cutlass.UniqueDestination(
				traffic, fmt.Sprintf("%s.%s", destURL.Hostname(), destURL.Port()),
			)).To(Succeed())
		})
	})
}
