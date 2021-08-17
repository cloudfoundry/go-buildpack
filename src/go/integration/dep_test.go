package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDep(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		app *cutlass.App
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "dep", "simple"))
	})

	it.After(func() {
		app = DestroyApp(t, app)
	})

	it("builds app with dep", func() {
		PushAppAndConfirm(t, app)

		Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
	})

	context("when there is no lockfile", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "dep", "no_lockfile"))
		})

		it("successfully stages", func() {
			PushAppAndConfirm(t, app)

			Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
		})
	})

	context("when the dependencies are vendored", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "dep", "vendored"))
		})

		it("successfully stages", func() {
			PushAppAndConfirm(t, app)

			Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
		})
	})
}
