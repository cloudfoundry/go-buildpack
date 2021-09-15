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

func testDefault(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
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

		it("builds and runs the app", func() {
			deployment, logs, err := platform.Deploy.
				Execute(name, filepath.Join(fixtures, "default", "simple"))
			Expect(err).NotTo(HaveOccurred())

			Expect(logs).To(ContainLines(MatchRegexp(`Installing go [\d\.]+`)))
			Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
		})

		context("when BP_DEBUG is enabled", func() {
			it("staging output includes before/after compile hooks", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{"BP_DEBUG": "1"}).
					Execute(name, filepath.Join(fixtures, "default", "simple"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(MatchRegexp("HOOKS 1: BeforeCompile")))
				Expect(logs).To(ContainLines(MatchRegexp("HOOKS 2: AfterCompile")))

				Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
			})
		})

		context("when the app is a single file and GOPACKAGENAME specified", func() {
			it("successfully stages", func() {
				deployment, _, err := platform.Deploy.
					WithEnv(map[string]string{
						"GOPACKAGENAME": "go-single-file",
					}).
					Execute(name, filepath.Join(fixtures, "default", "single_file"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("simple apps are good")))
			})
		})

		context("when given a custom package spec", func() {
			it("installs the custom package", func() {
				deployment, logs, err := platform.Deploy.
					WithEnv(map[string]string{
						"GOPACKAGENAME":           "example.com/install_pkg_spec",
						"GO_INSTALL_PACKAGE_SPEC": "example.com/install_pkg_spec/app",
					}).
					Execute(name, filepath.Join(fixtures, "default", "install_package_spec", "simple"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(ContainSubstring("Running: go install -tags cloudfoundry -buildmode pie example.com/install_pkg_spec/app")))
				Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
			})
		})

		context("when the packagename is the same as a bash builtin or on path", func() {
			it("sets the start command to run this app", func() {
				deployment, _, err := platform.Deploy.
					WithEnv(map[string]string{
						"GOPACKAGENAME": "test",
					}).
					Execute(name, filepath.Join(fixtures, "default", "builtin"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("foo:")))
			})
		})

		context("app has vendored dependencies", func() {
			it("builds and runs the app", func() {
				deployment, _, err := platform.Deploy.
					WithEnv(map[string]string{
						"GOPACKAGENAME": "example.com/user/go-online",
					}).
					Execute(name, filepath.Join(fixtures, "default", "vendored"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("Read: a.A == 1")))
			})

			context("when given a custom package spec", func() {
				it("builds and runs the app", func() {
					deployment, _, err := platform.Deploy.
						WithEnv(map[string]string{
							"GOPACKAGENAME":           "go-online",
							"GO_INSTALL_PACKAGE_SPEC": "./app",
						}).
						Execute(name, filepath.Join(fixtures, "default", "install_package_spec", "vendored"))
					Expect(err).NotTo(HaveOccurred())

					Eventually(deployment).Should(Serve(ContainSubstring("Read: a.A == 1")))
				})
			})
		})

		context("when app has no Procfile", func() {
			it("builds and runs the app", func() {
				deployment, logs, err := platform.Deploy.
					Execute(name, filepath.Join(fixtures, "default", "no_procfile"))
				Expect(err).NotTo(HaveOccurred())

				Expect(logs).To(ContainLines(MatchRegexp(`Installing go [\d\.]+`)))
				Eventually(deployment).Should(Serve(ContainSubstring("go, world")))
			})
		})

		context("when the app specifies LDFLAGS", func() {
			it("links correctly", func() {
				deployment, _, err := platform.Deploy.
					WithEnv(map[string]string{
						"GO_LINKER_SYMBOL": "main.linker_flag",
						"GO_LINKER_VALUE":  "flag_linked",
					}).
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

		context("when the app contains a symlink to a directory", func() {
			it("sets the start command to run this app", func() {
				deployment, _, err := platform.Deploy.
					WithEnv(map[string]string{
						"GOPACKAGENAME": "symlink_dir",
					}).
					Execute(name, filepath.Join(fixtures, "default", "symlink_dir"))
				Expect(err).NotTo(HaveOccurred())

				Eventually(deployment).Should(Serve(ContainSubstring("simple apps are good")))
			})
		})
	}
}
