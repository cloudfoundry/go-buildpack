package golang

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/cloudfoundry/libbuildpack"
)

type Compiler struct {
	Compiler         *libbuildpack.Compiler
	VendorTool       string
	GoVersion        string
	MainPackageName  string
	GoPath           string
	PackageList      []string
	BuildFlags       []string
	Godep            Godep
	VendorExperiment bool
}

type Godep struct {
	ImportPath      string   `json:"ImportPath"`
	GoVersion       string   `json:"GoVersion"`
	Packages        []string `json:"Packages"`
	WorkspaceExists bool
}

func (gc *Compiler) SelectVendorTool() error {
	godepsJSONFile := filepath.Join(gc.Compiler.BuildDir, "Godeps", "Godeps.json")

	godirFile := filepath.Join(gc.Compiler.BuildDir, ".godir")
	isGodir, err := libbuildpack.FileExists(godirFile)
	if err != nil {
		return err
	}
	if isGodir {
		gc.Compiler.Log.Error(godirError())
		return errors.New(".godir deprecated")
	}

	isGB, err := gc.isGB()
	if err != nil {
		return err
	}
	if isGB {
		gc.Compiler.Log.Error(gbError())
		return errors.New("gb unsupported")
	}

	isGodep, err := libbuildpack.FileExists(godepsJSONFile)
	if err != nil {
		return err
	}
	if isGodep {
		gc.Compiler.Log.BeginStep("Checking Godeps/Godeps.json file")

		err = libbuildpack.NewJSON().Load(filepath.Join(gc.Compiler.BuildDir, "Godeps", "Godeps.json"), &gc.Godep)
		if err != nil {
			gc.Compiler.Log.Error("Bad Godeps/Godeps.json file")
			return err
		}

		gc.Godep.WorkspaceExists, err = libbuildpack.FileExists(filepath.Join(gc.Compiler.BuildDir, "Godeps", "_workspace", "src"))
		if err != nil {
			return err
		}

		gc.VendorTool = "godep"
		return nil
	}

	glideFile := filepath.Join(gc.Compiler.BuildDir, "glide.yaml")
	isGlide, err := libbuildpack.FileExists(glideFile)
	if err != nil {
		return err
	}
	if isGlide {
		gc.VendorTool = "glide"
		return nil
	}

	gc.VendorTool = "go_nativevendoring"
	return nil
}

func (gc *Compiler) InstallVendorTool(tmpDir string) error {
	if gc.VendorTool == "go_nativevendoring" {
		return nil
	}

	installDir := filepath.Join(tmpDir, gc.VendorTool)

	err := gc.Compiler.Manifest.InstallOnlyVersion(gc.VendorTool, installDir)
	if err != nil {
		return err
	}

	return addToPath(filepath.Join(installDir, "bin"))
}

func (gc *Compiler) SelectGoVersion() error {
	goVersion := os.Getenv("GOVERSION")

	if gc.VendorTool == "godep" {
		if goVersion != "" {
			gc.Compiler.Log.Warning(goVersionOverride(goVersion))
		} else {
			goVersion = gc.Godep.GoVersion
		}
	} else {
		if goVersion == "" {
			defaultGo, err := gc.Compiler.Manifest.DefaultVersion("go")
			if err != nil {
				return err
			}
			goVersion = fmt.Sprintf("go%s", defaultGo.Version)
		}
	}

	parsed, err := gc.ParseGoVersion(goVersion)
	if err != nil {
		return err
	}

	gc.GoVersion = parsed
	return nil
}

