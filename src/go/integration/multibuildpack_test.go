package integration_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testMultiBuildpack(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		context("when supplied with dotnet-core", func() {
			it("finds the supplied dependency in the runtime container", func() {
				deployment, logs, err := platform.Deploy.
					WithBuildpacks(
						"https://github.com/cloudfoundry/dotnet-core-buildpack#master",
						"go_buildpack",
					).
					WithEnv(map[string]string{
						"GOPACKAGENAME": "whatever",
					}).
					Execute(name, filepath.Join(fixtures, "multibuildpack", "dotnet"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Supplying Dotnet Core")))
				Eventually(deployment).Should(Serve(MatchRegexp(`dotnet: \d+\.\d+\.\d+`)))
			})
		})

		context("when supplied with nodejs", func() {
			it("finds the supplied dependency in the runtime container", func() {
				deployment, logs, err := platform.Deploy.
					WithBuildpacks(
						"https://github.com/cloudfoundry/nodejs-buildpack#master",
						"go_buildpack",
					).
					WithEnv(map[string]string{
						"GOPACKAGENAME": "go-online",
					}).
					Execute(name, filepath.Join(fixtures, "multibuildpack", "nodejs"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Nodejs Buildpack version")))
				Eventually(deployment).Should(Serve(ContainSubstring("INFO hello world")))
			})
		})

		context("when supplied with python", func() {
			it("an app is pushed which uses pip dependencies", func() {
				deployment, logs, err := platform.Deploy.
					WithBuildpacks(
						"https://github.com/cloudfoundry/python-buildpack#master",
						"go_buildpack",
					).
					WithEnv(map[string]string{
						"GOPACKAGENAME": "go-online",
					}).
					Execute(name, filepath.Join(fixtures, "multibuildpack", "python"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Installing python")))
				Eventually(deployment).Should(Serve(ContainSubstring(`[{"hello":"world"}]`)))
			})

			it("an app is pushed which uses miniconda", func() {
				deployment, logs, err := platform.Deploy.
					WithBuildpacks(
						"https://github.com/cloudfoundry/python-buildpack#master",
						"go_buildpack",
					).
					WithEnv(map[string]string{
						"GOPACKAGENAME": "go-online",
					}).
					Execute(name, filepath.Join(fixtures, "multibuildpack", "miniconda"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Installing Miniconda")))
				Eventually(deployment).Should(Serve(ContainSubstring(`[{"hello":"world"}]`)))
			})

			it("an app is pushed which uses NLTK corpus", func() {
				deployment, logs, err := platform.Deploy.
					WithBuildpacks(
						"https://github.com/cloudfoundry/python-buildpack#master",
						"go_buildpack",
					).
					WithEnv(map[string]string{
						"GOPACKAGENAME": "go-online",
					}).
					Execute(name, filepath.Join(fixtures, "multibuildpack", "nltk"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Downloading NLTK corpora...")))
				Eventually(deployment).Should(Serve(ContainSubstring("The Fulton County Grand Jury said Friday an investigation of Atlanta's recent primary election produced")))
			})
		})

		context("when supplied with ruby", func() {
			it("finds the supplied dependency in the runtime container", func() {
				deploymentProcess := platform.Deploy.
					WithBuildpacks(
						"https://github.com/cloudfoundry/ruby-buildpack#master",
						"go_buildpack",
					).
					WithEnv(map[string]string{
						"GOPACKAGENAME": "go-online",
					})
				// TODO: remove this once ruby-buildpack runs on cflinuxfs4
				// This is done to have the sample app written in ruby up and running
				if os.Getenv("CF_STACK") == "cflinuxfs4" {
					deploymentProcess = deploymentProcess.WithStack("cflinuxfs3")
				}

				deployment, logs, err := deploymentProcess.Execute(name, filepath.Join(fixtures, "multibuildpack", "ruby"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(MatchRegexp("Installing ruby \\d+\\.\\d+\\.\\d+")))
				Expect(logs).To(ContainLines(ContainSubstring("Go Buildpack version")))
				Eventually(deployment).Should(Serve(MatchRegexp("Bundler version \\d+\\.\\d+\\.\\d+")))
			})
		})
	}
}
