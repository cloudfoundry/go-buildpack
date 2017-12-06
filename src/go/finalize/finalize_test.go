package finalize_test

import (
	"go/finalize"
	"go/godep"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"bytes"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=finalize.go --destination=mocks_test.go --package=finalize_test

var _ = Describe("Finalize", func() {
	var (
		vendorTool       string
		buildDir         string
		depsDir          string
		depsIdx          string
		gf               *finalize.Finalizer
		logger           *libbuildpack.Logger
		stager           *libbuildpack.Stager
		buffer           *bytes.Buffer
		err              error
		mockCtrl         *gomock.Controller
		mockCommand      *MockCommand
		goVersion        string
		mainPackageName  string
		goPath           string
		packageList      []string
		buildFlags       []string
		godepConfig      godep.Godep
		vendorExperiment bool
	)

	BeforeEach(func() {
		depsDir, err = ioutil.TempDir("", "go-buildpack.build.")
		Expect(err).To(BeNil())

		depsIdx = "06"
		err = os.MkdirAll(filepath.Join(depsDir, depsIdx), 0755)
		Expect(err).To(BeNil())

		buildDir, err = ioutil.TempDir("", "go-buildpack.build.")
		Expect(err).To(BeNil())

		buffer = new(bytes.Buffer)

		logger = libbuildpack.NewLogger(ansicleaner.New(buffer))

		mockCtrl = gomock.NewController(GinkgoT())
		mockCommand = NewMockCommand(mockCtrl)
	})

	JustBeforeEach(func() {
		args := []string{buildDir, "", depsDir, depsIdx}
		stager = libbuildpack.NewStager(args, logger, &libbuildpack.Manifest{})

		gf = &finalize.Finalizer{
			Stager:           stager,
			Command:          mockCommand,
			Log:              logger,
			VendorTool:       vendorTool,
			GoVersion:        goVersion,
			MainPackageName:  mainPackageName,
			GoPath:           goPath,
			PackageList:      packageList,
			BuildFlags:       buildFlags,
			Godep:            godepConfig,
			VendorExperiment: vendorExperiment,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()

		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())
	})

	Describe("NewFinalizer", func() {
		Context("the vendor tool is godep", func() {
			BeforeEach(func() {
				ioutil.WriteFile(filepath.Join(depsDir, depsIdx, "config.yml"), []byte(`name: "go"
config:
  GoVersion: 1.4.2
  VendorTool: godep
  Godep: '{"ImportPath":"an-import-path"}'
`), 0644)
			})

			It("initializes values from config.yml", func() {
				finalizer, err := finalize.NewFinalizer(stager, mockCommand, logger)
				Expect(err).To(BeNil())

				Expect(finalizer.GoVersion).To(Equal("1.4.2"))
				Expect(finalizer.VendorTool).To(Equal("godep"))
				Expect(finalizer.Godep.ImportPath).To(Equal("an-import-path"))
			})
		})
		Context("the vendor tool is glide", func() {
			BeforeEach(func() {
				ioutil.WriteFile(filepath.Join(depsDir, depsIdx, "config.yml"), []byte(`name: "go"
config:
  GoVersion: 1.2.4
  VendorTool: glide
`), 0644)
			})

			It("initializes values from config.yml", func() {
				finalizer, err := finalize.NewFinalizer(stager, mockCommand, logger)
				Expect(err).To(BeNil())

				Expect(finalizer.GoVersion).To(Equal("1.2.4"))
				Expect(finalizer.VendorTool).To(Equal("glide"))
			})
		})
		Context("the vendor tool is dep", func() {
			BeforeEach(func() {
				ioutil.WriteFile(filepath.Join(depsDir, depsIdx, "config.yml"), []byte(`name: "go"
config:
  GoVersion: 1.9.0
  VendorTool: dep
`), 0644)
			})

			It("initializes values from config.yml", func() {
				finalizer, err := finalize.NewFinalizer(stager, mockCommand, logger)
				Expect(err).To(BeNil())

				Expect(finalizer.GoVersion).To(Equal("1.9.0"))
				Expect(finalizer.VendorTool).To(Equal("dep"))
			})
		})
	})

	Describe("SetMainPackageName", func() {
		Context("the vendor tool is godep", func() {
			BeforeEach(func() {
				vendorTool = "godep"
				godepConfig = godep.Godep{ImportPath: "go-online", GoVersion: "go1.6"}
			})

			It("sets the main package name from Godeps.json", func() {
				err = gf.SetMainPackageName()
				Expect(err).To(BeNil())

				Expect(gf.MainPackageName).To(Equal("go-online"))
			})
		})

		Context("the vendor tool is glide", func() {
			BeforeEach(func() {
				vendorTool = "glide"
			})
			It("sets the main package name to the value of 'glide name'", func() {

				mockCommand.EXPECT().Execute(buildDir, gomock.Any(), gomock.Any(), "glide", "name").Do(func(_ string, buffer, _ io.Writer, _, _ string) {
					_, err := buffer.Write([]byte("go-package-name\n"))
					Expect(err).To(BeNil())
				}).Return(nil)

				err = gf.SetMainPackageName()
				Expect(err).To(BeNil())
				Expect(gf.MainPackageName).To(Equal("go-package-name"))
			})
		})

		AssertRequiresAndUsesGOPACKAGENAME := func() {
			Context("GOPACKAGENAME is not set", func() {
				It("logs an error", func() {
					err = gf.SetMainPackageName()
					Expect(err).NotTo(BeNil())

					Expect(buffer.String()).To(ContainSubstring("**ERROR** To use go native vendoring set the $GOPACKAGENAME"))
					Expect(buffer.String()).To(ContainSubstring("environment variable to your app's package name"))
				})
			})
			Context("GOPACKAGENAME is set", func() {
				var oldGOPACKAGENAME string

				BeforeEach(func() {
					oldGOPACKAGENAME = os.Getenv("GOPACKAGENAME")
					err = os.Setenv("GOPACKAGENAME", "my-go-app")
					Expect(err).To(BeNil())
				})

				AfterEach(func() {
					err = os.Setenv("GOPACKAGENAME", oldGOPACKAGENAME)
					Expect(err).To(BeNil())
				})

				It("returns the package name from GOPACKAGENAME", func() {
					err = gf.SetMainPackageName()
					Expect(err).To(BeNil())

					Expect(gf.MainPackageName).To(Equal("my-go-app"))
				})
			})
		}

		Context("the vendor tool is dep", func() {
			BeforeEach(func() {
				vendorTool = "dep"
			})

			AssertRequiresAndUsesGOPACKAGENAME()
		})

		Context("the vendor tool is go_nativevendoring", func() {
			BeforeEach(func() {
				vendorTool = "go_nativevendoring"
			})

			AssertRequiresAndUsesGOPACKAGENAME()
		})
	})

	Describe("SetupGoPath", func() {
		var (
			oldGoPath               string
			oldGoBin                string
			oldGoSetupGopathInImage string
		)

		BeforeEach(func() {
			mainPackageName = "a/package/name"
			oldGoPath = os.Getenv("GOPATH")
			oldGoBin = os.Getenv("GOBIN")
			oldGoSetupGopathInImage = os.Getenv("GO_SETUP_GOPATH_IN_IMAGE")

			err := ioutil.WriteFile(filepath.Join(buildDir, "main.go"), []byte("xx"), 0644)
			Expect(err).To(BeNil())

			err = os.MkdirAll(filepath.Join(buildDir, "vendor"), 0755)
			Expect(err).To(BeNil())

			err = ioutil.WriteFile(filepath.Join(buildDir, "vendor", "lib.go"), []byte("xx"), 0644)
			Expect(err).To(BeNil())

			err = ioutil.WriteFile(filepath.Join(buildDir, "Procfile"), []byte("xx"), 0644)
			Expect(err).To(BeNil())

			err = ioutil.WriteFile(filepath.Join(buildDir, ".profile"), []byte("xx"), 0644)
			Expect(err).To(BeNil())

			err = os.MkdirAll(filepath.Join(buildDir, ".cloudfoundry"), 0755)
			Expect(err).To(BeNil())

			err = os.MkdirAll(filepath.Join(buildDir, ".profile.d"), 0755)
			Expect(err).To(BeNil())

			err = ioutil.WriteFile(filepath.Join(buildDir, ".profile.d", "filename.sh"), []byte("xx"), 0644)
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			err = os.Setenv("GOPATH", oldGoPath)
			Expect(err).To(BeNil())

			err = os.Setenv("GOBIN", oldGoBin)
			Expect(err).To(BeNil())

			err = os.Setenv("GO_SETUP_GOPATH_IN_IMAGE", oldGoSetupGopathInImage)
			Expect(err).To(BeNil())
		})

		It("creates <buildDir>/bin", func() {
			err = gf.SetupGoPath()
			Expect(err).To(BeNil())

			Expect(filepath.Join(buildDir, "bin")).To(BeADirectory())
		})

		Context("GO_SETUP_GOPATH_IN_IMAGE != true", func() {
			It("sets GOPATH to a temp directory", func() {
				err = gf.SetupGoPath()
				Expect(err).To(BeNil())

				dirRegex := regexp.MustCompile(`\/.{3,}\/gobuildpack\.gopath[0-9]{8,}\/\.go`)
				Expect(dirRegex.Match([]byte(os.Getenv("GOPATH")))).To(BeTrue())
			})

			It("sets GoPath in the compiler", func() {
				err = gf.SetupGoPath()
				Expect(err).To(BeNil())

				dirRegex := regexp.MustCompile(`\/.{3,}\/gobuildpack\.gopath[0-9]{8,}\/\.go`)
				Expect(dirRegex.Match([]byte(gf.GoPath))).To(BeTrue())
			})

			It("copies the buildDir contents to <tempdir>/.go/src/<mainPackageName>", func() {
				err = gf.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(filepath.Join(gf.GoPath, "src", mainPackageName, "main.go")).To(BeAnExistingFile())
				Expect(filepath.Join(gf.GoPath, "src", mainPackageName, "vendor", "lib.go")).To(BeAnExistingFile())
			})

			It("sets GOBIN to <buildDir>/bin", func() {
				err = gf.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(os.Getenv("GOBIN")).To(Equal(filepath.Join(buildDir, "bin")))
			})
		})

		Context("GO_SETUP_GOPATH_IN_IMAGE = true", func() {
			BeforeEach(func() {
				err = os.Setenv("GO_SETUP_GOPATH_IN_IMAGE", "true")
				Expect(err).To(BeNil())
			})

			It("does not move the .profile.d directory", func() {
				err = gf.SetupGoPath()
				Expect(err).To(BeNil())
				Expect(filepath.Join(gf.GoPath, "src", mainPackageName, ".profile.d")).NotTo(BeAnExistingFile())
				Expect(filepath.Join(buildDir, ".profile.d")).To(BeAnExistingFile())
				Expect(filepath.Join(buildDir, ".profile.d", "filename.sh")).To(BeAnExistingFile())
			})

			It("sets GOPATH to the build directory", func() {
				err = gf.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(os.Getenv("GOPATH")).To(Equal(buildDir))
			})

			It("sets GoPath in the compiler", func() {
				err = gf.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(gf.GoPath).To(Equal(buildDir))
			})

			It("moves the buildDir contents to <buildDir>/src/<mainPackageName>", func() {
				err = gf.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(filepath.Join(gf.GoPath, "src", mainPackageName, "main.go")).To(BeAnExistingFile())
				Expect(filepath.Join(gf.GoPath, "src", mainPackageName, "vendor", "lib.go")).To(BeAnExistingFile())
				Expect(filepath.Join(gf.GoPath, "src", mainPackageName, "src", "a/package/name")).NotTo(BeAnExistingFile())
			})

			It("does not move the Procfile", func() {
				err = gf.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(filepath.Join(gf.GoPath, "src", mainPackageName, "Procfile")).NotTo(BeAnExistingFile())
				Expect(filepath.Join(buildDir, "Procfile")).To(BeAnExistingFile())
			})

			It("does not move the .profile script", func() {
				err = gf.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(filepath.Join(gf.GoPath, "src", mainPackageName, ".profile")).NotTo(BeAnExistingFile())
				Expect(filepath.Join(buildDir, ".profile")).To(BeAnExistingFile())
			})

			It("does not move the .cloudfoundry directory", func() {
				err = gf.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(filepath.Join(gf.GoPath, "src", mainPackageName, ".cloudfoundry")).NotTo(BeAnExistingFile())
				Expect(filepath.Join(buildDir, ".cloudfoundry")).To(BeAnExistingFile())
			})

			It("does not set GOBIN", func() {
				err = gf.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(os.Getenv("GOBIN")).To(Equal(oldGoBin))
			})
		})
	})

	Describe("SetBuildFlags", func() {
		Context("link environment variables not set", func() {
			It("contains the default flags", func() {
				gf.SetBuildFlags()
				Expect(gf.BuildFlags).To(Equal([]string{"-tags", "cloudfoundry", "-buildmode", "pie"}))
			})
		})

		Context("link environment variables are set set", func() {
			var (
				oldGoLinkerSymbol string
				oldGoLinkerValue  string
			)

			BeforeEach(func() {
				oldGoLinkerSymbol = os.Getenv("GO_LINKER_SYMBOL")
				oldGoLinkerValue = os.Getenv("GO_LINKER_VALUE")

				err = os.Setenv("GO_LINKER_SYMBOL", "package.main.thing")
				Expect(err).To(BeNil())

				err = os.Setenv("GO_LINKER_VALUE", "some_string")
				Expect(err).To(BeNil())

			})

			AfterEach(func() {
				err = os.Setenv("GO_LINKER_SYMBOL", oldGoLinkerSymbol)
				Expect(err).To(BeNil())

				err = os.Setenv("GO_LINKER_VALUE", oldGoLinkerValue)
				Expect(err).To(BeNil())
			})

			It("contains the ldflags argument", func() {
				gf.SetBuildFlags()
				Expect(gf.BuildFlags).To(Equal([]string{"-tags", "cloudfoundry", "-buildmode", "pie", "-ldflags", "-X package.main.thing=some_string"}))
			})
		})
	})

	Describe("RunGlideInstall", func() {
		var mainPackagePath string

		BeforeEach(func() {
			mainPackageName = "a/package/name"
			goPath, err = ioutil.TempDir("", "go-buildpack.package")
			Expect(err).To(BeNil())

			mainPackagePath = filepath.Join(goPath, "src", mainPackageName)
			err = os.MkdirAll(mainPackagePath, 0755)
			Expect(err).To(BeNil())

			vendorTool = "glide"
		})

		AfterEach(func() {
			err = os.RemoveAll(goPath)
			Expect(err).To(BeNil())
		})

		Context("packages are not already vendored", func() {
			It("uses glide to install the packages", func() {
				mockCommand.EXPECT().Execute(mainPackagePath, gomock.Any(), gomock.Any(), "glide", "install").Return(nil)

				err = gf.RunGlideInstall()
				Expect(err).To(BeNil())
			})
		})

		Context("packages are already vendored", func() {
			BeforeEach(func() {
				err = os.MkdirAll(filepath.Join(mainPackagePath, "vendor", "another-package"), 0755)
				Expect(err).To(BeNil())
			})

			It("does not use glide to install the packages", func() {
				err = gf.RunGlideInstall()
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("RunDepEnsure", func() {
		var mainPackagePath string

		BeforeEach(func() {
			mainPackageName = "a/package/name"
			goPath, err = ioutil.TempDir("", "go-buildpack.package")
			Expect(err).To(BeNil())

			mainPackagePath = filepath.Join(goPath, "src", mainPackageName)
			err = os.MkdirAll(mainPackagePath, 0755)
			Expect(err).To(BeNil())

			vendorTool = "dep"
		})

		AfterEach(func() {
			err = os.RemoveAll(goPath)
			Expect(err).To(BeNil())
		})

		Context("packages are not already vendored", func() {
			It("uses dep to ensure the vendor tree is correct", func() {
				mockCommand.EXPECT().Execute(mainPackagePath, gomock.Any(), gomock.Any(), "dep", "ensure").Return(nil)

				err = gf.RunDepEnsure()
				Expect(err).To(BeNil())
			})
		})

		Context("packages are already vendored", func() {
			BeforeEach(func() {
				err = os.MkdirAll(filepath.Join(mainPackagePath, "vendor", "another-package"), 0755)
				Expect(err).To(BeNil())
			})

			It("does not use dep to ensure the vendor tree is correct", func() {
				err = gf.RunDepEnsure()
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("HandleVendorExperiment", func() {
		Context("version is go1.6", func() {
			var (
				oldGO15VENDOREXPERIMENT string
				newGO15VENDOREXPERIMENT string
			)

			BeforeEach(func() {
				goVersion = "1.6.3"
			})

			JustBeforeEach(func() {
				oldGO15VENDOREXPERIMENT = os.Getenv("GO15VENDOREXPERIMENT")
				err = os.Setenv("GO15VENDOREXPERIMENT", newGO15VENDOREXPERIMENT)
				Expect(err).To(BeNil())
			})

			AfterEach(func() {
				err = os.Setenv("GO15VENDOREXPERIMENT", oldGO15VENDOREXPERIMENT)
				Expect(err).To(BeNil())
			})

			Context("GO15VENDOREXPERIMENT is 0", func() {
				BeforeEach(func() {
					newGO15VENDOREXPERIMENT = "0"
				})

				It("sets VendorExperiment to false", func() {
					err = gf.HandleVendorExperiment()
					Expect(err).To(BeNil())
					Expect(gf.VendorExperiment).To(BeFalse())
				})
			})
			Context("GO15VENDOREXPERIMENT is not 0", func() {
				BeforeEach(func() {
					newGO15VENDOREXPERIMENT = "1"
				})

				It("sets VendorExperiment to true", func() {
					err = gf.HandleVendorExperiment()
					Expect(err).To(BeNil())
					Expect(gf.VendorExperiment).To(BeTrue())
				})
			})
		})

		Context("version is not go1.6", func() {
			BeforeEach(func() {
				goVersion = "1.7.3"
			})

			Context("GO15VENDOREXPERIMENT is set", func() {
				var oldGO15VENDOREXPERIMENT string

				BeforeEach(func() {
					oldGO15VENDOREXPERIMENT = os.Getenv("GO15VENDOREXPERIMENT")
					err = os.Setenv("GO15VENDOREXPERIMENT", "foo")
					Expect(err).To(BeNil())
				})

				AfterEach(func() {
					err = os.Setenv("GO15VENDOREXPERIMENT", oldGO15VENDOREXPERIMENT)
					Expect(err).To(BeNil())
				})

				It("returns an error and logs a message", func() {
					err = gf.HandleVendorExperiment()
					Expect(err).NotTo(BeNil())

					Expect(buffer.String()).To(ContainSubstring("**ERROR** GO15VENDOREXPERIMENT is set, but is not supported by go1.7 and later"))
					Expect(buffer.String()).To(ContainSubstring("Run 'cf unset-env <app> GO15VENDOREXPERIMENT' before pushing again"))
				})

			})
			Context("GO15VENDOREXPERIMENT is not set", func() {
				It("sets VendorExperiment to true", func() {
					err = gf.HandleVendorExperiment()
					Expect(err).To(BeNil())
					Expect(gf.VendorExperiment).To(BeTrue())
				})
			})
		})
	})

	Describe("SetInstallPackages", func() {
		var mainPackagePath string

		BeforeEach(func() {
			mainPackageName = "a/package/name"
			goPath, err = ioutil.TempDir("", "go-buildpack.package")
			Expect(err).To(BeNil())

			mainPackagePath = filepath.Join(goPath, "src", mainPackageName)
			err = os.MkdirAll(mainPackagePath, 0755)
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			err = os.RemoveAll(goPath)
			Expect(err).To(BeNil())
		})

		Context("the vendor tool is godep", func() {
			BeforeEach(func() {
				vendorTool = "godep"
				vendorExperiment = true
			})

			Context("GO_INSTALL_PACKAGE_SPEC is set", func() {
				var oldGoInstallPackageSpec string

				BeforeEach(func() {
					oldGoInstallPackageSpec = os.Getenv("GO_INSTALL_PACKAGE_SPEC")
					err = os.Setenv("GO_INSTALL_PACKAGE_SPEC", "a-package-name another-package")
					Expect(err).To(BeNil())
				})

				AfterEach(func() {
					err = os.Setenv("GO_INSTALL_PACKAGE_SPEC", oldGoInstallPackageSpec)
					Expect(err).To(BeNil())
				})

				It("sets the packages from the env var", func() {
					err = gf.SetInstallPackages()
					Expect(err).To(BeNil())
					Expect(gf.PackageList).To(Equal([]string{"a-package-name", "another-package"}))
				})

				It("logs a warning that it overrode the Godeps.json packages", func() {
					err := gf.SetInstallPackages()
					Expect(err).To(BeNil())
					Expect(buffer.String()).To(ContainSubstring("**WARNING** Using $GO_INSTALL_PACKAGE_SPEC override."))
					Expect(buffer.String()).To(ContainSubstring("    $GO_INSTALL_PACKAGE_SPEC = a-package-name"))
					Expect(buffer.String()).To(ContainSubstring("If this isn't what you want please run:"))
					Expect(buffer.String()).To(ContainSubstring("    cf unset-env <app> GO_INSTALL_PACKAGE_SPEC"))
				})
			})

			Context("GO_INSTALL_PACKAGE_SPEC is not set", func() {
				BeforeEach(func() {
					godepConfig = godep.Godep{ImportPath: "go-online", GoVersion: "go1.6", Packages: []string{"foo", "bar"}}
				})

				Context("No packages in Godeps.json", func() {
					BeforeEach(func() {
						godepConfig = godep.Godep{ImportPath: "go-online", GoVersion: "go1.6"}
					})

					It("sets packages to the default", func() {
						err = gf.SetInstallPackages()
						Expect(err).To(BeNil())
						Expect(gf.PackageList).To(Equal([]string{"."}))
					})

					It("logs a warning that it is using the default", func() {
						err = gf.SetInstallPackages()
						Expect(err).To(BeNil())
						Expect(buffer.String()).To(ContainSubstring("**WARNING** Installing package '.' (default)"))
					})
				})

				Context("there is no vendor directory and no Godeps workspace", func() {
					It("logs a warning that ther is no vendor directory", func() {
						err = gf.SetInstallPackages()
						Expect(err).To(BeNil())

						Expect(buffer.String()).To(ContainSubstring("**WARNING** vendor/ directory does not exist"))
					})
				})

				Context("packages are vendored", func() {
					BeforeEach(func() {
						err = os.MkdirAll(filepath.Join(mainPackagePath, "vendor", "foo"), 0755)
						Expect(err).To(BeNil())
					})

					It("handles the vendoring correctly", func() {
						err = gf.SetInstallPackages()
						Expect(err).To(BeNil())

						Expect(gf.PackageList).To(Equal([]string{filepath.Join(mainPackageName, "vendor", "foo"), "bar"}))
					})

					Context("packages are also in the Godeps/_workspace", func() {
						BeforeEach(func() {
							godepConfig = godep.Godep{ImportPath: "go-online", GoVersion: "go1.6", Packages: []string{"foo", "bar"}, WorkspaceExists: true}
						})

						It("uses the packages from Godeps.json", func() {
							err = gf.SetInstallPackages()
							Expect(err).To(BeNil())

							Expect(gf.PackageList).To(Equal([]string{"foo", "bar"}))
						})

						It("logs a warning about vendor and godeps both existing", func() {
							err = gf.SetInstallPackages()
							Expect(err).To(BeNil())

							Expect(buffer.String()).To(ContainSubstring("**WARNING** Godeps/_workspace/src and vendor/ exist"))
							Expect(buffer.String()).To(ContainSubstring("code may not compile. Please convert all deps to vendor/"))
						})
					})

					Context("vendor experiment is false", func() {
						BeforeEach(func() {
							vendorExperiment = false
						})

						It("uses the packages from Godeps.json", func() {
							err = gf.SetInstallPackages()
							Expect(err).To(BeNil())

							Expect(gf.PackageList).To(Equal([]string{"foo", "bar"}))
						})
					})
				})

				Context("packages are in the Godeps/_workspace", func() {
					BeforeEach(func() {
						godepConfig = godep.Godep{ImportPath: "go-online", GoVersion: "go1.6", Packages: []string{"foo", "bar"}, WorkspaceExists: true}
					})

					It("uses the packages from Godeps.json", func() {
						err = gf.SetInstallPackages()
						Expect(err).To(BeNil())

						Expect(gf.PackageList).To(Equal([]string{"foo", "bar"}))
					})

					It("doesn't log any warnings", func() {
						err = gf.SetInstallPackages()
						Expect(err).To(BeNil())

						Expect(buffer.String()).To(Equal(""))
					})
				})
			})
		})
		Context("the vendor tool is go_nativevendoring", func() {
			BeforeEach(func() {
				vendorTool = "go_nativevendoring"
				vendorExperiment = true
			})

			Context("GO_INSTALL_PACKAGE_SPEC is set", func() {
				var oldGoInstallPackageSpec string

				BeforeEach(func() {
					oldGoInstallPackageSpec = os.Getenv("GO_INSTALL_PACKAGE_SPEC")
					err = os.Setenv("GO_INSTALL_PACKAGE_SPEC", "a-package-name another-package")
					Expect(err).To(BeNil())
				})

				AfterEach(func() {
					err = os.Setenv("GO_INSTALL_PACKAGE_SPEC", oldGoInstallPackageSpec)
					Expect(err).To(BeNil())
				})

				Context("packages are vendored", func() {
					BeforeEach(func() {
						err = os.MkdirAll(filepath.Join(mainPackagePath, "vendor", "another-package"), 0755)
						Expect(err).To(BeNil())
					})
					It("handles the vendoring correctly", func() {
						err = gf.SetInstallPackages()
						Expect(err).To(BeNil())

						Expect(gf.PackageList).To(Equal([]string{"a-package-name", filepath.Join(mainPackageName, "vendor", "another-package")}))
					})
				})
				Context("packages are not vendored", func() {
					It("sets the packages", func() {
						err = gf.SetInstallPackages()
						Expect(err).To(BeNil())

						Expect(gf.PackageList).To(Equal([]string{"a-package-name", "another-package"}))
					})
				})
			})

			Context("GO_INSTALL_PACKAGE_SPEC is not set", func() {
				It("sets packages to  default", func() {
					err = gf.SetInstallPackages()
					Expect(err).To(BeNil())
					Expect(gf.PackageList).To(Equal([]string{"."}))
				})

				It("logs a warning that it is using the default", func() {
					err = gf.SetInstallPackages()
					Expect(err).To(BeNil())
					Expect(buffer.String()).To(ContainSubstring("**WARNING** Installing package '.' (default)"))
				})
			})

			Context("VendorExperiment is false", func() {
				BeforeEach(func() {
					vendorExperiment = false
				})

				It("logs a error and returns an error", func() {
					err = gf.SetInstallPackages()
					Expect(err).NotTo(BeNil())

					Expect(buffer.String()).To(ContainSubstring("**ERROR** $GO15VENDOREXPERIMENT=0. To vendor your packages in vendor/"))
					Expect(buffer.String()).To(ContainSubstring("with go 1.6 this environment variable must unset or set to 1."))
				})
			})
		})
		Context("the vendor tool is dep", func() {
			BeforeEach(func() {
				vendorTool = "dep"
				vendorExperiment = true
			})

			Context("GO_INSTALL_PACKAGE_SPEC is set", func() {
				var oldGoInstallPackageSpec string

				BeforeEach(func() {
					oldGoInstallPackageSpec = os.Getenv("GO_INSTALL_PACKAGE_SPEC")
					err = os.Setenv("GO_INSTALL_PACKAGE_SPEC", "a-package-name another-package")
					Expect(err).To(BeNil())
				})

				AfterEach(func() {
					err = os.Setenv("GO_INSTALL_PACKAGE_SPEC", oldGoInstallPackageSpec)
					Expect(err).To(BeNil())
				})

				Context("packages are vendored", func() {
					BeforeEach(func() {
						err = os.MkdirAll(filepath.Join(mainPackagePath, "vendor", "another-package"), 0755)
						Expect(err).To(BeNil())
					})
					It("handles the vendoring correctly", func() {
						err = gf.SetInstallPackages()
						Expect(err).To(BeNil())

						Expect(gf.PackageList).To(Equal([]string{"a-package-name", filepath.Join(mainPackageName, "vendor", "another-package")}))
					})
				})
				Context("packages are not vendored", func() {
					It("sets the packages", func() {
						err = gf.SetInstallPackages()
						Expect(err).To(BeNil())

						Expect(gf.PackageList).To(Equal([]string{"a-package-name", "another-package"}))
					})
				})
			})

			Context("GO_INSTALL_PACKAGE_SPEC is not set", func() {
				It("sets packages to  default", func() {
					err = gf.SetInstallPackages()
					Expect(err).To(BeNil())
					Expect(gf.PackageList).To(Equal([]string{"."}))
				})

				It("logs a warning that it is using the default", func() {
					err = gf.SetInstallPackages()
					Expect(err).To(BeNil())
					Expect(buffer.String()).To(ContainSubstring("**WARNING** Installing package '.' (default)"))
				})
			})
		})

		Context("the vendor tool is glide", func() {
			BeforeEach(func() {
				vendorTool = "glide"
			})

			Context("GO_INSTALL_PACKAGE_SPEC is set", func() {
				var oldGoInstallPackageSpec string

				BeforeEach(func() {
					oldGoInstallPackageSpec = os.Getenv("GO_INSTALL_PACKAGE_SPEC")
					err = os.Setenv("GO_INSTALL_PACKAGE_SPEC", "a-package-name another-package")
					Expect(err).To(BeNil())
				})

				AfterEach(func() {
					err = os.Setenv("GO_INSTALL_PACKAGE_SPEC", oldGoInstallPackageSpec)
					Expect(err).To(BeNil())
				})

				Context("packages are not vendored", func() {
					It("sets the packages in the compiler", func() {
						err = gf.SetInstallPackages()
						Expect(err).To(BeNil())

						Expect(gf.PackageList).To(Equal([]string{"a-package-name", "another-package"}))
					})
				})

				Context("packages are already vendored", func() {
					BeforeEach(func() {
						err = os.MkdirAll(filepath.Join(mainPackagePath, "vendor", "another-package"), 0755)
						Expect(err).To(BeNil())
					})
					It("handles the vendoring correctly", func() {
						err = gf.SetInstallPackages()
						Expect(err).To(BeNil())

						Expect(gf.PackageList).To(Equal([]string{"a-package-name", filepath.Join(mainPackageName, "vendor", "another-package")}))
					})
				})

			})
			Context("GO_INSTALL_PACKAGE_SPEC is not set", func() {
				It("returns default", func() {
					err = gf.SetInstallPackages()
					Expect(err).To(BeNil())
					Expect(gf.PackageList).To(Equal([]string{"."}))
				})

				It("logs a warning that it is using the default", func() {
					err := gf.SetInstallPackages()
					Expect(err).To(BeNil())
					Expect(buffer.String()).To(ContainSubstring("**WARNING** Installing package '.' (default)"))
				})
			})
		})
	})

	Describe("CompileApp", func() {
		var mainPackagePath string

		BeforeEach(func() {
			mainPackageName = "first"
			packageList = []string{"first", "second"}
			buildFlags = []string{"-a=1", "-b=2"}

			goPath, err = ioutil.TempDir("", "go-buildpack.gopath")
			Expect(err).To(BeNil())

			mainPackagePath = filepath.Join(goPath, "src", "first")
			err = os.MkdirAll(mainPackagePath, 0755)
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			err = os.RemoveAll(goPath)
			Expect(err).To(BeNil())
		})

		Context("the tool is godep", func() {
			BeforeEach(func() {
				vendorTool = "godep"
			})
			Context("godeps workspace dir exists", func() {
				BeforeEach(func() {
					godepConfig = godep.Godep{WorkspaceExists: true}
				})

				It("wraps the install command with godep", func() {
					mockCommand.EXPECT().Execute(mainPackagePath, gomock.Any(), gomock.Any(), "godep", "go", "install", "-a=1", "-b=2", "first", "second").Return(nil)

					err = gf.CompileApp()
					Expect(err).To(BeNil())

					Expect(buffer.String()).To(ContainSubstring("-----> Running: godep go install -a=1 -b=2 first second"))
				})
			})

			Context("godeps workspace dir does not exist", func() {
				BeforeEach(func() {
					godepConfig = godep.Godep{WorkspaceExists: false}
				})

				Context("vendor experiment is true", func() {
					BeforeEach(func() {
						vendorExperiment = true
					})

					It("does not wrap the install command with godep", func() {
						mockCommand.EXPECT().Execute(mainPackagePath, gomock.Any(), gomock.Any(), "go", "install", "-a=1", "-b=2", "first", "second").Return(nil)

						err = gf.CompileApp()
						Expect(err).To(BeNil())

						Expect(buffer.String()).To(ContainSubstring("-----> Running: go install -a=1 -b=2 first second"))
					})

				})

				Context("vendor experiment is false", func() {
					BeforeEach(func() {
						vendorExperiment = false
					})

					It("logs and runs the install command wrapped with godep", func() {
						mockCommand.EXPECT().Execute(mainPackagePath, gomock.Any(), gomock.Any(), "godep", "go", "install", "-a=1", "-b=2", "first", "second").Return(nil)

						err = gf.CompileApp()
						Expect(err).To(BeNil())

						Expect(buffer.String()).To(ContainSubstring("-----> Running: godep go install -a=1 -b=2 first second"))
					})
				})
			})
		})

		AssertLogsAndRunsGenericInstallCommand := func() {
			It("logs and runs the install command it is going to run", func() {
				mockCommand.EXPECT().Execute(mainPackagePath, gomock.Any(), gomock.Any(), "go", "install", "-a=1", "-b=2", "first", "second").Return(nil)

				err = gf.CompileApp()
				Expect(err).To(BeNil())

				Expect(buffer.String()).To(ContainSubstring("-----> Running: go install -a=1 -b=2 first second"))
			})
		}

		Context("the tool is glide", func() {
			BeforeEach(func() {
				vendorTool = "glide"
			})
			AssertLogsAndRunsGenericInstallCommand()
		})

		Context("the tool is dep", func() {
			BeforeEach(func() {
				vendorTool = "dep"
			})
			AssertLogsAndRunsGenericInstallCommand()
		})

		Context("the tool is go_nativevendoring", func() {
			BeforeEach(func() {
				vendorTool = "go_nativevendoring"
			})
			AssertLogsAndRunsGenericInstallCommand()
		})
	})

	Describe("CreateStartupEnvironment", func() {
		var tempDir string

		BeforeEach(func() {
			goVersion = "3.4.5"
			mainPackageName = "a-go-app"
			goPath = buildDir

			goDir := filepath.Join(depsDir, depsIdx, "go"+goVersion, "go")
			err = os.MkdirAll(goDir, 0755)
			Expect(err).To(BeNil())

			tempDir, err = ioutil.TempDir("", "gobuildpack.releaseyml")
			Expect(err).To(BeNil())
		})

		It("writes the buildpack-release-step.yml file", func() {
			err = gf.CreateStartupEnvironment(tempDir)
			Expect(err).To(BeNil())

			contents, err := ioutil.ReadFile(filepath.Join(tempDir, "buildpack-release-step.yml"))
			Expect(err).To(BeNil())

			yaml := `---
default_process_types:
    web: ./bin/a-go-app
`
			Expect(string(contents)).To(Equal(yaml))
		})

		It("writes the go.sh script to <depDir>/profile.d", func() {
			err = gf.CreateStartupEnvironment(tempDir)
			Expect(err).To(BeNil())

			contents, err := ioutil.ReadFile(filepath.Join(gf.Stager.DepDir(), "profile.d", "go.sh"))
			Expect(err).To(BeNil())

			Expect(string(contents)).To(Equal("PATH=$PATH:$HOME/bin\n"))
		})

		Context("GO_INSTALL_TOOLS_IN_IMAGE is not set", func() {
			BeforeEach(func() {
				err = os.MkdirAll(filepath.Join(depsDir, "06", "go3.4.5"), 0755)
				Expect(err).To(BeNil())

				err = ioutil.WriteFile(filepath.Join(depsDir, "06", "go3.4.5", "thing.txt"), []byte("abc"), 0644)
				Expect(err).To(BeNil())

				err = ioutil.WriteFile(filepath.Join(depsDir, "06", "config.yml"), []byte("some yaml"), 0644)
				Expect(err).To(BeNil())
			})

			It("clears the dep dir", func() {
				err = gf.CreateStartupEnvironment(tempDir)
				Expect(err).To(BeNil())

				Expect(filepath.Join(depsDir, "06", "go3.4.5")).NotTo(BeADirectory())

				content, err := ioutil.ReadFile(filepath.Join(filepath.Join(depsDir, "06"), "config.yml"))
				Expect(err).To(BeNil())
				Expect(string(content)).To(Equal("some yaml"))
			})
		})

		Context("GO_INSTALL_TOOLS_IN_IMAGE = true", func() {
			var oldGoInstallToolsInImage string

			BeforeEach(func() {
				oldGoInstallToolsInImage = os.Getenv("GO_INSTALL_TOOLS_IN_IMAGE")
				err = os.Setenv("GO_INSTALL_TOOLS_IN_IMAGE", "true")
				Expect(err).To(BeNil())
			})

			AfterEach(func() {
				err = os.Setenv("GO_INSTALL_TOOLS_IN_IMAGE", oldGoInstallToolsInImage)
				Expect(err).To(BeNil())
			})

			It("does not remove the go toolchain", func() {
				err = gf.CreateStartupEnvironment(tempDir)
				Expect(err).To(BeNil())

				Expect(filepath.Join(depsDir, "06", "go3.4.5")).To(BeADirectory())
			})

			It("logs that the tool chain was copied", func() {
				err = gf.CreateStartupEnvironment(tempDir)
				Expect(err).To(BeNil())

				Expect(buffer.String()).To(ContainSubstring("-----> Leaving go tool chain in $GOROOT=$DEPS_DIR/06/go3.4.5/go"))
			})
		})

		Context("GO_SETUP_GOPATH_IN_IMAGE = true", func() {
			var (
				oldGoSetupGopathInImage string
			)

			BeforeEach(func() {
				oldGoSetupGopathInImage = os.Getenv("GO_SETUP_GOPATH_IN_IMAGE")
				err = os.Setenv("GO_SETUP_GOPATH_IN_IMAGE", "true")
				Expect(err).To(BeNil())

				err = os.MkdirAll(filepath.Join(buildDir, "pkg"), 0755)
				Expect(err).To(BeNil())
			})

			AfterEach(func() {
				err = os.Setenv("GO_SETUP_GOPATH_IN_IMAGE", oldGoSetupGopathInImage)
				Expect(err).To(BeNil())
			})

			It("cleans up the pkg directory", func() {
				err = gf.CreateStartupEnvironment(tempDir)
				Expect(err).To(BeNil())

				Expect(buffer.String()).To(ContainSubstring("-----> Cleaning up $GOPATH/pkg"))
				Expect(filepath.Join(buildDir, "pkg")).ToNot(BeADirectory())
			})

			It("writes the zzgopath.sh script to <depDir>/profile.d", func() {
				err = gf.CreateStartupEnvironment(tempDir)
				Expect(err).To(BeNil())

				contents, err := ioutil.ReadFile(filepath.Join(gf.Stager.DepDir(), "profile.d", "zzgopath.sh"))
				Expect(err).To(BeNil())

				Expect(string(contents)).To(ContainSubstring("export GOPATH=$HOME"))
				Expect(string(contents)).To(ContainSubstring("cd $GOPATH/src/" + mainPackageName))
			})
		})
	})

})
