package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testErrors(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect = NewWithT(t).Expect

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

		context("when specifying a non-packaged version of go", func() {
			it("displays useful understandable errors", func() {
				_, logs, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "errors", "go_version"))
				Expect(err).To(MatchError(ContainSubstring("App staging failed")))

				Expect(logs).To(ContainLines(MatchRegexp("Unable to determine Go version to install: no match found for 99.99.99")))
				Expect(logs).ToNot(ContainLines(MatchRegexp("Installing go99.99.99")))
				Expect(logs).ToNot(ContainLines(MatchRegexp("Uploading droplet")))
			})
		})

		context("when a .godir file is detected", func() {
			it("errors with a deprecation message", func() {
				_, logs, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "errors", "godir"))
				Expect(err).To(MatchError(ContainSubstring("App staging failed")))

				Expect(logs).To(ContainLines(
					ContainSubstring("Deprecated, .godir file found! Please update to supported Godep or Glide dependency managers."),
					ContainSubstring("See https://github.com/tools/godep or https://github.com/Masterminds/glide for usage information."),
				))
			})
		})

		context("when using Godep and no vendor dir or Godeps/_workspace dir", func() {
			it("fails with a helpful error message", func() {
				_, logs, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "errors", "unvendored_godep"))
				Expect(err).To(MatchError(ContainSubstring("App staging failed")))

				Expect(logs).To(ContainLines(MatchRegexp("vendor/ directory does not exist.")))
			})
		})

		context("when the app is a single file, vendored, and no GOPACKAGENAME specified", func() {
			it("fails with helpful error", func() {
				_, logs, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "errors", "no_gopackagename"))
				Expect(err).To(MatchError(ContainSubstring("App staging failed")))

				Expect(logs).To(ContainLines(MatchRegexp(`To use go native vendoring set the \$GOPACKAGENAME`)))
			})
		})
	}
}
