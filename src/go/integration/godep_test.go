package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testGodep(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		it("builds app with Godep", func() {
			deployment, _, err := platform.Deploy.
				Execute(name, filepath.Join(fixtures, "godep", "vendored"))
			Expect(err).NotTo(HaveOccurred())

			Eventually(deployment).Should(Serve(ContainSubstring("hello, world")))
		})

		context("with a wildcard go version", func() {
			it("uses the default version", func() {
				deployment, logs, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "godep", "wildcard_go_version"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(MatchRegexp(`Installing go 1\.\d+(\.\d+)?`)))
				Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
			})
		})

		context("with a Godeps/_workspace dir", func() {
			it("builds the app", func() {
				deployment, logs, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "godep", "simple"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Running: godep go install -tags cloudfoundry -buildmode pie .")))
				Eventually(deployment).Should(Serve(ContainSubstring("hello, world")))
			})
		})
	}
}
