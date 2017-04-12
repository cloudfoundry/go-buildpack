package supply

import (
	"encoding/json"
	"errors"
	"fmt"
	"golang"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudfoundry/libbuildpack"
)

type Supplier struct {
	Stager     *libbuildpack.Stager
	VendorTool string
	GoVersion  string
	Godep      golang.Godep
}

func Run(gs *Supplier) error {
	if err := gs.SelectVendorTool(); err != nil {
		gs.Stager.Log.Error("Unable to select Go vendor tool: %s", err.Error())
		return err
	}

	if err := gs.InstallVendorTools(); err != nil {
		gs.Stager.Log.Error("Unable to install vendor tools", err.Error())
		return err
	}

	if err := gs.SelectGoVersion(); err != nil {
		gs.Stager.Log.Error("Unable to determine Go version to install: %s", err.Error())
		return err
	}

	if err := gs.InstallGo(); err != nil {
		gs.Stager.Log.Error("Error installing Go: %s", err.Error())
		return err
	}

	if err := gs.ConfigureFinalizeEnv(); err != nil {
		gs.Stager.Log.Error("Error writing environment vars: %s", err.Error())
		return nil
	}

	if err := gs.Stager.WriteConfigYml(); err != nil {
		gs.Stager.Log.Error("Error writing config.yml: %s", err.Error())
		return err
	}

	return nil
}

func (gs *Supplier) SelectVendorTool() error {
	godepsJSONFile := filepath.Join(gs.Stager.BuildDir, "Godeps", "Godeps.json")

	godirFile := filepath.Join(gs.Stager.BuildDir, ".godir")
	isGodir, err := libbuildpack.FileExists(godirFile)
	if err != nil {
		return err
	}
	if isGodir {
		gs.Stager.Log.Error(golang.GodirError())
		return errors.New(".godir deprecated")
	}

	isGoPath, err := gs.isGoPath()
	if err != nil {
		return err
	}
	if isGoPath {
		gs.Stager.Log.Error(golang.GBError())
		return errors.New("gb unsupported")
	}

	isGodep, err := libbuildpack.FileExists(godepsJSONFile)
	if err != nil {
		return err
	}
	if isGodep {
		gs.Stager.Log.BeginStep("Checking Godeps/Godeps.json file")

		err = libbuildpack.NewJSON().Load(filepath.Join(gs.Stager.BuildDir, "Godeps", "Godeps.json"), &gs.Godep)
		if err != nil {
			gs.Stager.Log.Error("Bad Godeps/Godeps.json file")
			return err
		}

		gs.Godep.WorkspaceExists, err = libbuildpack.FileExists(filepath.Join(gs.Stager.BuildDir, "Godeps", "_workspace", "src"))
		if err != nil {
			return err
		}

		gs.VendorTool = "godep"
		return nil
	}

	glideFile := filepath.Join(gs.Stager.BuildDir, "glide.yaml")
	isGlide, err := libbuildpack.FileExists(glideFile)
	if err != nil {
		return err
	}
	if isGlide {
		gs.VendorTool = "glide"
		return nil
	}

	gs.VendorTool = "go_nativevendoring"
	return nil
}

func (gs *Supplier) InstallVendorTools() error {
	tools := []string{"godep", "glide"}

	for _, tool := range tools {
		installDir := filepath.Join(gs.Stager.DepDir(), tool)
		if err := gs.Stager.Manifest.InstallOnlyVersion(tool, installDir); err != nil {
			return err
		}

		if err := gs.Stager.AddBinDependencyLink(filepath.Join(installDir, "bin", tool), tool); err != nil {
			return err
		}
	}

	return nil
}

func (gs *Supplier) SelectGoVersion() error {
	goVersion := os.Getenv("GOVERSION")

	if gs.VendorTool == "godep" {
		if goVersion != "" {
			gs.Stager.Log.Warning(golang.GoVersionOverride(goVersion))
		} else {
			goVersion = gs.Godep.GoVersion
		}
	} else {
		if goVersion == "" {
			defaultGo, err := gs.Stager.Manifest.DefaultVersion("go")
			if err != nil {
				return err
			}
			goVersion = fmt.Sprintf("go%s", defaultGo.Version)
		}
	}

	parsed, err := gs.parseGoVersion(goVersion)
	if err != nil {
		return err
	}

	gs.GoVersion = parsed
	return nil
}

func (gs *Supplier) InstallGo() error {
	goInstallDir := filepath.Join(gs.Stager.DepDir(), "go"+gs.GoVersion)

	dep := libbuildpack.Dependency{Name: "go", Version: gs.GoVersion}
	if err := gs.Stager.Manifest.InstallDependency(dep, goInstallDir); err != nil {
		return err
	}

	if err := gs.Stager.AddBinDependencyLink(filepath.Join(goInstallDir, "go", "bin", "go"), "go"); err != nil {
		return err
	}

	return gs.Stager.WriteEnvFile("GOROOT", filepath.Join(goInstallDir, "go"))
}

func (gs *Supplier) ConfigureFinalizeEnv() error {
	if err := gs.Stager.WriteEnvFile("supply_GoVersion", gs.GoVersion); err != nil {
		return err
	}

	if err := gs.Stager.WriteEnvFile("supply_VendorTool", gs.VendorTool); err != nil {
		return err
	}

	if gs.VendorTool == "godep" {
		data, err := json.Marshal(&gs.Godep)
		if err != nil {
			return err
		}

		if err := gs.Stager.WriteEnvFile("supply_Godep", string(data)); err != nil {
			return err
		}
	}

	return nil
}

func (gs *Supplier) parseGoVersion(partialGoVersion string) (string, error) {
	existingVersions := gs.Stager.Manifest.AllDependencyVersions("go")

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
	srcDir := filepath.Join(gs.Stager.BuildDir, "src")
	srcDirAtAppRoot, err := libbuildpack.FileExists(srcDir)
	if err != nil {
		return false, err
	}

	if !srcDirAtAppRoot {
		return false, nil
	}

	files, err := ioutil.ReadDir(filepath.Join(gs.Stager.BuildDir, "src"))
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
