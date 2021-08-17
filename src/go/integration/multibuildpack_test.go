package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testMultiBuildpack(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		app *cutlass.App
	)

	it.After(func() {
		app = DestroyApp(t, app)
	})

	context("when supplied with dotnet-core", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "multibuildpack", "dotnet"))
			app.Buildpacks = []string{
				"https://github.com/cloudfoundry/dotnet-core-buildpack#master",
				"go_buildpack",
			}
			app.SetEnv("GOPACKAGENAME", "whatever")
		})

		it("finds the supplied dependency in the runtime container", func() {
			PushAppAndConfirm(t, app)

			Expect(app.Stdout.String()).To(ContainSubstring("Supplying Dotnet Core"))
			Expect(app.GetBody("/")).To(MatchRegexp(`dotnet: \d+\.\d+\.\d+`))
		})
	})

	context("when supplied with nodejs", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "multibuildpack", "nodejs"))
			app.Buildpacks = []string{
				"https://github.com/cloudfoundry/nodejs-buildpack#master",
				"go_buildpack",
			}
		})

		it("finds the supplied dependency in the runtime container", func() {
			PushAppAndConfirm(t, app)

			Expect(app.Stdout.String()).To(ContainSubstring("Nodejs Buildpack version"))
			Expect(app.GetBody("/")).To(MatchRegexp("INFO hello world"))
		})
	})

	context("when supplied with python", func() {
		pushApp := func(t *testing.T, fixture string) {
			app = cutlass.New(fixture)
			app.Buildpacks = []string{
				"https://github.com/cloudfoundry/python-buildpack#master",
				"go_buildpack",
			}
			app.Disk = "1G"
			PushAppAndConfirm(t, app)
		}

		it("an app is pushed which uses pip dependencies", func() {
			pushApp(t, filepath.Join(settings.FixturesPath, "multibuildpack", "python"))

			Expect(app.Stdout.String()).To(ContainSubstring("Installing python"))
			Expect(app.GetBody("/")).To(ContainSubstring(`[{"hello":"world"}]`))
		})

		it("an app is pushed which uses miniconda", func() {
			pushApp(t, filepath.Join(settings.FixturesPath, "multibuildpack", "miniconda"))

			Expect(app.Stdout.String()).To(ContainSubstring("Installing Miniconda"))
			Expect(app.GetBody("/")).To(ContainSubstring(`[{"hello":"world"}]`))
		})

		it("an app is pushed which uses NLTK corpus", func() {
			pushApp(t, filepath.Join(settings.FixturesPath, "multibuildpack", "nltk"))

			Expect(app.Stdout.String()).To(ContainSubstring("Downloading NLTK corpora..."))
			Expect(app.GetBody("/")).To(ContainSubstring("The Fulton County Grand Jury said Friday an investigation of Atlanta's recent primary election produced"))
		})
	})

	context("when supplied with ruby", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "multibuildpack", "ruby"))
			app.Buildpacks = []string{
				"https://github.com/cloudfoundry/ruby-buildpack#master",
				"go_buildpack",
			}
		})

		it("finds the supplied dependency in the runtime container", func() {
			PushAppAndConfirm(t, app)

			Expect(app.Stdout.String()).To(MatchRegexp("Installing ruby \\d+\\.\\d+\\.\\d+"))
			Expect(app.Stdout.String()).To(ContainSubstring("Go Buildpack version"))
			Expect(app.GetBody("/")).To(MatchRegexp("Bundler version \\d+\\.\\d+\\.\\d+"))
		})
	})
}
