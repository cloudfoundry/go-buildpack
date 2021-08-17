package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testDefault(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		app *cutlass.App
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "default", "simple"))
	})

	it.After(func() {
		app = DestroyApp(t, app)
	})

	it("builds and runs the app", func() {
		PushAppAndConfirm(t, app)

		Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
		Expect(app.Stdout.String()).To(MatchRegexp(`Installing go [\d\.]+`))
	})

	context("when BP_DEBUG is enabled", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "default", "simple"))
			app.SetEnv("BP_DEBUG", "1")
		})

		it("staging output includes before/after compile hooks", func() {
			PushAppAndConfirm(t, app)

			Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
			Expect(app.Stdout.String()).To(MatchRegexp("HOOKS 1: BeforeCompile"))
			Expect(app.Stdout.String()).To(MatchRegexp("HOOKS 2: AfterCompile"))
		})
	})

	context("when the app is a single file and GOPACKAGENAME specified", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "default", "single_file"))
		})

		it("successfully stages", func() {
			PushAppAndConfirm(t, app)

			Expect(app.GetBody("/")).To(ContainSubstring("simple apps are good"))
		})
	})

	context("when given a custom package spec", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "default", "install_package_spec", "simple"))
		})

		it("installs the custom package", func() {
			PushAppAndConfirm(t, app)

			Expect(app.Stdout.String()).To(ContainSubstring("Running: go install -tags cloudfoundry -buildmode pie example.com/install_pkg_spec/app"))
			Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
		})
	})

	context("when the packagename is the same as a bash builtin or on path", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "default", "builtin"))
		})

		it("sets the start command to run this app", func() {
			PushAppAndConfirm(t, app)

			Expect(app.GetBody("/")).To(ContainSubstring("foo:"))
		})
	})

	context("app has vendored dependencies", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "default", "vendored"))
		})

		it("builds and runs the app", func() {
			PushAppAndConfirm(t, app)

			Expect(app.Stdout.String()).To(MatchRegexp("Init: a.A == 1"))
			Expect(app.GetBody("/")).To(ContainSubstring("Read: a.A == 1"))
		})

		context("when given a custom package spec", func() {
			it.Before(func() {
				app = cutlass.New(filepath.Join(settings.FixturesPath, "default", "install_package_spec", "vendored"))
			})

			it("builds and runs the app", func() {
				PushAppAndConfirm(t, app)

				Expect(app.Stdout.String()).To(MatchRegexp("Init: a.A == 1"))
				Expect(app.GetBody("/")).To(ContainSubstring("Read: a.A == 1"))
			})
		})
	})

	context("when app has no Procfile", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "default", "no_procfile"))
		})

		it("builds and runs the pap", func() {
			PushAppAndConfirm(t, app)

			Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
			Expect(app.Stdout.String()).To(MatchRegexp(`Installing go [\d\.]+`))
		})
	})

	context("when the app specifies LDFLAGS", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "default", "ldflags"))
		})

		it("links correctly", func() {
			PushAppAndConfirm(t, app)

			body, err := app.GetBody("/")
			Expect(err).NotTo(HaveOccurred())

			Expect(body).To(ContainSubstring("linker_flag=flag_linked"))
			Expect(body).To(ContainSubstring("other_linker_flag=other_flag_linked"))
			Expect(body).NotTo(ContainSubstring("flag_linked_should_not_appear"))
		})
	})

	context("when the app contains a symlink to a directory", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "default", "symlink_dir"))
		})

		it("sets the start command to run this app", func() {
			PushAppAndConfirm(t, app)

			Expect(app.GetBody("/")).To(ContainSubstring("simple apps are good"))
		})
	})
}
