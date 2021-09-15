package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testDep(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		it("builds app with dep", func() {
			deployment, _, err := platform.Deploy.
				WithEnv(map[string]string{
					"GOPACKAGENAME": "simple",
				}).
				Execute(name, filepath.Join(fixtures, "dep", "simple"))
			Expect(err).NotTo(HaveOccurred())

			Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
		})

		context("when there is no lockfile", func() {
			it("successfully stages", func() {
				deployment, _, err := platform.Deploy.
					WithEnv(map[string]string{
						"GOPACKAGENAME": "no-lockfile",
					}).
					Execute(name, filepath.Join(fixtures, "dep", "no_lockfile"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
			})
		})

		context("when the dependencies are vendored", func() {
			it("successfully stages", func() {
				deployment, _, err := platform.Deploy.
					WithEnv(map[string]string{
						"GOPACKAGENAME": "vendored",
					}).
					Execute(name, filepath.Join(fixtures, "dep", "vendored"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
			})
		})
	}
}
