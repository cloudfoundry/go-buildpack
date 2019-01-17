package integration_test

import (
	"path/filepath"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("CF Go Buildpack", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	Context("without cached buildpack dependencies", func() {
		BeforeEach(func() {
			if cutlass.Cached {
				Skip("but running cached tests")
			}
		})

		Context("app uses glide", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "with_glide", "src", "with_glide"))
			})

			It("", func() {
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(ContainSubstring("Hello from foo!"))
				Expect(app.GetBody("/")).To(ContainSubstring("hello, world"))
			})
			AssertUsesProxyDuringStagingIfPresent("with_glide/src/with_glide")
		})

		Context("app has dependencies", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "with_dependencies", "src", "with_dependencies"))
			})

			It("", func() {
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(MatchRegexp("Hello from foo!"))
				Expect(app.GetBody("/")).To(ContainSubstring("hello, world"))
			})

			AssertUsesProxyDuringStagingIfPresent("with_dependencies/src/with_dependencies")
		})

		Context("app has no dependencies", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "go_app"))
			})

			It("", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
				Expect(app.Stdout.String()).To(MatchRegexp(`Installing go [\d\.]+`))
				Expect(app.Stdout.String()).To(MatchRegexp(`Download \[https:\/\/.*\]`))
			})
		})

		Context("app has go modules and go version > 1.11", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "go_mod_app"))
			})

			It("", func() {
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(MatchRegexp("Installing go 1.11.4"))
				Expect(app.Stdout.String()).To(MatchRegexp("go: downloading github.com/BurntSushi/toml"))
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
			})
		})

		Context("app has go modules and modules are vendored", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "go_mod_vendored_app"))
			})

			It("", func() {
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(MatchRegexp("Installing go 1.11.4"))
				Expect(app.Stdout.String()).NotTo(MatchRegexp("go: downloading github.com/BurntSushi/toml"))
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
			})
		})

		Context("expects a non-packaged version of go", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "go99"))
			})

			It("displays useful understandable errors", func() {
				Expect(app.Push()).To(HaveOccurred())
				Eventually(app.Stdout.String, 3*time.Second).Should(MatchRegexp("(?i)failed"))

				Expect(app.Stdout.String()).To(MatchRegexp("Unable to determine Go version to install: no match found for 99.99.99"))

				Expect(app.Stdout.String()).ToNot(MatchRegexp("Installing go99.99.99"))
				Expect(app.Stdout.String()).ToNot(MatchRegexp("Uploading droplet"))
			})
		})

		Context("heroku example", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "heroku_example"))
			})

			It("", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("hello, heroku"))
			})
			AssertUsesProxyDuringStagingIfPresent("heroku_example")
		})

		Context("app uses glide and has vendored dependencies", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "glide_and_vendoring", "src", "glide_and_vendoring"))
			})

			It("", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("hello, world"))
				Expect(app.Stdout.String()).To(ContainSubstring("Hello from foo!"))
				Expect(app.Stdout.String()).To(ContainSubstring("Note: skipping (glide install) due to non-empty vendor directory."))
			})

			AssertUsesProxyDuringStagingIfPresent("glide_and_vendoring/src/glide_and_vendoring")
		})

		Context("a .godir file is detected", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "deprecated_heroku", "src", "deprecated_heroku"))
			})

			It("fails with a deprecation message", func() {
				Expect(app.Push()).To(HaveOccurred())
				Eventually(app.Stdout.String, 3*time.Second).Should(MatchRegexp("(?i)failed"))

				Expect(app.Stdout.String()).To(ContainSubstring("Deprecated, .godir file found! Please update to supported Godep or Glide dependency managers."))
				Expect(app.Stdout.String()).To(ContainSubstring("See https://github.com/tools/godep or https://github.com/Masterminds/glide for usage information."))
			})
		})

		Context("a go app with wildcard matcher", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "wildcard_go_version", "src", "go_app"))
			})

			It("fails with a deprecation message", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
				Expect(app.Stdout.String()).To(MatchRegexp(`Installing go 1\.\d+\.\d+`))
			})
		})

		Context("a go app with an invalid wildcard matcher", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "invalid_wildcard_version", "src", "go_app"))
			})

			It("fails with a helpful warning message", func() {
				Expect(app.Push()).To(HaveOccurred())
				Eventually(app.Stdout.String, 3*time.Second).Should(MatchRegexp("(?i)failed"))

				Expect(app.Stdout.String()).To(MatchRegexp(`Unable to determine Go version to install: no match found for 1.3.x`))
				Expect(app.Stdout.String()).ToNot(MatchRegexp(`Installing go1.3`))
			})
		})

		Context("a go app with a custom package spec", func() {
			It("installs the custom package", func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "install_pkg_spec"))
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(ContainSubstring("Running: go install -tags cloudfoundry -buildmode pie example.com/install_pkg_spec/app"))
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
			})

			It("installs the custom package using go modules", func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "install_pkg_spec_go_modules"))
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(ContainSubstring("Running: go install -tags cloudfoundry -buildmode pie github.com/full/path/cmd/app"))
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
			})

			It("installs the custom package using go modules and relative paths", func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "install_pkg_spec_mod_relative_pkg"))
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(ContainSubstring("Running: go install -tags cloudfoundry -buildmode pie ./cmd/app"))
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
			})

			It("installs the custom package using vendored go modules", func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "install_pkg_spec_vendored_go_modules"))
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(ContainSubstring("Running: go install -tags cloudfoundry -buildmode pie github.com/full/path/cmd/app"))
				Expect(app.Stdout.String()).NotTo(MatchRegexp("go: downloading github.com/deckarep"))
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
			})
		})
	})

	Context("with cached buildpack dependencies", func() {
		BeforeEach(func() {
			if !cutlass.Cached {
				Skip("running uncached tests")
			}
		})

		Context("app has dependencies", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "with_dependencies", "src", "with_dependencies"))
			})

			It("", func() {
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(MatchRegexp("Hello from foo!"))
				Expect(app.GetBody("/")).To(ContainSubstring("hello, world"))
			})

			AssertNoInternetTraffic("with_dependencies/src/with_dependencies")
		})

		Context("app uses go1.8 and dep", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "go18_dep", "src", "go18_dep"))
			})

			It("successfully stages", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
			})
		})

		Context("app uses go1.8 and dep but no lockfile", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "go18_dep_nolockfile", "src", "go18_dep_nolockfile"))
			})

			It("successfully stages", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
			})
		})

		Context("app uses go1.8 and dep with vendored dependencies", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "go18_dep_vendored", "src", "go18_dep"))
			})

			It("successfully stages", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
			})
		})

		Context("app uses godep with Godeps/_workspace dir", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "go_dependencies"))
			})

			It("", func() {
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(MatchRegexp("Hello from foo!"))
				Expect(app.GetBody("/")).To(ContainSubstring("hello, world"))
			})
		})

		Context("app uses godep and no vendor dir or Godeps/_workspace dir", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "go_no_vendor", "src", "go_no_vendor"))
			})

			It("", func() {
				Expect(app.Push()).To(HaveOccurred())
				Eventually(app.Stdout.String, 3*time.Second).Should(MatchRegexp("(?i)failed"))

				Expect(app.Stdout.String()).To(MatchRegexp("vendor/ directory does not exist."))
			})
		})

		Context("app has vendored dependencies and no Godeps folder", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "native_vendoring"))
			})

			It("successfully stages", func() {
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(MatchRegexp("Init: a.A == 1"))
				Expect(app.GetBody("/")).To(ContainSubstring("Read: a.A == 1"))
			})

			AssertNoInternetTraffic("native_vendoring")
		})

		Context("app has vendored dependencies and custom package spec", func() {
			It("successfully stages", func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "vendored_custom_install_spec"))
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(MatchRegexp("Init: a.A == 1"))
				Expect(app.GetBody("/")).To(ContainSubstring("Read: a.A == 1"))
			})

			AssertNoInternetTraffic("vendored_custom_install_spec")
		})

		Context("app has vendored dependencies and custom package spec", func() {
			It("installs the custom package when vendoring with go modules", func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "install_pkg_spec_vendored_go_modules"))
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(ContainSubstring("Running: go install -tags cloudfoundry -buildmode pie github.com/full/path/cmd/app"))
				Expect(app.Stdout.String()).NotTo(MatchRegexp("go: downloading github.com/deckarep"))
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
			})

			AssertNoInternetTraffic("install_pkg_spec_vendored_go_modules")
		})

		Context("app has vendored dependencies and a vendor.json file", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "with_vendor_json"))
			})

			It("successfully stages", func() {
				PushAppAndConfirm(app)
				Expect(app.Stdout.String()).To(MatchRegexp("Init: a.A == 1"))
				Expect(app.GetBody("/")).To(ContainSubstring("Read: a.A == 1"))
			})
		})

		Context("app with only a single go file and GOPACKAGENAME specified", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "single_file"))
			})

			It("successfully stages", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("simple apps are good"))
			})
		})

		Context("app with only a single go file, a vendor directory, and no GOPACKAGENAME specified", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "vendored_no_gopackagename"))
			})

			It("fails with helpful error", func() {
				Expect(app.Push()).To(HaveOccurred())
				Eventually(app.Stdout.String, 3*time.Second).Should(MatchRegexp("(?i)failed"))

				Expect(app.Stdout.String()).To(MatchRegexp(`To use go native vendoring set the \$GOPACKAGENAME`))
			})
		})

		Context("app has no dependencies", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "go_app"))
			})

			It("", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
				Expect(app.Stdout.String()).To(MatchRegexp(`Installing go [\d\.]+`))
				Expect(app.Stdout.String()).To(MatchRegexp(`Copy \[\/tmp\/`))
			})
		})

		Context("app has before/after compile hooks", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "go_app"))
				app.SetEnv("BP_DEBUG", "1")
			})

			It("", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
				Expect(app.Stdout.String()).To(MatchRegexp("HOOKS 1: BeforeCompile"))
				Expect(app.Stdout.String()).To(MatchRegexp("HOOKS 2: AfterCompile"))
			})
		})

		Context("app has no Procfile", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "no_procfile", "src", "no_procfile"))
			})

			It("", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("go, world"))
				Expect(app.Stdout.String()).To(MatchRegexp(`Installing go [\d\.]+`))
				Expect(app.Stdout.String()).To(MatchRegexp(`Copy \[\/tmp\/`))
			})

			AssertNoInternetTraffic("no_procfile/src/no_procfile")
		})

		Context("expects a non-packaged version of go", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "go99"))
			})

			It("displays useful understandable errors", func() {
				Expect(app.Push()).To(HaveOccurred())
				Eventually(app.Stdout.String, 3*time.Second).Should(MatchRegexp("(?i)failed"))

				Expect(app.Stdout.String()).To(MatchRegexp("Unable to determine Go version to install: no match found for 99.99.99"))

				Expect(app.Stdout.String()).ToNot(MatchRegexp("Installing go99.99.99"))
				Expect(app.Stdout.String()).ToNot(MatchRegexp("Uploading droplet"))
			})
		})

		Context("heroku example", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "heroku_example"))
			})

			It("", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("hello, heroku"))
			})
			AssertNoInternetTraffic("heroku_example")
		})

		Context("a go app using ldflags", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "go_ldflags", "src", "go_app"))
			})

			It("links correctly", func() {
				PushAppAndConfirm(app)

				Expect(app.GetBody("/")).To(ContainSubstring("flag_linked"))
				Expect(app.Stdout.String()).To(ContainSubstring("main.linker_flag=flag_linked"))
			})

			AssertNoInternetTraffic("go_ldflags/src/go_app")
		})

		Context("app uses glide and has vendored dependencies", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "glide_and_vendoring", "src", "glide_and_vendoring"))
			})

			It("", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("hello, world"))
				Expect(app.Stdout.String()).To(ContainSubstring("Hello from foo!"))
				Expect(app.Stdout.String()).To(ContainSubstring("Note: skipping (glide install) due to non-empty vendor directory."))
			})

			AssertNoInternetTraffic("glide_and_vendoring/src/glide_and_vendoring")
		})

		Context("go 1.X app with GO_SETUP_GOPATH_IN_IMAGE", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "gopath_in_container", "src", "go_app"))
				app.SetEnv("GO_SETUP_GOPATH_IN_IMAGE", "true")
			})

			It("", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("GOPATH: /home/vcap/app"))
			})
		})

		Context("go 1.X app with GO_INSTALL_TOOLS_IN_IMAGE", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "toolchain_in_container", "src", "go_app"))
				app.SetEnv("GO_INSTALL_TOOLS_IN_IMAGE", "true")
				app.Disk = "1G"
			})

			It("displays the go version", func() {
				PushAppAndConfirm(app)

				Expect(app.GetBody("/")).To(MatchRegexp(`go version go1\.\d+\.\d+ linux/amd64`))
			})

			Context("running a task", func() {
				BeforeEach(func() {
					if !ApiHasTask() {
						Skip("Running against CF without run task support")
					}
				})

				It("can find the specifed go in the container", func() {
					PushAppAndConfirm(app)

					_, err := app.RunTask(`echo "RUNNING A TASK: $(go version)"`)
					Expect(err).ToNot(HaveOccurred())

					Eventually(func() string { return app.Stdout.String() }, 1*time.Minute).Should(MatchRegexp(`RUNNING A TASK: go version go1\.\d+\.\d+ linux/amd64`))
				})
			})

			Context("and GO_SETUP_GOPATH_IN_IMAGE", func() {
				BeforeEach(func() {
					app.SetEnv("GO_INSTALL_TOOLS_IN_IMAGE", "true")
					app.SetEnv("GO_SETUP_GOPATH_IN_IMAGE", "true")
				})

				It("displays the go version", func() {
					PushAppAndConfirm(app)

					Expect(app.GetBody("/")).To(MatchRegexp(`go version go1\.\d+\.\d+ linux/amd64`))
					Expect(app.GetBody("/gopath")).To(ContainSubstring("GOPATH: /home/vcap/app"))
				})
			})
		})

		Context("packagename is the same as a bash builtin or on path", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "bashbuiltin"))
			})
			It("sets the start command to run this app", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("foo:"))
			})
		})

		Context("app contains a symlink to a directory", func() {
			BeforeEach(func() {
				app = cutlass.New(filepath.Join(bpDir, "fixtures", "symlink_dir"))
			})
			It("sets the start command to run this app", func() {
				PushAppAndConfirm(app)
				Expect(app.GetBody("/")).To(ContainSubstring("simple apps are good"))
			})
		})

	})
})
