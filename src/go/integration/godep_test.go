package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testGodep(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		app *cutlass.App
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "godep", "vendored"))
	})

	it.After(func() {
		app = DestroyApp(t, app)
	})

	it("builds app with Godep", func() {
		PushAppAndConfirm(t, app)

		Expect(app.Stdout.String()).To(MatchRegexp("Hello from foo!"))
		Expect(app.GetBody("/")).To(ContainSubstring("hello, world"))
	})

	context("with a wildcard go version", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "godep", "wildcard_go_version"))
		})

		it("uses the default version", func() {
			PushAppAndConfirm(t, app)

			Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
			Expect(app.Stdout.String()).To(MatchRegexp(`Installing go 1\.\d+(\.\d+)?`))
		})
	})

	context("with a Godeps/_workspace dir", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "godep", "simple"))
		})

		it("builds the app", func() {
			PushAppAndConfirm(t, app)

			Expect(app.Stdout.String()).To(MatchRegexp("Hello from foo!"))
			Expect(app.GetBody("/")).To(ContainSubstring("hello, world"))
		})
	})
}
