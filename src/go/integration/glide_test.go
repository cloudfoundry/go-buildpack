package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testGlide(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		app *cutlass.App
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "glide", "simple"))
	})

	it.After(func() {
		app = DestroyApp(t, app)
	})

	it("builds app with Glide", func() {
		PushAppAndConfirm(t, app)

		Expect(app.Stdout.String()).To(ContainSubstring("Hello from foo!"))
		Expect(app.GetBody("/")).To(ContainSubstring("hello, world"))
	})

	context("when the dependencies are vendored", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "glide", "vendored"))
		})

		it("builds app with Glide", func() {
			PushAppAndConfirm(t, app)

			Expect(app.GetBody("/")).To(ContainSubstring("hello, world"))
			Expect(app.Stdout.String()).To(ContainSubstring("Hello from foo!"))
			Expect(app.Stdout.String()).To(ContainSubstring("Note: skipping (glide install) due to non-empty vendor directory."))
		})
	})
}
