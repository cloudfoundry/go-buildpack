package supply

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/cloudfoundry/go-buildpack/src/go/data"
	"github.com/cloudfoundry/go-buildpack/src/go/godep"
	"github.com/cloudfoundry/go-buildpack/src/go/gomod"
	"github.com/cloudfoundry/go-buildpack/src/go/warnings"
	"github.com/cloudfoundry/libbuildpack"
)

type Manifest interface {
	AllDependencyVersions(string) []string
	DefaultVersion(string) (libbuildpack.Dependency, error)
}

type Installer interface {
	InstallDependency(libbuildpack.Dependency, string) error
	InstallOnlyVersion(string, string) error
}

type Stager interface {
	AddBinDependencyLink(string, string) error
	BuildDir() string
	DepDir() string
	DepsIdx() string
	WriteConfigYml(interface{}) error
	WriteEnvFile(string, string) error
	WriteProfileD(string, string) error
}

type Supplier struct {
	Stager     Stager
	Manifest   Manifest
	Installer  Installer
	Log        *libbuildpack.Logger
	VendorTool string
	GoVersion  string
	Godep      godep.Godep
	GoMod      gomod.GoMod
}

func Run(gs *Supplier) error {
	if err := gs.SelectVendorTool(); err != nil {
		gs.Log.Error("Unable to select Go vendor tool: %s", err.Error())
		return err
	}

	if err := gs.InstallVendorTools(); err != nil {
		gs.Log.Error("Unable to install vendor tools: %s", err.Error())
		return err
	}

	if err := gs.SelectGoVersion(); err != nil {
		gs.Log.Error("Unable to determine Go version to install: %s", err.Error())
		return err
	}

	if err := gs.InstallGo(); err != nil {
		gs.Log.Error("Error installing Go: %s", err.Error())
		return err
	}

	if err := gs.WriteGoRootToProfileD(); err != nil {
		gs.Log.Error("Error writing GOROOT to profile.d: %s", err.Error())
		return err
	}

	if err := gs.WriteConfigYml(); err != nil {
		gs.Log.Error("Error writing config.yml: %s", err.Error())
		return err
	}

	return nil
}

func (gs *Supplier) SelectVendorTool() error {
	godepsJSONFile := filepath.Join(gs.Stager.BuildDir(), "Godeps", "Godeps.json")

	godirFile := filepath.Join(gs.Stager.BuildDir(), ".godir")
	isGodir, err := libbuildpack.FileExists(godirFile)
	if err != nil {
		return err
	}
	if isGodir {
		gs.Log.Error(warnings.GodirError())
		return errors.New(".godir deprecated")
	}

	isGoMod, err := libbuildpack.FileExists(filepath.Join(gs.Stager.BuildDir(), "go.mod"))
	if err != nil {
		return err
	}
	if isGoMod {
		gs.Stager.WriteEnvFile("GO111MODULE", "on")
		gs.VendorTool = "gomod"

		// Check if the go.mod file contains a Go version
		gs.Log.BeginStep("Checking Go version in go.mod file")
		if err := gs.GoMod.Load(filepath.Join(gs.Stager.BuildDir(), "go.mod")); err != nil {
			gs.Log.Error("Unable to load go version from go.mod: %s", err)
		}

		if gs.GoMod.GoVersion != "" {
			gs.Log.Info("Go version found in go.mod")
		} else {
			gs.Log.Info("No Go version found in go.mod")
		}

		return nil
	}

	isGodep, err := libbuildpack.FileExists(godepsJSONFile)
	if err != nil {
		return err
	}
	if isGodep {
		gs.Log.BeginStep("Checking Godeps/Godeps.json file")

		err = libbuildpack.NewJSON().Load(filepath.Join(gs.Stager.BuildDir(), "Godeps", "Godeps.json"), &gs.Godep)
		if err != nil {
			gs.Log.Error("Bad Godeps/Godeps.json file")
			return err
		}

		gs.Godep.WorkspaceExists, err = libbuildpack.FileExists(filepath.Join(gs.Stager.BuildDir(), "Godeps", "_workspace", "src"))
		if err != nil {
			return err
		}

		gs.Stager.WriteEnvFile("GO111MODULE", "auto")
		gs.VendorTool = "godep"
		return nil
	}

	glideFile := filepath.Join(gs.Stager.BuildDir(), "glide.yaml")
	isGlide, err := libbuildpack.FileExists(glideFile)
	if err != nil {
		return err
	}
	if isGlide {
		gs.VendorTool = "glide"
		return nil
	}

	depFile := filepath.Join(gs.Stager.BuildDir(), "Gopkg.toml")
	isDep, err := libbuildpack.FileExists(depFile)
	if err != nil {
		return err
	}
	if isDep {
		gs.VendorTool = "dep"
		return nil
	}

	gs.VendorTool = "go_nativevendoring"
	return nil
}

