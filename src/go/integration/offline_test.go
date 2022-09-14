package integration_test

import (
	"io"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testOffline(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		context("when deploying a simple vendored app", func() {
			it("builds and runs the app", func() {
				deployment, _, err := platform.Deploy.
					WithEnv(map[string]string{
						"GOPACKAGENAME": "example.com/user/go-online",
					}).
					WithoutInternetAccess().
					Execute(name, filepath.Join(fixtures, "default", "vendored"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("Read: a.A == 1")))
			})

			context("given a GO_INSTALL_PACKAGE_SPEC", func() {
				it("builds and runs the app", func() {
					deployment, _, err := platform.Deploy.
						WithEnv(map[string]string{
							"GOPACKAGENAME":           "go-online",
							"GO_INSTALL_PACKAGE_SPEC": "./app",
						}).
						WithoutInternetAccess().
						Execute(name, filepath.Join(fixtures, "default", "install_package_spec", "vendored"))
					Expect(err).NotTo(HaveOccurred())

					Eventually(deployment).Should(Serve(ContainSubstring("Read: a.A == 1")))
				})
			})
		})

		context("when specifying LDFLAGS", func() {
			it("links correctly", func() {
				deployment, _, err := platform.Deploy.
					WithEnv(map[string]string{
						"GO_LINKER_SYMBOL": "main.linker_flag",
						"GO_LINKER_VALUE":  "flag_linked",
					}).
					WithoutInternetAccess().
					Execute(name, filepath.Join(fixtures, "default", "ldflags"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("hello world")))

				response, err := http.Get(deployment.ExternalURL)
				Expect(err).NotTo(HaveOccurred())
				defer response.Body.Close()

				body, err := io.ReadAll(response.Body)
				Expect(err).NotTo(HaveOccurred())

				Expect(string(body)).To(ContainSubstring("linker_flag=flag_linked"))
				Expect(string(body)).To(ContainSubstring("other_linker_flag=other_flag_linked"))
				Expect(string(body)).NotTo(ContainSubstring("flag_linked_should_not_appear"))
			})
		})

		context("without a Procfile", func() {
			it("builds and runs the app", func() {
				deployment, logs, err := platform.Deploy.
					WithoutInternetAccess().
					Execute(name, filepath.Join(fixtures, "default", "no_procfile"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(MatchRegexp(`Installing go [\d\.]+`)))
				Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
			})
		})

		context("when using glide", func() {
			it("builds app with Glide", func() {
				deployment, logs, err := platform.Deploy.
					WithoutInternetAccess().
					Execute(name, filepath.Join(fixtures, "glide", "vendored"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Note: skipping (glide install) due to non-empty vendor directory.")))

				Eventually(deployment).Should(Serve(ContainSubstring("hello, world")))
			})
		})

		context("when using godep", func() {
			it("builds app with Godep", func() {
				deployment, _, err := platform.Deploy.
					WithoutInternetAccess().
					Execute(name, filepath.Join(fixtures, "godep", "vendored"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("hello, world")))
			})
		})

		context("when using modules", func() {
			it("installs the custom package using vendored go modules", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"GO_INSTALL_PACKAGE_SPEC": "github.com/full/path/cmd/app",
						"GOVERSION":               "go1.18",
					}).
					WithoutInternetAccess().
					Execute(name, filepath.Join(fixtures, "mod", "install_package_spec", "vendored"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Running: go install -tags cloudfoundry -buildmode pie github.com/full/path/cmd/app")))
				Expect(logs).NotTo(ContainLines(ContainSubstring("go: downloading github.com/deckarep")))
				Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
			})
		})
	}
}
