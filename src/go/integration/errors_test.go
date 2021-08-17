package integration_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testErrors(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		app *cutlass.App
	)

	it.After(func() {
		app = DestroyApp(t, app)
	})

	context("when specifying a non-packaged version of go", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "errors", "go_version"))
		})

		it("displays useful understandable errors", func() {
			Expect(app.Push()).To(HaveOccurred())
			Eventually(app.Stdout.String, 3*time.Second).Should(MatchRegexp("(?i)failed"))

			Expect(app.Stdout.String()).To(MatchRegexp("Unable to determine Go version to install: no match found for 99.99.99"))

			Expect(app.Stdout.String()).ToNot(MatchRegexp("Installing go99.99.99"))
			Expect(app.Stdout.String()).ToNot(MatchRegexp("Uploading droplet"))
		})
	})

	context("when a .godir file is detected", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "errors", "godir"))
		})

		it("errors with a deprecation message", func() {
			Expect(app.Push()).To(HaveOccurred())
			Eventually(app.Stdout.String, 3*time.Second).Should(MatchRegexp("(?i)failed"))

			Expect(app.Stdout.String()).To(ContainSubstring("Deprecated, .godir file found! Please update to supported Godep or Glide dependency managers."))
			Expect(app.Stdout.String()).To(ContainSubstring("See https://github.com/tools/godep or https://github.com/Masterminds/glide for usage information."))
		})
	})

	context("when using Godep and no vendor dir or Godeps/_workspace dir", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "errors", "unvendored_godep"))
		})

		it("fails with a helpful error message", func() {
			Expect(app.Push()).To(HaveOccurred())
			Eventually(app.Stdout.String, 3*time.Second).Should(MatchRegexp("(?i)failed"))

			Expect(app.Stdout.String()).To(MatchRegexp("vendor/ directory does not exist."))
		})
	})

	context("when the app is a single file, vendored, and no GOPACKAGENAME specified", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "errors", "no_gopackagename"))
		})

		it("fails with helpful error", func() {
			Expect(app.Push()).To(HaveOccurred())
			Eventually(app.Stdout.String, 3*time.Second).Should(MatchRegexp("(?i)failed"))

			Expect(app.Stdout.String()).To(MatchRegexp(`To use go native vendoring set the \$GOPACKAGENAME`))
		})
	})
}
