package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testGlide(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		it("builds app with Glide", func() {
			deployment, _, err := platform.Deploy.
				Execute(name, filepath.Join(fixtures, "glide", "simple"))
			Expect(err).NotTo(HaveOccurred())

			Eventually(deployment).Should(Serve(ContainSubstring("hello, world")))
		})

		context("when the dependencies are vendored", func() {
			it("builds app with Glide", func() {
				deployment, logs, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "glide", "vendored"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Note: skipping (glide install) due to non-empty vendor directory.")))

				Eventually(deployment).Should(Serve(ContainSubstring("hello, world")))
			})
		})
	}
}