func (gc *Compiler) ParseGoVersion(partialGoVersion string) (string, error) {
	existingVersions := gc.Compiler.Manifest.AllDependencyVersions("go")

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

func (gc *Compiler) InstallGo() error {
	err := os.MkdirAll(filepath.Join(gc.Compiler.BuildDir, "bin"), 0755)
	if err != nil {
		return err
	}

	goInstallDir := gc.goInstallLocation()

	goInstalled, err := libbuildpack.FileExists(filepath.Join(goInstallDir, "go"))
	if err != nil {
		return err
	}

	if goInstalled {
		gc.Compiler.Log.BeginStep("Using go %s", gc.GoVersion)
	} else {
		err = gc.Compiler.ClearCache()
		if err != nil {
			return fmt.Errorf("clearing cache: %s", err.Error())
		}

		dep := libbuildpack.Dependency{Name: "go", Version: gc.GoVersion}
		err = gc.Compiler.Manifest.InstallDependency(dep, goInstallDir)
		if err != nil {
			return err
		}
	}

	err = os.Setenv("GOROOT", filepath.Join(goInstallDir, "go"))
	if err != nil {
		return err
	}

	return addToPath(filepath.Join(goInstallDir, "go", "bin"))
}

func (gc *Compiler) SetMainPackageName() error {
	switch gc.VendorTool {
	case "godep":
		gc.MainPackageName = gc.Godep.ImportPath

	case "glide":
		gc.Compiler.Command.SetDir(gc.Compiler.BuildDir)
		defer gc.Compiler.Command.SetDir("")

		stdout, err := gc.Compiler.Command.CaptureStdout("glide", "name")
		if err != nil {
			return err
		}
		gc.MainPackageName = strings.TrimSpace(stdout)

	case "go_nativevendoring":
		gc.MainPackageName = os.Getenv("GOPACKAGENAME")
		if gc.MainPackageName == "" {
			gc.Compiler.Log.Error(noGOPACKAGENAMEerror())
			return errors.New("GOPACKAGENAME unset")
		}

	default:
		return errors.New("invalid vendor tool")
	}
	return nil
}

func (gc *Compiler) CheckBinDirectory() error {
	fi, err := os.Stat(filepath.Join(gc.Compiler.BuildDir, "bin"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}

		return err
	}

	if fi.Mode().IsDir() {
		return nil
	}

	gc.Compiler.Log.Error("File bin exists and is not a directory.")
	return errors.New("invalid bin")
}

func (gc *Compiler) SetupGoPath() error {
	var skipMoveFile = map[string]bool{
		"Procfile": true,
		".profile": true,
		"src":      true,
	}

	var goPath string
	goPathInImage := os.Getenv("GO_SETUP_GOPATH_IN_IMAGE") == "true"

	if goPathInImage {
		goPath = gc.Compiler.BuildDir
	} else {
		tmpDir, err := ioutil.TempDir("", "gobuildpack.gopath")
		if err != nil {
			return err
		}
		goPath = filepath.Join(tmpDir, ".go")
	}

	err := os.Setenv("GOPATH", goPath)
	if err != nil {
		return err
	}
	gc.GoPath = goPath

	binDir := filepath.Join(gc.Compiler.BuildDir, "bin")
	err = os.MkdirAll(binDir, 0755)
	if err != nil {
		return err
	}

	packageDir := gc.mainPackagePath()
	err = os.MkdirAll(packageDir, 0755)
	if err != nil {
		return err
	}

	if goPathInImage {
		files, err := ioutil.ReadDir(gc.Compiler.BuildDir)
		if err != nil {
			return err
		}
		for _, f := range files {
			if !skipMoveFile[f.Name()] {
				src := filepath.Join(gc.Compiler.BuildDir, f.Name())
				dest := filepath.Join(packageDir, f.Name())

				err = os.Rename(src, dest)
				if err != nil {
					return err
				}
			}
		}
	} else {
		err = os.Setenv("GOBIN", binDir)
		if err != nil {
			return err
		}

		err = libbuildpack.CopyDirectory(gc.Compiler.BuildDir, packageDir)
		if err != nil {
			return err
		}
	}

	// unset git dir or it will mess with go install
	return os.Unsetenv("GIT_DIR")
}

func (gc *Compiler) SetBuildFlags() {
	flags := []string{"-tags", "cloudfoundry", "-buildmode", "pie"}

	if os.Getenv("GO_LINKER_SYMBOL") != "" && os.Getenv("GO_LINKER_VALUE") != "" {
		ld_flags := []string{"-ldflags", fmt.Sprintf("-X %s=%s", os.Getenv("GO_LINKER_SYMBOL"), os.Getenv("GO_LINKER_VALUE"))}

		flags = append(flags, ld_flags...)
	}

	gc.BuildFlags = flags
	return
}

func (gc *Compiler) RunGlideInstall() error {
	if gc.VendorTool != "glide" {
		return nil
	}

	vendorDirExists, err := libbuildpack.FileExists(filepath.Join(gc.mainPackagePath(), "vendor"))
	if err != nil {
		return err
	}
	runGlideInstall := true

	if vendorDirExists {
		numSubDirs := 0
		files, err := ioutil.ReadDir(filepath.Join(gc.mainPackagePath(), "vendor"))
		if err != nil {
			return err
		}
		for _, file := range files {
			if file.IsDir() {
				numSubDirs++
			}
		}

		if numSubDirs > 0 {
			runGlideInstall = false
		}
	}

	if runGlideInstall {
		gc.Compiler.Log.BeginStep("Fetching any unsaved dependencies (glide install)")
		gc.Compiler.Command.SetDir(gc.mainPackagePath())
		defer gc.Compiler.Command.SetDir("")

		err := gc.Compiler.Command.Run("glide", "install")
		if err != nil {
			return err
		}
	} else {
		gc.Compiler.Log.Info("Note: skipping (glide install) due to non-empty vendor directory.")
	}

	return nil
}

func (gc *Compiler) HandleVendorExperiment() error {
	gc.VendorExperiment = true

	if os.Getenv("GO15VENDOREXPERIMENT") == "" {
		return nil
	}

	ver, err := semver.NewVersion(gc.GoVersion)
	if err != nil {
		return err
	}

	go16 := ver.Major() == 1 && ver.Minor() == 6
	if !go16 {
		gc.Compiler.Log.Error(unsupportedGO15VENDOREXPERIMENTerror())
		return errors.New("unsupported GO15VENDOREXPERIMENT")
	}

	if os.Getenv("GO15VENDOREXPERIMENT") == "0" {
		gc.VendorExperiment = false
	}

	return nil
}

func (gc *Compiler) SetInstallPackages() error {
	var packages []string
	vendorDirExists, err := libbuildpack.FileExists(filepath.Join(gc.mainPackagePath(), "vendor"))
	if err != nil {
		return err
	}

	if os.Getenv("GO_INSTALL_PACKAGE_SPEC") != "" {
		packages = append(packages, strings.Split(os.Getenv("GO_INSTALL_PACKAGE_SPEC"), " ")...)
	}

	if gc.VendorTool == "godep" {
		useVendorDir := gc.VendorExperiment && !gc.Godep.WorkspaceExists

		if gc.Godep.WorkspaceExists && vendorDirExists {
			gc.Compiler.Log.Warning(godepsWorkspaceWarning())
		}

		if useVendorDir && !vendorDirExists {
			gc.Compiler.Log.Warning("vendor/ directory does not exist.")
		}

		if len(packages) != 0 {
			gc.Compiler.Log.Warning(packageSpecOverride(packages))
		} else if len(gc.Godep.Packages) != 0 {
			packages = gc.Godep.Packages
		} else {
			gc.Compiler.Log.Warning("Installing package '.' (default)")
			packages = append(packages, ".")
		}

		if useVendorDir {
			packages = gc.updatePackagesForVendor(packages)
		}
	} else {
		if !gc.VendorExperiment && gc.VendorTool == "go_nativevendoring" {
			gc.Compiler.Log.Error(mustUseVendorError())
			return errors.New("must use vendor/ for go native vendoring")
		}

		if len(packages) == 0 {
			packages = append(packages, ".")
			gc.Compiler.Log.Warning("Installing package '.' (default)")
		}

		packages = gc.updatePackagesForVendor(packages)
	}

	gc.PackageList = packages
	return nil
}

func (gc *Compiler) CompileApp() error {
	cmd := "go"
	args := []string{"install", "-v"}
	args = append(args, gc.BuildFlags...)
	args = append(args, gc.PackageList...)

	if gc.VendorTool == "godep" && (gc.Godep.WorkspaceExists || !gc.VendorExperiment) {
		args = append([]string{"go"}, args...)
		cmd = "godep"
	}

	gc.Compiler.Log.BeginStep(fmt.Sprintf("Running: %s %s", cmd, strings.Join(args, " ")))

	gc.Compiler.Command.SetDir(gc.mainPackagePath())
	defer gc.Compiler.Command.SetDir("")

	err := gc.Compiler.Command.Run(cmd, args...)
	if err != nil {
		return err
	}
	return nil
}

func (gc *Compiler) CreateStartupEnvironment(tempDir string) error {
	err := ioutil.WriteFile(filepath.Join(tempDir, "buildpack-release-step.yml"), []byte(releaseYAML(gc.MainPackageName)), 0644)
	if err != nil {
		gc.Compiler.Log.Error("Unable to write relase yml: %s", err.Error())
		return err
	}

	if os.Getenv("GO_INSTALL_TOOLS_IN_IMAGE") == "true" {
		gc.Compiler.Log.BeginStep("Copying go tool chain to $GOROOT=$HOME/.cloudfoundry/go")

		imageDir := filepath.Join(gc.Compiler.BuildDir, ".cloudfoundry")
		err = os.MkdirAll(imageDir, 0755)
		if err != nil {
			return err
		}
		err = libbuildpack.CopyDirectory(gc.goInstallLocation(), imageDir)
		if err != nil {
			return err
		}

		err = libbuildpack.WriteProfileD(gc.Compiler.BuildDir, "goroot.sh", goRootScript())
		if err != nil {
			return err
		}
	}

	if os.Getenv("GO_SETUP_GOPATH_IN_IMAGE") == "true" {
		gc.Compiler.Log.BeginStep("Cleaning up $GOPATH/pkg")
		err = os.RemoveAll(filepath.Join(gc.GoPath, "pkg"))
		if err != nil {
			return err
		}

		err = libbuildpack.WriteProfileD(gc.Compiler.BuildDir, "zzgopath.sh", zzGoPathScript(gc.MainPackageName))
		if err != nil {
			return err
		}
	}

	return libbuildpack.WriteProfileD(gc.Compiler.BuildDir, "go.sh", goScript())
}

func (gc *Compiler) mainPackagePath() string {
	return filepath.Join(gc.GoPath, "src", gc.MainPackageName)
}

func (gc *Compiler) goInstallLocation() string {
	return filepath.Join(gc.Compiler.CacheDir, "go"+gc.GoVersion)
}

func (gc *Compiler) updatePackagesForVendor(packages []string) []string {
	var newPackages []string

	for _, pkg := range packages {
		vendored, _ := libbuildpack.FileExists(filepath.Join(gc.mainPackagePath(), "vendor", pkg))
		if pkg == "." || !vendored {
			newPackages = append(newPackages, pkg)
		} else {
			newPackages = append(newPackages, filepath.Join(gc.MainPackageName, "vendor", pkg))
		}
	}

	return newPackages
}

func (gc *Compiler) isGB() (bool, error) {
	srcDir := filepath.Join(gc.Compiler.BuildDir, "src")
	srcDirAtAppRoot, err := libbuildpack.FileExists(srcDir)
	if err != nil {
		return false, err
	}

	if !srcDirAtAppRoot {
		return false, nil
	}

	files, err := ioutil.ReadDir(filepath.Join(gc.Compiler.BuildDir, "src"))
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

func addToPath(newPaths string) error {
	oldPath := os.Getenv("PATH")
	return os.Setenv("PATH", fmt.Sprintf("%s:%s", newPaths, oldPath))
}