func (gs *Supplier) WriteGoRootToProfileD() error {
	goRuntimeLocation := filepath.Join("$DEPS_DIR", gs.Stager.DepsIdx(), "go"+gs.GoVersion)
	if err := gs.Stager.WriteProfileD("goroot.sh", data.GoRootScript(goRuntimeLocation)); err != nil {
		return err
	}
	return nil
}

func (gs *Supplier) InstallVendorTools() error {
	tools := []string{"godep", "glide", "dep"}

	for _, tool := range tools {
		installDir := filepath.Join(gs.Stager.DepDir(), tool)
		if err := gs.Installer.InstallOnlyVersion(tool, installDir); err != nil {
			return err
		}

		if err := gs.Stager.AddBinDependencyLink(filepath.Join(installDir, "bin", tool), tool); err != nil {
			return err
		}
	}

	return nil
}

func (gs *Supplier) SelectGoVersion() error {
	goVersion, err := resolveGoVersion(gs)
	if err != nil {
		return err
	}

	parsed, err := gs.parseGoVersion(goVersion)
	if err != nil {
		return err
	}

	gs.GoVersion = parsed
	if gs.VendorTool == "gomod" {
		goVersion, err := semver.NewVersion(gs.GoVersion)
		if err != nil {
			return err
		}

		goModConstraint, err := semver.NewConstraint(">= 1.11.0")
		if err != nil {
			return err
		}

		if !goModConstraint.Check(goVersion) {
			return fmt.Errorf("go version %s does not support go modules", gs.GoVersion)
		}

		if exists, err := libbuildpack.FileExists(filepath.Join(gs.Stager.BuildDir(), "vendor")); err != nil {
			return err
		} else if exists {
			gs.Stager.WriteEnvFile("GOFLAGS", "-mod=vendor")
		}
	}

	return nil
}

func resolveGoVersion(gs *Supplier) (string, error) {
	goVersion := os.Getenv("GOVERSION")

	if goVersion != "" {
		gs.Log.Warning(warnings.GoVersionOverride(goVersion))
		return goVersion, nil
	}

	if gs.VendorTool == "godep" {
		return gs.Godep.GoVersion, nil
	}

	if gs.VendorTool == "gomod" && gs.GoMod.GoVersion != "" {
		return gs.GoMod.GoVersion, nil
	}

	defaultGo, err := gs.Manifest.DefaultVersion("go")
	if err != nil {
		return "", err
	}
	goVersion = fmt.Sprintf("go%s", defaultGo.Version)

	return goVersion, nil
}

func (gs *Supplier) InstallGo() error {
	goInstallDir := filepath.Join(gs.Stager.DepDir(), "go"+gs.GoVersion)

	dep := libbuildpack.Dependency{Name: "go", Version: gs.GoVersion}
	if err := gs.Installer.InstallDependency(dep, goInstallDir); err != nil {
		return err
	}

	if err := gs.Stager.AddBinDependencyLink(filepath.Join(goInstallDir, "bin", "go"), "go"); err != nil {
		return err
	}

	if err := gs.Stager.WriteEnvFile("GO111MODULE", "auto"); err != nil {
		return err
	}

	return gs.Stager.WriteEnvFile("GOROOT", filepath.Join(goInstallDir))
}

func (gs *Supplier) WriteConfigYml() error {
	config := map[string]string{
		"GoVersion":  gs.GoVersion,
		"VendorTool": gs.VendorTool,
	}

	if gs.VendorTool == "godep" {
		data, err := json.Marshal(&gs.Godep)
		if err != nil {
			return err
		}

		config["Godep"] = string(data)
	}

	return gs.Stager.WriteConfigYml(config)
}

func (gs *Supplier) parseGoVersion(partialGoVersion string) (string, error) {
	existingVersions := gs.Manifest.AllDependencyVersions("go")

	if len(strings.Split(partialGoVersion, ".")) < 3 {
		partialGoVersion += ".x"
	}

	strippedGoVersion := strings.TrimLeft(partialGoVersion, "go")

	expandedVer, err := libbuildpack.FindMatchingVersion(strippedGoVersion, existingVersions)
	if err != nil {
		return "", err
	}

	return expandedVer, nil
}

func (gs *Supplier) isGoPath() (bool, error) {
	srcDir := filepath.Join(gs.Stager.BuildDir(), "src")
	srcDirAtAppRoot, err := libbuildpack.FileExists(srcDir)
	if err != nil {
		return false, err
	}

	if !srcDirAtAppRoot {
		return false, nil
	}

	files, err := ioutil.ReadDir(filepath.Join(gs.Stager.BuildDir(), "src"))
	if err != nil {
		return false, err
	}

	for _, file := range files {
		if file.Mode().IsDir() {
			err = filepath.Walk(filepath.Join(srcDir, file.Name()), isGoFile)
			if err != nil {
				if err.Error() == "found Go file" {
					return true, nil
				}

				return false, err
			}
		}
	}

	return false, nil
}

func isGoFile(path string, info os.FileInfo, err error) error {
	if err != nil {
		return err
	}

	if strings.HasSuffix(path, ".go") {
		return errors.New("found Go file")
	}

	return nil
}
