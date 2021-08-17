package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testModules(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		app *cutlass.App
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "mod", "simple"))
	})

	it.After(func() {
		app = DestroyApp(t, app)
	})

	it("builds the app with modules", func() {
		PushAppAndConfirm(t, app)

		Expect(app.Stdout.String()).To(MatchRegexp("go: downloading github.com/BurntSushi/toml"))
		Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
	})

	context("when given a custom package spec", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "mod", "install_package_spec", "absolute"))
		})

		it("installs the custom package using go modules", func() {
			PushAppAndConfirm(t, app)

			Expect(app.Stdout.String()).To(ContainSubstring("Running: go install -tags cloudfoundry -buildmode pie github.com/full/path/cmd/app"))
			Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
		})

		context("when using relative paths", func() {
			it.Before(func() {
				app = cutlass.New(filepath.Join(settings.FixturesPath, "mod", "install_package_spec", "relative"))
			})

			it("installs the custom package using go modules and relative paths", func() {
				PushAppAndConfirm(t, app)

				Expect(app.Stdout.String()).To(ContainSubstring("Running: go install -tags cloudfoundry -buildmode pie ./cmd/app"))
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
			})
		})
	})

	context("when the modules are vendored", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "mod", "vendored"))
		})

		it("builds the app with modules", func() {
			PushAppAndConfirm(t, app)

			Expect(app.Stdout.String()).NotTo(MatchRegexp("go: downloading github.com/BurntSushi/toml"))
			Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
		})

		context("when given a custom package spec", func() {
			it.Before(func() {
				app = cutlass.New(filepath.Join(settings.FixturesPath, "mod", "install_package_spec", "vendored"))
			})

			it("installs the custom package using vendored go modules", func() {
				PushAppAndConfirm(t, app)

				Expect(app.Stdout.String()).To(ContainSubstring("Running: go install -tags cloudfoundry -buildmode pie github.com/full/path/cmd/app"))
				Expect(app.Stdout.String()).NotTo(MatchRegexp("go: downloading github.com/deckarep"))
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
			})
		})
	})
}
