package supply_test

import (
	"golang"
	"io/ioutil"
	"os"
	"path/filepath"

	"bytes"

	"golang/supply"

	"github.com/cloudfoundry/libbuildpack"
	"github.com/cloudfoundry/libbuildpack/ansicleaner"

	"github.com/golang/mock/gomock"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

//go:generate mockgen -source=../vendor/github.com/cloudfoundry/libbuildpack/manifest.go --destination=mocks_manifest_test.go --package=supply_test --imports=.=github.com/cloudfoundry/libbuildpack

var _ = Describe("Supply", func() {
	var (
		buildDir     string
		depsDir      string
		depsIdx      string
		gs           *supply.Supplier
		logger       libbuildpack.Logger
		buffer       *bytes.Buffer
		err          error
		mockCtrl     *gomock.Controller
		mockManifest *MockManifest
		goVersion    string
		vendorTool   string
		godep        golang.Godep
	)

	BeforeEach(func() {
		buildDir, err = ioutil.TempDir("", "go-buildpack.build.")
		Expect(err).To(BeNil())

		depsDir, err = ioutil.TempDir("", "go-buildpack.deps.")
		Expect(err).To(BeNil())

		depsIdx = "04"

		err = os.MkdirAll(filepath.Join(depsDir, depsIdx), 0755)
		Expect(err).To(BeNil())

		buffer = new(bytes.Buffer)

		logger = libbuildpack.NewLogger()
		logger.SetOutput(ansicleaner.New(buffer))

		mockCtrl = gomock.NewController(GinkgoT())
		mockManifest = NewMockManifest(mockCtrl)
	})

	JustBeforeEach(func() {
		bps := &libbuildpack.Stager{
			BuildDir: buildDir,
			DepsDir:  depsDir,
			DepsIdx:  depsIdx,
			Manifest: mockManifest,
			Log:      logger,
		}

		gs = &supply.Supplier{
			Stager:     bps,
			GoVersion:  goVersion,
			VendorTool: vendorTool,
			Godep:      godep,
		}
	})

	AfterEach(func() {
		mockCtrl.Finish()

		err = os.RemoveAll(buildDir)
		Expect(err).To(BeNil())

		err = os.RemoveAll(depsDir)
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
					err = gs.SelectVendorTool()
					Expect(err).To(BeNil())

					Expect(gs.VendorTool).To(Equal("godep"))
				})
				It("logs that it is checking the Godeps.json file", func() {
					err = gs.SelectVendorTool()
					Expect(err).To(BeNil())

					Expect(buffer.String()).To(ContainSubstring("-----> Checking Godeps/Godeps.json file"))
				})
				It("stores the Godep info in the supplier struct", func() {
					err = gs.SelectVendorTool()
					Expect(err).To(BeNil())

					Expect(gs.Godep.ImportPath).To(Equal("go-online"))
					Expect(gs.Godep.GoVersion).To(Equal("go1.6"))

					var empty []string
					Expect(gs.Godep.Packages).To(Equal(empty))
				})

				Context("godeps workspace exists", func() {
					BeforeEach(func() {
						err = os.MkdirAll(filepath.Join(buildDir, "Godeps", "_workspace", "src"), 0755)
						Expect(err).To(BeNil())
					})

					It("sets Godep.WorkspaceExists to true", func() {
						err = gs.SelectVendorTool()
						Expect(err).To(BeNil())

						Expect(gs.Godep.WorkspaceExists).To(BeTrue())
					})
				})

				Context("godeps workspace does not exist", func() {
					It("sets Godep.WorkspaceExists to false", func() {
						err = gs.SelectVendorTool()
						Expect(err).To(BeNil())

						Expect(godep.WorkspaceExists).To(BeFalse())
					})
				})
			})

			Context("bad Godeps.json file", func() {
				BeforeEach(func() {
					godepsJsonContents = "not actually JSON"
				})

				It("logs that the Godeps.json file is invalid and returns an error", func() {
					err = gs.SelectVendorTool()
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
				err = gs.SelectVendorTool()
				Expect(err).NotTo(BeNil())

				Expect(buffer.String()).To(ContainSubstring("**ERROR** Deprecated, .godir file found! Please update to supported Godep or Glide dependency managers."))
				Expect(buffer.String()).To(ContainSubstring("See https://github.com/tools/godep or https://github.com/Masterminds/glide for usage information."))
			})
		})

		Context("there is a glide.yaml file", func() {
			BeforeEach(func() {
				err = ioutil.WriteFile(filepath.Join(buildDir, "glide.yaml"), []byte("xxx"), 0644)
				Expect(err).To(BeNil())
			})

			It("sets the tool to glide", func() {
				err = gs.SelectVendorTool()
				Expect(err).To(BeNil())

				Expect(gs.VendorTool).To(Equal("glide"))
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
				err = gs.SelectVendorTool()
				Expect(err).NotTo(BeNil())

				Expect(buffer.String()).To(ContainSubstring("**ERROR** Cloud Foundry does not support the GB package manager."))
				Expect(buffer.String()).To(ContainSubstring("We currently only support the Godep and Glide package managers for go apps"))
				Expect(buffer.String()).To(ContainSubstring("For support please file an issue: https://github.com/cloudfoundry/go-buildpack/issues"))

			})
		})

		Context("none of the above", func() {
			It("sets the tool to go_nativevendoring", func() {
				err = gs.SelectVendorTool()
				Expect(err).To(BeNil())

				Expect(gs.VendorTool).To(Equal("go_nativevendoring"))
			})
		})
	})

	Describe("InstallVendorTools", func() {
		It("installs godep + glide to the depDir, creating a symlink in <depDir>/bin", func() {
			godepInstallDir := filepath.Join(depsDir, depsIdx, "godep")
			glideInstallDir := filepath.Join(depsDir, depsIdx, "glide")

			mockManifest.EXPECT().InstallOnlyVersion("godep", godepInstallDir).Return(nil)
			mockManifest.EXPECT().InstallOnlyVersion("glide", glideInstallDir).Return(nil)

			err = gs.InstallVendorTools()
			Expect(err).To(BeNil())

			link, err := os.Readlink(filepath.Join(depsDir, depsIdx, "bin", "godep"))
			Expect(err).To(BeNil())

			Expect(link).To(Equal("../godep/bin/godep"))

			link, err = os.Readlink(filepath.Join(depsDir, depsIdx, "bin", "glide"))
			Expect(err).To(BeNil())

			Expect(link).To(Equal("../glide/bin/glide"))
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
				godep = golang.Godep{ImportPath: "go-online", GoVersion: "go1.6"}
			})

			Context("GOVERSION not set", func() {
				It("sets the go version from Godeps.json", func() {
					err = gs.SelectGoVersion()
					Expect(err).To(BeNil())

					Expect(gs.GoVersion).To(Equal("1.6.4"))
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
					err = gs.SelectGoVersion()
					Expect(err).To(BeNil())

					Expect(gs.GoVersion).To(Equal("34.34.0"))
					Expect(buffer.String()).To(ContainSubstring("**WARNING** Using $GOVERSION override.\n"))
					Expect(buffer.String()).To(ContainSubstring("    $GOVERSION = go34.34\n"))
					Expect(buffer.String()).To(ContainSubstring("If this isn't what you want please run:\n"))
					Expect(buffer.String()).To(ContainSubstring("    cf unset-env <app> GOVERSION"))
				})
			})
		})

		Context("glide or go_nativevendoring", func() {
			Context("GOVERSION is notset", func() {
				BeforeEach(func() {
					vendorTool = "glide"
					dep := libbuildpack.Dependency{Name: "go", Version: "1.14.3"}
					mockManifest.EXPECT().DefaultVersion("go").Return(dep, nil)
				})

				It("sets the go version to the default from the manifest.yml", func() {
					err = gs.SelectGoVersion()
					Expect(err).To(BeNil())

					Expect(gs.GoVersion).To(Equal("1.14.3"))
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
					err = gs.SelectGoVersion()
					Expect(err).To(BeNil())

					Expect(gs.GoVersion).To(Equal("34.34.0"))
				})
			})
		})
	})

	Describe("InstallGo", func() {
		var (
			goInstallDir string
			dep          libbuildpack.Dependency
		)

		BeforeEach(func() {
			goVersion = "1.3.4"
			goInstallDir = filepath.Join(depsDir, depsIdx, "go1.3.4")
			dep = libbuildpack.Dependency{Name: "go", Version: "1.3.4"}
			err = os.MkdirAll(filepath.Join(goInstallDir, "go"), 0755)
			Expect(err).To(BeNil())
			mockManifest.EXPECT().InstallDependency(dep, goInstallDir).Return(nil)
		})

		It("Write GOROOT to envfile", func() {
			err = gs.InstallGo()
			Expect(err).To(BeNil())

			contents, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "env", "GOROOT"))
			Expect(err).To(BeNil())
			Expect(string(contents)).To(Equal(filepath.Join(goInstallDir, "go")))
		})

		It("installs go to the depDir, creating a symlink in <depDir>/bin", func() {
			err = gs.InstallGo()
			Expect(err).To(BeNil())

			link, err := os.Readlink(filepath.Join(depsDir, depsIdx, "bin", "go"))
			Expect(err).To(BeNil())

			Expect(link).To(Equal("../go1.3.4/go/bin/go"))

		})
	})

	Describe("ConfigureFinalizeEnv", func() {
		BeforeEach(func() {
			goVersion = "1.3.4"
		})

		Context("The vendor tool is Godep", func() {
			BeforeEach(func() {
				vendorTool = "godep"
				godep = golang.Godep{
					ImportPath:      "an-import-path",
					GoVersion:       "go1.3",
					Packages:        []string{"package1", "package2"},
					WorkspaceExists: true,
				}
			})

			It("Writes the go version to supply_GoVersion envfile", func() {
				err = gs.ConfigureFinalizeEnv()
				Expect(err).To(BeNil())

				contents, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "env", "supply_GoVersion"))
				Expect(err).To(BeNil())
				Expect(string(contents)).To(Equal("1.3.4"))
			})

			It("Writes the vendor tool to supply_VendorTool envfile", func() {
				err = gs.ConfigureFinalizeEnv()
				Expect(err).To(BeNil())

				contents, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "env", "supply_VendorTool"))
				Expect(err).To(BeNil())
				Expect(string(contents)).To(Equal("godep"))
			})

			It("Writes the godep info to supply_Godep envfile", func() {
				godepsJsonContents := `{"ImportPath":"an-import-path","GoVersion":"go1.3","Packages":["package1","package2"],"WorkspaceExists":true}`
				err = gs.ConfigureFinalizeEnv()
				Expect(err).To(BeNil())

				contents, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "env", "supply_Godep"))
				Expect(err).To(BeNil())
				Expect(string(contents)).To(Equal(godepsJsonContents))
			})
		})

		Context("The vendor tool is not Godep", func() {
			BeforeEach(func() {
				vendorTool = "glide"
			})

			It("Writes the go version to supply_GoVersion envfile", func() {
				err = gs.ConfigureFinalizeEnv()
				Expect(err).To(BeNil())

				contents, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "env", "supply_GoVersion"))
				Expect(err).To(BeNil())
				Expect(string(contents)).To(Equal("1.3.4"))
			})

			It("Writes the vendor tool to supply_VendorTool envfile", func() {
				err = gs.ConfigureFinalizeEnv()
				Expect(err).To(BeNil())

				contents, err := ioutil.ReadFile(filepath.Join(depsDir, depsIdx, "env", "supply_VendorTool"))
				Expect(err).To(BeNil())
				Expect(string(contents)).To(Equal("glide"))
			})
		})
	})
})
