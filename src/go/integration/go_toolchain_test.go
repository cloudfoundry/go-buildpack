package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testGoToolchain(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect     = NewWithT(t).Expect
			Eventually = NewWithT(t).Eventually

			name string
		)

		it.Before(func() {
			var err error
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(platform.Delete.Execute(name)).To(Succeed())
		})

		context("when GO_INSTALL_TOOLS_IN_IMAGE is specified", func() {
			it("keeps the go toolchain in the droplet", func() {
				deployment, _, err := platform.Deploy.
					WithEnv(map[string]string{
						"GO_INSTALL_TOOLS_IN_IMAGE": "true",
					}).
					Execute(name, filepath.Join(fixtures, "go_toolchain", "toolchain"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(MatchRegexp(`go version go1\.\d+(\.\d+)? linux/amd64`)))
			})
		})
	}
}
