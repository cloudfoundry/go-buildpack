package golang_test

import (
	g "compile/golang"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"bytes"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=vendor/github.com/cloudfoundry/libbuildpack/manifest.go --destination=mocks_manifest_test.go --package=golang_test --imports=.=github.com/cloudfoundry/libbuildpack
//go:generate mockgen -source=vendor/github.com/cloudfoundry/libbuildpack/command_runner.go --destination=mocks_command_runner_test.go --package=golang_test

var _ = Describe("Compile", func() {
	var (
		buildDir          string
		cacheDir          string
		depsDir           string
		gc                *g.Compiler
		logger            libbuildpack.Logger
		buffer            *bytes.Buffer
		err               error
		mockCtrl          *gomock.Controller
		mockManifest      *MockManifest
		mockCommandRunner *MockCommandRunner
		vendorTool        string
		goVersion         string
		mainPackageName   string
		goPath            string
		packageList       []string
		buildFlags        []string
		godep             g.Godep
		vendorExperiment  bool
	)

	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "go-buildpack.build.")
		Expect(err).To(BeNil())

		cacheDir, err = ioutil.TempDir("", "go-buildpack.cache.")
		Expect(err).To(BeNil())

		depsDir = ""

		buffer = new(bytes.Buffer)

		logger = libbuildpack.NewLogger()
		logger.SetOutput(buffer)

		mockCtrl = gomock.NewController(GinkgoT())
		mockManifest = NewMockManifest(mockCtrl)
		mockCommandRunner = NewMockCommandRunner(mockCtrl)
	})

	JustBeforeEach(func() {
		bpc := &libbuildpack.Compiler{BuildDir: buildDir,
			CacheDir: cacheDir,
			DepsDir:  depsDir,
			Manifest: mockManifest,
			Log:      logger,
			Command:  mockCommandRunner,
		}

		gc = &g.Compiler{Compiler: bpc,
			VendorTool:       vendorTool,
			GoVersion:        goVersion,
			MainPackageName:  mainPackageName,
			GoPath:           goPath,
			PackageList:      packageList,
			BuildFlags:       buildFlags,
			Godep:            godep,
			VendorExperiment: vendorExperiment,
		}
	})

	AfterEach(func() {
		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())

		err = os.RemoveAll(cacheDir)
		Expect(err).To(BeNil())
	})

	Describe("SelectVendorTool", func() {
		Context("There is a Godeps.json", func() {
			var (
				godepsJson         string
				godepsJsonContents string
			)

			JustBeforeEach(func() {
				err = os.MkdirAll(filepath.Join(buildDir, "Godeps"), 0755)
				Expect(err).To(BeNil())

				godepsJson = filepath.Join(buildDir, "Godeps", "Godeps.json")
				err = ioutil.WriteFile(godepsJson, []byte(godepsJsonContents), 0644)
				Expect(err).To(BeNil())
			})

			Context("the json is valid", func() {
				BeforeEach(func() {
					godepsJsonContents = `
{
	"ImportPath": "go-online",
	"GoVersion": "go1.6",
	"Deps": []
}					
`
				})
				It("sets the tool to godep", func() {
					err = gc.SelectVendorTool()
					Expect(err).To(BeNil())

					Expect(gc.VendorTool).To(Equal("godep"))
				})
				It("logs that it is checking the Godeps.json file", func() {
					err = gc.SelectVendorTool()
					Expect(err).To(BeNil())

					Expect(buffer.String()).To(ContainSubstring("-----> Checking Godeps/Godeps.json file"))
				})
				It("stores the Godep info in the GoCompiler struct", func() {
					err = gc.SelectVendorTool()
					Expect(err).To(BeNil())

					Expect(gc.Godep.ImportPath).To(Equal("go-online"))
					Expect(gc.Godep.GoVersion).To(Equal("go1.6"))

					var empty []string
					Expect(gc.Godep.Packages).To(Equal(empty))
				})

				Context("godeps workspace exists", func() {
					BeforeEach(func() {
						err = os.MkdirAll(filepath.Join(buildDir, "Godeps", "_workspace", "src"), 0755)
						Expect(err).To(BeNil())
					})

					It("sets Godep.WorkspaceExists to true", func() {
						err = gc.SelectVendorTool()
						Expect(err).To(BeNil())

						Expect(gc.Godep.WorkspaceExists).To(BeTrue())
					})
				})

				Context("godeps workspace does not exist", func() {
					It("sets Godep.WorkspaceExists to false", func() {
						err = gc.SelectVendorTool()
						Expect(err).To(BeNil())

						Expect(gc.Godep.WorkspaceExists).To(BeFalse())
					})
				})
			})

			Context("bad Godeps.json file", func() {
				BeforeEach(func() {
					godepsJsonContents = "not actually JSON"
				})

				It("logs that the Godeps.json file is invalid and returns an error", func() {
					err := gc.SelectVendorTool()
					Expect(err).NotTo(BeNil())

					Expect(buffer.String()).To(ContainSubstring("**ERROR** Bad Godeps/Godeps.json file"))
				})
			})
		})

		Context("there is a .godir file", func() {
			BeforeEach(func() {
				err = ioutil.WriteFile(filepath.Join(buildDir, ".godir"), []byte("xxx"), 0644)
			})

			It("logs that .godir is deprecated and returns an error", func() {
				err = gc.SelectVendorTool()
				Expect(err).NotTo(BeNil())

				Expect(buffer.String()).To(ContainSubstring("**ERROR** Deprecated, .godir file found! Please update to supported Godep or Glide dependency managers."))
				Expect(buffer.String()).To(ContainSubstring("See https://github.com/tools/godep or https://github.com/Masterminds/glide for usage information."))
			})
		})

		Context("there is a glide.yaml file", func() {
			BeforeEach(func() {
				err = ioutil.WriteFile(filepath.Join(buildDir, "glide.yaml"), []byte("xxx"), 0644)
				dep := libbuildpack.Dependency{Name: "go", Version: "1.14.3"}

				mockManifest.EXPECT().DefaultVersion("go").Return(dep, nil)
			})

			It("sets the tool to glide", func() {
				err = gc.SelectVendorTool()
				Expect(err).To(BeNil())

				Expect(gc.VendorTool).To(Equal("glide"))
			})
		})

		Context("the app contains src/**/**/*.go", func() {
			BeforeEach(func() {
				err = os.MkdirAll(filepath.Join(buildDir, "src", "package"), 0755)
				Expect(err).To(BeNil())

				err = ioutil.WriteFile(filepath.Join(buildDir, "src", "package", "thing.go"), []byte("xxx"), 0644)
				Expect(err).To(BeNil())
			})

			It("logs that gb is deprecated and returns an error", func() {
				err = gc.SelectVendorTool()
				Expect(err).NotTo(BeNil())

				Expect(buffer.String()).To(ContainSubstring("**ERROR** Cloud Foundry does not support the GB package manager."))
				Expect(buffer.String()).To(ContainSubstring("We currently only support the Godep and Glide package managers for go apps"))
				Expect(buffer.String()).To(ContainSubstring("For support please file an issue: https://github.com/cloudfoundry/go-buildpack/issues"))

			})
		})

		Context("none of the above", func() {
			BeforeEach(func() {
				dep := libbuildpack.Dependency{Name: "go", Version: "2.0.1"}
				mockManifest.EXPECT().DefaultVersion("go").Return(dep, nil)
			})

			It("sets the tool to go_nativevendoring", func() {
				err = gc.SelectVendorTool()
				Expect(err).To(BeNil())

				Expect(gc.VendorTool).To(Equal("go_nativevendoring"))
			})
		})
	})

	Describe("InstallVendorTool", func() {
		var (
			oldPath string
			tempDir string
		)

		BeforeEach(func() {
			oldPath = os.Getenv("PATH")
			tempDir, err = ioutil.TempDir("", "go-buildpack.tmp")
			Expect(err).To(BeNil())
		})

		AfterEach(func() {
			err = os.Setenv("PATH", oldPath)
			Expect(err).To(BeNil())
		})

		Context("the tool is godep", func() {
			BeforeEach(func() {
				vendorTool = "godep"
			})

			It("installs godep to the requested dir, adding it to the PATH", func() {
				installDir := filepath.Join(tempDir, "godep")

				mockManifest.EXPECT().InstallOnlyVersion("godep", installDir).Return(nil)

				err = gc.InstallVendorTool(tempDir)
				Expect(err).To(BeNil())

				newPath := os.Getenv("PATH")
				Expect(newPath).To(Equal(fmt.Sprintf("%s:%s", filepath.Join(installDir, "bin"), oldPath)))
			})
		})
		Context("the tool is glide", func() {
			BeforeEach(func() {
				vendorTool = "glide"
			})

			It("installs glide to the requested dir, adding it to the PATH", func() {
				installDir := filepath.Join(tempDir, "glide")

				mockManifest.EXPECT().InstallOnlyVersion("glide", installDir).Return(nil)

				err = gc.InstallVendorTool(tempDir)
				Expect(err).To(BeNil())

				newPath := os.Getenv("PATH")
				Expect(newPath).To(Equal(fmt.Sprintf("%s:%s", filepath.Join(installDir, "bin"), oldPath)))
			})
		})
		Context("the tool is go_nativevendoring", func() {
			BeforeEach(func() {
				vendorTool = "go_nativevendoring"
			})

			It("does not install anything", func() {
				err = gc.InstallVendorTool(tempDir)
				Expect(err).To(BeNil())

				newPath := os.Getenv("PATH")
				Expect(newPath).To(Equal(oldPath))
			})
		})
	})

	Describe("SelectGoVersion", func() {
		BeforeEach(func() {
			versions := []string{"1.8.0", "1.7.5", "1.7.4", "1.6.3", "1.6.4", "34.34.0", "1.14.3"}
			mockManifest.EXPECT().AllDependencyVersions("go").Return(versions)
		})
		Context("godep", func() {
			BeforeEach(func() {
				vendorTool = "godep"
				godep = g.Godep{ImportPath: "go-online", GoVersion: "go1.6"}
			})

			Context("GOVERSION not set", func() {
				It("sets the go version from Godeps.json", func() {
					err = gc.SelectGoVersion()
					Expect(err).To(BeNil())

					Expect(gc.GoVersion).To(Equal("1.6.4"))
				})
			})

			Context("GOVERSION is set", func() {
				var oldGOVERSION string

				BeforeEach(func() {
					oldGOVERSION = os.Getenv("GOVERSION")
					err = os.Setenv("GOVERSION", "go34.34")
					Expect(err).To(BeNil())
				})

				AfterEach(func() {
					err = os.Setenv("GOVERSION", oldGOVERSION)
					Expect(err).To(BeNil())
				})

				It("sets the go version from GOVERSION and logs a warning", func() {
					err = gc.SelectGoVersion()
					Expect(err).To(BeNil())

					Expect(gc.GoVersion).To(Equal("34.34.0"))
					Expect(buffer.String()).To(ContainSubstring("**WARNING** Using $GOVERSION override.\n"))
					Expect(buffer.String()).To(ContainSubstring("    $GOVERSION = go34.34\n"))
					Expect(buffer.String()).To(ContainSubstring("If this isn't what you want please run:\n"))
					Expect(buffer.String()).To(ContainSubstring("    cf unset-env <app> GOVERSION"))
				})
			})
		})
		Context("glide or go_nativevendoring", func() {
			BeforeEach(func() {
				dep := libbuildpack.Dependency{Name: "go", Version: "1.14.3"}

				mockManifest.EXPECT().DefaultVersion("go").Return(dep, nil)
			})

			Context("GOVERSION is notset", func() {
				BeforeEach(func() {
					vendorTool = "glide"
				})

				It("sets the go version to the default from the manifest.yml", func() {
					err = gc.SelectGoVersion()
					Expect(err).To(BeNil())

					Expect(gc.GoVersion).To(Equal("1.14.3"))
				})
			})
			Context("GOVERSION is set", func() {
				var oldGOVERSION string

				BeforeEach(func() {
					oldGOVERSION = os.Getenv("GOVERSION")
					err = os.Setenv("GOVERSION", "go34.34")
					Expect(err).To(BeNil())
					vendorTool = "go_nativevendoring"
				})

				AfterEach(func() {
					err = os.Setenv("GOVERSION", oldGOVERSION)
					Expect(err).To(BeNil())
				})

				It("sets the go version from GOVERSION", func() {
					err = gc.SelectGoVersion()
					Expect(err).To(BeNil())

					Expect(gc.GoVersion).To(Equal("34.34.0"))
				})
			})
		})
	})

	Describe("ParseGoVersion", func() {
		BeforeEach(func() {
			versions := []string{"1.8.0", "1.7.5", "1.7.4", "1.6.3", "1.6.4"}
			mockManifest.EXPECT().AllDependencyVersions("go").Return(versions)
		})

		Context("a fully specified version is passed in", func() {
			It("returns the same value", func() {
				ver, err := gc.ParseGoVersion("go1.7.4")
				Expect(err).To(BeNil())

				Expect(ver).To(Equal("1.7.4"))
			})
		})

		Context("a version line is passed in", func() {
			It("returns the latest version of that line", func() {
				ver, err := gc.ParseGoVersion("go1.6")
				Expect(err).To(BeNil())

				Expect(ver).To(Equal("1.6.4"))
			})
		})

	})

	Describe("InstallGo", func() {
		var (
			oldGoRoot    string
			oldPath      string
			goInstallDir string
			dep          libbuildpack.Dependency
		)

		BeforeEach(func() {
			goVersion = "1.3.4"
			oldPath = os.Getenv("PATH")
			oldPath = os.Getenv("GOROOT")
			goInstallDir = filepath.Join(cacheDir, "go1.3.4")
			dep = libbuildpack.Dependency{Name: "go", Version: "1.3.4"}
		})

		AfterEach(func() {
			err = os.Setenv("PATH", oldPath)
			Expect(err).To(BeNil())

			err = os.Setenv("GOROOT", oldGoRoot)
			Expect(err).To(BeNil())
		})

		Context("go is already cached", func() {
			BeforeEach(func() {
				err = os.MkdirAll(filepath.Join(goInstallDir, "go"), 0755)
				Expect(err).To(BeNil())
			})

			It("uses the cached version", func() {
				err = gc.InstallGo()
				Expect(err).To(BeNil())

				Expect(buffer.String()).To(ContainSubstring("-----> Using go 1.3.4"))
			})

			It("Creates a bin directory", func() {
				err = gc.InstallGo()
				Expect(err).To(BeNil())

				Expect(filepath.Join(buildDir, "bin")).To(BeADirectory())
			})

			It("Sets up GOROOT", func() {
				err = gc.InstallGo()
				Expect(err).To(BeNil())

				Expect(os.Getenv("GOROOT")).To(Equal(filepath.Join(goInstallDir, "go")))
			})

			It("adds go to the PATH", func() {
				err = gc.InstallGo()
				Expect(err).To(BeNil())

				newPath := fmt.Sprintf("%s:%s", filepath.Join(goInstallDir, "go", "bin"), oldPath)
				Expect(os.Getenv("PATH")).To(Equal(newPath))
			})
		})

		Context("go is not already cached", func() {
			BeforeEach(func() {
				err = os.MkdirAll(filepath.Join(cacheDir, "go4.3.2", "go"), 0755)
				Expect(err).To(BeNil())
				mockManifest.EXPECT().InstallDependency(dep, goInstallDir).Return(nil)
			})

			It("Creates a bin directory", func() {
				err = gc.InstallGo()
				Expect(err).To(BeNil())

				Expect(filepath.Join(buildDir, "bin")).To(BeADirectory())
			})

			It("Sets up GOROOT", func() {
				err = gc.InstallGo()
				Expect(err).To(BeNil())

				Expect(os.Getenv("GOROOT")).To(Equal(filepath.Join(goInstallDir, "go")))
			})

			It("adds go to the PATH", func() {
				err = gc.InstallGo()
				Expect(err).To(BeNil())

				newPath := fmt.Sprintf("%s:%s", filepath.Join(goInstallDir, "go", "bin"), oldPath)
				Expect(os.Getenv("PATH")).To(Equal(newPath))
			})

			It("installs go", func() {
				err = gc.InstallGo()
				Expect(err).To(BeNil())
			})

			It("clears the cache", func() {
				err = gc.InstallGo()
				Expect(err).To(BeNil())

				Expect(filepath.Join(cacheDir, "go4.3.2", "go")).NotTo(BeADirectory())
			})
		})
	})

	Describe("SetMainPackageName", func() {
		Context("the vendor tool is godep", func() {
			BeforeEach(func() {
				vendorTool = "godep"
				godep = g.Godep{ImportPath: "go-online", GoVersion: "go1.6"}
			})

			It("sets the main package name from Godeps.json", func() {
				err = gc.SetMainPackageName()
				Expect(err).To(BeNil())

				Expect(gc.MainPackageName).To(Equal("go-online"))
			})
		})

		Context("the vendor tool is glide", func() {
			BeforeEach(func() {
				vendorTool = "glide"
			})
			It("sets the main package name to the value of 'glide name'", func() {
				gomock.InOrder(
					mockCommandRunner.EXPECT().SetDir(buildDir),
					mockCommandRunner.EXPECT().CaptureStdout("glide", "name").Return("go-package-name\n", nil),
					mockCommandRunner.EXPECT().SetDir(""),
				)

				err = gc.SetMainPackageName()
				Expect(err).To(BeNil())
				Expect(gc.MainPackageName).To(Equal("go-package-name"))
			})
		})

		Context("the vendor tool is go_nativevendoring", func() {
			BeforeEach(func() {
				vendorTool = "go_nativevendoring"
			})

			Context("GOPACKAGENAME is not set", func() {
				It("logs an error", func() {
					err = gc.SetMainPackageName()
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
					err = gc.SetMainPackageName()
					Expect(err).To(BeNil())

					Expect(gc.MainPackageName).To(Equal("my-go-app"))
				})
			})
		})
	})

	Describe("CheckBinDirectory", func() {
		Context("no directory exists", func() {
			It("returns nil", func() {
				err = gc.CheckBinDirectory()
				Expect(err).To(BeNil())
			})
		})

		Context("a bin directory exists", func() {
			BeforeEach(func() {
				err = os.MkdirAll(filepath.Join(buildDir, "bin"), 0755)
				Expect(err).To(BeNil())
			})

			It("returns nil", func() {
				err := gc.CheckBinDirectory()
				Expect(err).To(BeNil())
			})
		})

		Context("a bin file exists", func() {
			BeforeEach(func() {
				err = ioutil.WriteFile(filepath.Join(buildDir, "bin"), []byte("xxx"), 0644)
				Expect(err).To(BeNil())
			})

			It("returns and logs an error", func() {
				err := gc.CheckBinDirectory()
				Expect(err).NotTo(BeNil())

				Expect(buffer.String()).To(ContainSubstring("**ERROR** File bin exists and is not a directory."))
			})
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
			err = gc.SetupGoPath()
			Expect(err).To(BeNil())

			Expect(filepath.Join(buildDir, "bin")).To(BeADirectory())
		})

		Context("GO_SETUP_GOPATH_IN_IMAGE != true", func() {
			It("sets GOPATH to a temp directory", func() {
				err = gc.SetupGoPath()
				Expect(err).To(BeNil())

				dirRegex := regexp.MustCompile(`\/.{3,}\/gobuildpack\.gopath[0-9]{8,}\/\.go`)
				Expect(dirRegex.Match([]byte(os.Getenv("GOPATH")))).To(BeTrue())
			})

			It("sets GoPath in the compiler", func() {
				err = gc.SetupGoPath()
				Expect(err).To(BeNil())

				dirRegex := regexp.MustCompile(`\/.{3,}\/gobuildpack\.gopath[0-9]{8,}\/\.go`)
				Expect(dirRegex.Match([]byte(gc.GoPath))).To(BeTrue())
			})

			It("copies the buildDir contents to <tempdir>/.go/src/<mainPackageName>", func() {
				err = gc.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(filepath.Join(gc.GoPath, "src", mainPackageName, "main.go")).To(BeAnExistingFile())
				Expect(filepath.Join(gc.GoPath, "src", mainPackageName, "vendor", "lib.go")).To(BeAnExistingFile())
			})

			It("sets GOBIN to <buildDir>/bin", func() {
				err = gc.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(os.Getenv("GOBIN")).To(Equal(filepath.Join(buildDir, "bin")))
			})
		})

		Context("GO_SETUP_GOPATH_IN_IMAGE = true", func() {
			BeforeEach(func() {
				err = os.Setenv("GO_SETUP_GOPATH_IN_IMAGE", "true")
			})

			It("sets GOPATH to the build directory", func() {
				err = gc.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(os.Getenv("GOPATH")).To(Equal(buildDir))
			})

			It("sets GoPath in the compiler", func() {
				err = gc.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(gc.GoPath).To(Equal(buildDir))
			})

			It("moves the buildDir contents to <buildDir>/src/<mainPackageName>", func() {
				err = gc.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(filepath.Join(gc.GoPath, "src", mainPackageName, "main.go")).To(BeAnExistingFile())
				Expect(filepath.Join(gc.GoPath, "src", mainPackageName, "vendor", "lib.go")).To(BeAnExistingFile())
				Expect(filepath.Join(gc.GoPath, "src", mainPackageName, "src", "a/package/name")).NotTo(BeAnExistingFile())
			})

			It("does not move the Procfile", func() {
				err = gc.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(filepath.Join(gc.GoPath, "src", mainPackageName, "Procfile")).NotTo(BeAnExistingFile())
				Expect(filepath.Join(buildDir, "Procfile")).To(BeAnExistingFile())
			})

			It("does not move the .profile script", func() {
				err = gc.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(filepath.Join(gc.GoPath, "src", mainPackageName, ".profile")).NotTo(BeAnExistingFile())
				Expect(filepath.Join(buildDir, ".profile")).To(BeAnExistingFile())
			})

			It("does not set GOBIN", func() {
				err = gc.SetupGoPath()
				Expect(err).To(BeNil())

				Expect(os.Getenv("GOBIN")).To(Equal(oldGoBin))
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
					err = gc.HandleVendorExperiment()
					Expect(err).To(BeNil())
					Expect(gc.VendorExperiment).To(BeFalse())
				})
			})
			Context("GO15VENDOREXPERIMENT is not 0", func() {
				BeforeEach(func() {
					newGO15VENDOREXPERIMENT = "1"
				})

				It("sets VendorExperiment to true", func() {
					err = gc.HandleVendorExperiment()
					Expect(err).To(BeNil())
					Expect(gc.VendorExperiment).To(BeTrue())
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
					err = gc.HandleVendorExperiment()
					Expect(err).NotTo(BeNil())

					Expect(buffer.String()).To(ContainSubstring("**ERROR** GO15VENDOREXPERIMENT is set, but is not supported by go1.7 and later"))
					Expect(buffer.String()).To(ContainSubstring("Run 'cf unset-env <app> GO15VENDOREXPERIMENT' before pushing again"))
				})

			})
			Context("GO15VENDOREXPERIMENT is not set", func() {
				It("sets VendorExperiment to true", func() {
					err = gc.HandleVendorExperiment()
					Expect(err).To(BeNil())
					Expect(gc.VendorExperiment).To(BeTrue())
				})
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
				gomock.InOrder(
					mockCommandRunner.EXPECT().SetDir(mainPackagePath),
					mockCommandRunner.EXPECT().Run("glide", "install").Return(nil),
					mockCommandRunner.EXPECT().SetDir(""),
				)

				err = gc.RunGlideInstall()
				Expect(err).To(BeNil())
			})
		})

		Context("packages are already vendored", func() {
			BeforeEach(func() {
				err = os.MkdirAll(filepath.Join(mainPackagePath, "vendor", "another-package"), 0755)
				Expect(err).To(BeNil())
			})

			It("does not use glide to install the packages", func() {
				err = gc.RunGlideInstall()
				Expect(err).To(BeNil())
			})
		})
	})

	Describe("SetBuildFlags", func() {
		Context("link environment variables not set", func() {
			It("contains the default flags", func() {
				gc.SetBuildFlags()
				Expect(gc.BuildFlags).To(Equal([]string{"-tags", "cloudfoundry", "-buildmode", "pie"}))
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
				gc.SetBuildFlags()
				Expect(gc.BuildFlags).To(Equal([]string{"-tags", "cloudfoundry", "-buildmode", "pie", "-ldflags", "-X package.main.thing=some_string"}))
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
					err = gc.SetInstallPackages()
					Expect(err).To(BeNil())
					Expect(gc.PackageList).To(Equal([]string{"a-package-name", "another-package"}))
				})

				It("logs a warning that it overrode the Godeps.json packages", func() {
					err := gc.SetInstallPackages()
					Expect(err).To(BeNil())
					Expect(buffer.String()).To(ContainSubstring("**WARNING** Using $GO_INSTALL_PACKAGE_SPEC override."))
					Expect(buffer.String()).To(ContainSubstring("    $GO_INSTALL_PACKAGE_SPEC = a-package-name"))
					Expect(buffer.String()).To(ContainSubstring("If this isn't what you want please run:"))
					Expect(buffer.String()).To(ContainSubstring("    cf unset-env <app> GO_INSTALL_PACKAGE_SPEC"))
				})
			})

			Context("GO_INSTALL_PACKAGE_SPEC is not set", func() {
				BeforeEach(func() {
					godep = g.Godep{ImportPath: "go-online", GoVersion: "go1.6", Packages: []string{"foo", "bar"}}
				})

				Context("No packages in Godeps.json", func() {
					BeforeEach(func() {
						godep = g.Godep{ImportPath: "go-online", GoVersion: "go1.6"}
					})

					It("sets packages to the default", func() {
						err = gc.SetInstallPackages()
						Expect(err).To(BeNil())
						Expect(gc.PackageList).To(Equal([]string{"."}))
					})

					It("logs a warning that it is using the default", func() {
						err = gc.SetInstallPackages()
						Expect(err).To(BeNil())
						Expect(buffer.String()).To(ContainSubstring("**WARNING** Installing package '.' (default)"))
					})
				})

				Context("there is no vendor directory and no Godeps workspace", func() {
					It("logs a warning that ther is no vendor directory", func() {
						err = gc.SetInstallPackages()
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
						err = gc.SetInstallPackages()
						Expect(err).To(BeNil())

						Expect(gc.PackageList).To(Equal([]string{filepath.Join(mainPackageName, "vendor", "foo"), "bar"}))
					})

					Context("packages are also in the Godeps/_workspace", func() {
						BeforeEach(func() {
							godep = g.Godep{ImportPath: "go-online", GoVersion: "go1.6", Packages: []string{"foo", "bar"}, WorkspaceExists: true}
						})

						It("uses the packages from Godeps.json", func() {
							err = gc.SetInstallPackages()
							Expect(err).To(BeNil())

							Expect(gc.PackageList).To(Equal([]string{"foo", "bar"}))
						})

						It("logs a warning about vendor and godeps both existing", func() {
							err = gc.SetInstallPackages()
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
							err = gc.SetInstallPackages()
							Expect(err).To(BeNil())

							Expect(gc.PackageList).To(Equal([]string{"foo", "bar"}))
						})
					})
				})

				Context("packages are in the Godeps/_workspace", func() {
					BeforeEach(func() {
						godep = g.Godep{ImportPath: "go-online", GoVersion: "go1.6", Packages: []string{"foo", "bar"}, WorkspaceExists: true}
					})

					It("uses the packages from Godeps.json", func() {
						err = gc.SetInstallPackages()
						Expect(err).To(BeNil())

						Expect(gc.PackageList).To(Equal([]string{"foo", "bar"}))
					})

					It("doesn't log any warnings", func() {
						err = gc.SetInstallPackages()
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
						err = gc.SetInstallPackages()
						Expect(err).To(BeNil())

						Expect(gc.PackageList).To(Equal([]string{"a-package-name", filepath.Join(mainPackageName, "vendor", "another-package")}))
					})
				})
				Context("packages are not vendored", func() {
					It("sets the packages", func() {
						err = gc.SetInstallPackages()
						Expect(err).To(BeNil())

						Expect(gc.PackageList).To(Equal([]string{"a-package-name", "another-package"}))
					})
				})
			})

			Context("GO_INSTALL_PACKAGE_SPEC is not set", func() {
				It("sets packages to  default", func() {
					err = gc.SetInstallPackages()
					Expect(err).To(BeNil())
					Expect(gc.PackageList).To(Equal([]string{"."}))
				})

				It("logs a warning that it is using the default", func() {
					err = gc.SetInstallPackages()
					Expect(err).To(BeNil())
					Expect(buffer.String()).To(ContainSubstring("**WARNING** Installing package '.' (default)"))
				})
			})

			Context("VendorExperiment is false", func() {
				BeforeEach(func() {
					vendorExperiment = false
				})

				It("logs a error and returns an error", func() {
					err = gc.SetInstallPackages()
					Expect(err).NotTo(BeNil())

					Expect(buffer.String()).To(ContainSubstring("**ERROR** $GO15VENDOREXPERIMENT=0. To vendor your packages in vendor/"))
					Expect(buffer.String()).To(ContainSubstring("with go 1.6 this environment variable must unset or set to 1."))
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
						err = gc.SetInstallPackages()
						Expect(err).To(BeNil())

						Expect(gc.PackageList).To(Equal([]string{"a-package-name", "another-package"}))
					})
				})

				Context("packages are already vendored", func() {
					BeforeEach(func() {
						err = os.MkdirAll(filepath.Join(mainPackagePath, "vendor", "another-package"), 0755)
						Expect(err).To(BeNil())
					})
					It("handles the vendoring correctly", func() {
						err = gc.SetInstallPackages()
						Expect(err).To(BeNil())

						Expect(gc.PackageList).To(Equal([]string{"a-package-name", filepath.Join(mainPackageName, "vendor", "another-package")}))
					})
				})

			})
			Context("GO_INSTALL_PACKAGE_SPEC is not set", func() {
				It("returns default", func() {
					err = gc.SetInstallPackages()
					Expect(err).To(BeNil())
					Expect(gc.PackageList).To(Equal([]string{"."}))
				})

				It("logs a warning that it is using the default", func() {
					err := gc.SetInstallPackages()
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
					godep = g.Godep{WorkspaceExists: true}
				})

				It("wraps the install command with godep", func() {
					gomock.InOrder(
						mockCommandRunner.EXPECT().SetDir(mainPackagePath),
						mockCommandRunner.EXPECT().Run("godep", "go", "install", "-v", "-a=1", "-b=2", "first", "second").Return(nil),
						mockCommandRunner.EXPECT().SetDir(""),
					)

					err = gc.CompileApp()
					Expect(err).To(BeNil())

					Expect(buffer.String()).To(ContainSubstring("-----> Running: godep go install -v -a=1 -b=2 first second"))
				})
			})

			Context("godeps workspace dir does not exist", func() {
				BeforeEach(func() {
					godep = g.Godep{WorkspaceExists: false}
				})

				Context("vendor experiment is true", func() {
					BeforeEach(func() {
						vendorExperiment = true
					})

					It("does not wrap the install command with godep", func() {
						gomock.InOrder(
							mockCommandRunner.EXPECT().SetDir(mainPackagePath),
							mockCommandRunner.EXPECT().Run("go", "install", "-v", "-a=1", "-b=2", "first", "second").Return(nil),
							mockCommandRunner.EXPECT().SetDir(""),
						)

						err = gc.CompileApp()
						Expect(err).To(BeNil())

						Expect(buffer.String()).To(ContainSubstring("-----> Running: go install -v -a=1 -b=2 first second"))
					})

				})

				Context("vendor experiment is false", func() {
					BeforeEach(func() {
						vendorExperiment = false
					})

					It("wraps the command with godep", func() {
						gomock.InOrder(
							mockCommandRunner.EXPECT().SetDir(mainPackagePath),
							mockCommandRunner.EXPECT().Run("godep", "go", "install", "-v", "-a=1", "-b=2", "first", "second").Return(nil),
							mockCommandRunner.EXPECT().SetDir(""),
						)

						err = gc.CompileApp()
						Expect(err).To(BeNil())

						Expect(buffer.String()).To(ContainSubstring("-----> Running: godep go install -v -a=1 -b=2 first second"))
					})
				})
			})
		})

		Context("the tool is glide", func() {
			BeforeEach(func() {
				vendorTool = "glide"
			})
			It("logs and runs the install command it is going to run", func() {
				gomock.InOrder(
					mockCommandRunner.EXPECT().SetDir(mainPackagePath),
					mockCommandRunner.EXPECT().Run("go", "install", "-v", "-a=1", "-b=2", "first", "second").Return(nil),
					mockCommandRunner.EXPECT().SetDir(""),
				)

				err = gc.CompileApp()
				Expect(err).To(BeNil())

				Expect(buffer.String()).To(ContainSubstring("-----> Running: go install -v -a=1 -b=2 first second"))
			})
		})

		Context("the tool is go_nativevendoring", func() {
			BeforeEach(func() {
				vendorTool = "go_nativevendoring"
			})

			It("logs and runs the install command it is going to run", func() {
				gomock.InOrder(
					mockCommandRunner.EXPECT().SetDir(mainPackagePath),
					mockCommandRunner.EXPECT().Run("go", "install", "-v", "-a=1", "-b=2", "first", "second").Return(nil),
					mockCommandRunner.EXPECT().SetDir(""),
				)

				err = gc.CompileApp()
				Expect(err).To(BeNil())

				Expect(buffer.String()).To(ContainSubstring("-----> Running: go install -v -a=1 -b=2 first second"))
			})
		})
	})

	Describe("CreateStartupEnvironment", func() {
		var tempDir string

		BeforeEach(func() {
			goVersion = "3.4.5"
			mainPackageName = "a-go-app"
			goPath = buildDir

			goDir := filepath.Join(cacheDir, "go"+goVersion, "go")
			err = os.MkdirAll(goDir, 0755)
			Expect(err).To(BeNil())

			tempDir, err = ioutil.TempDir("", "gobuildpack.releaseyml")
			Expect(err).To(BeNil())
		})

		It("writes the buildpack-release-step.yml file", func() {
			err = gc.CreateStartupEnvironment(tempDir)
			Expect(err).To(BeNil())

			contents, err := ioutil.ReadFile(filepath.Join(tempDir, "buildpack-release-step.yml"))
			Expect(err).To(BeNil())

			yaml := `---
default_process_types:
    web: a-go-app
`
			Expect(string(contents)).To(Equal(yaml))
		})

		It("writes the go.sh script to .profile.d", func() {
			err = gc.CreateStartupEnvironment(tempDir)
			Expect(err).To(BeNil())

			contents, err := ioutil.ReadFile(filepath.Join(buildDir, ".profile.d", "go.sh"))
			Expect(err).To(BeNil())

			Expect(string(contents)).To(Equal("PATH=$PATH:$HOME/bin"))
		})

		Context("GO_INSTALL_TOOLS_IN_IMAGE is not set", func() {
			It("does not copy the go toolchain", func() {
				err = gc.CreateStartupEnvironment(tempDir)
				Expect(err).To(BeNil())

				Expect(filepath.Join(buildDir, ".cloudfoundry", "go")).NotTo(BeADirectory())
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

			It("copies the go toolchain", func() {
				err = gc.CreateStartupEnvironment(tempDir)
				Expect(err).To(BeNil())

				Expect(filepath.Join(buildDir, ".cloudfoundry", "go")).To(BeADirectory())
			})

			It("logs that the tool chain was copied", func() {
				err = gc.CreateStartupEnvironment(tempDir)
				Expect(err).To(BeNil())

				Expect(buffer.String()).To(ContainSubstring("-----> Copying go tool chain to $GOROOT=$HOME/.cloudfoundry/go"))
			})

			It("writes the goroot.sh script to .profile.d", func() {
				err = gc.CreateStartupEnvironment(tempDir)
				Expect(err).To(BeNil())

				contents, err := ioutil.ReadFile(filepath.Join(buildDir, ".profile.d", "goroot.sh"))
				Expect(err).To(BeNil())

				Expect(string(contents)).To(ContainSubstring("export GOROOT=$HOME/.cloudfoundry/go"))
				Expect(string(contents)).To(ContainSubstring("PATH=$PATH:$GOROOT/bin"))
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
				err = gc.CreateStartupEnvironment(tempDir)
				Expect(err).To(BeNil())

				Expect(buffer.String()).To(ContainSubstring("-----> Cleaning up $GOPATH/pkg"))
				Expect(filepath.Join(buildDir, "pkg")).ToNot(BeADirectory())
			})

			It("writes the zzgopath.sh script to .profile.d", func() {
				err = gc.CreateStartupEnvironment(tempDir)
				Expect(err).To(BeNil())

				contents, err := ioutil.ReadFile(filepath.Join(buildDir, ".profile.d", "zzgopath.sh"))
				Expect(err).To(BeNil())

				Expect(string(contents)).To(ContainSubstring("export GOPATH=$HOME"))
				Expect(string(contents)).To(ContainSubstring("cd $GOPATH/src/" + mainPackageName))
			})
		})
	})

})
