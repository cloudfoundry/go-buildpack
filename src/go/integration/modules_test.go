package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testModules(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		context("when the version is obtained from the go.mod file", func() {
			context("when the go.mod file contains a go version", func() {
				it("builds the app with modules", func() {
					deployment, logs, err := platform.Deploy.
						Execute(name, filepath.Join(fixtures, "mod", "simple"))
					Expect(err).NotTo(HaveOccurred())

					Expect(logs).To(ContainLines(ContainSubstring("Go version found in go.mod")))
					Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
				})
			})

			context("when the go.mod file contains a non supported go version", func() {
				it("builds the app with modules", func() {
					deployment, logs, err := platform.Deploy.
						Execute(name, filepath.Join(fixtures, "mod", "simple_with_old_gomod_go_version"))
					Expect(err).NotTo(HaveOccurred())

					Expect(logs).To(ContainLines(ContainSubstring("Go version found in go.mod")))
					Expect(logs).To(ContainLines(ContainSubstring("Go version found in go.mod not supported by the Buildpack.")))
					Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
				})
			})

			context("when the go.mod file does not contains a go version", func() {
				it("builds the app with modules", func() {
					deployment, logs, err := platform.Deploy.
						Execute(name, filepath.Join(fixtures, "mod", "simple_without_version_in_gomod"))
					Expect(err).NotTo(HaveOccurred())

					Expect(logs).To(ContainLines(ContainSubstring("No Go version found in go.mod")))
					Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
				})
			})
		})

		context("when the version is obtained from the GOVERSION env variable", func() {
			it("builds the app with modules", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"GOVERSION": "go1.18",
					}).
					Execute(name, filepath.Join(fixtures, "mod", "simple"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Using $GOVERSION override.")))
				Expect(logs).To(ContainLines(ContainSubstring("Installing go 1.18")))
				Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
			})
		})

		context("when given a custom package spec", func() {
			it("installs the custom package using go modules", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"GO_INSTALL_PACKAGE_SPEC": "github.com/full/path/cmd/app",
						"GOVERSION":               "go1.17",
					}).
					Execute(name, filepath.Join(fixtures, "mod", "install_package_spec", "absolute"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Running: go install -tags cloudfoundry -buildmode pie github.com/full/path/cmd/app")))
				Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
			})

			context("when using relative paths", func() {
				it("installs the custom package using go modules and relative paths", func() {
					deployment, logs, err := platform.Deploy.
						WithEnv(map[string]string{
							"GO_INSTALL_PACKAGE_SPEC": "./cmd/app",
							"GOVERSION":               "go1.17",
						}).
						Execute(name, filepath.Join(fixtures, "mod", "install_package_spec", "relative"))
					Expect(err).NotTo(HaveOccurred())

					Expect(logs).To(ContainLines(ContainSubstring("Running: go install -tags cloudfoundry -buildmode pie ./cmd/app")))
					Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
				})
			})
		})

		context("when the modules are vendored", func() {
			it("builds the app with modules", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"GOPACKAGENAME": "go-online",
						"GOVERSION":     "go1.17",
					}).
					Execute(name, filepath.Join(fixtures, "mod", "vendored"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).NotTo(ContainLines(ContainSubstring("go: downloading github.com/BurntSushi/toml")))
				Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
			})

			context("when given a custom package spec", func() {
				it("installs the custom package using vendored go modules", func() {
					deployment, logs, err := platform.Deploy.
						WithEnv(map[string]string{
							"GO_INSTALL_PACKAGE_SPEC": "github.com/full/path/cmd/app",
							"GOVERSION":               "go1.17",
						}).
						Execute(name, filepath.Join(fixtures, "mod", "install_package_spec", "vendored"))
					Expect(err).NotTo(HaveOccurred())

					Expect(logs).To(ContainLines(ContainSubstring("Running: go install -tags cloudfoundry -buildmode pie github.com/full/path/cmd/app")))
					Expect(logs).NotTo(ContainLines(ContainSubstring("go: downloading github.com/deckarep")))
					Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
				})
			})
		})
	}
}
