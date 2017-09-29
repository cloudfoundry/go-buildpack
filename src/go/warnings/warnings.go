package warnings

import (
	"fmt"
	"strings"
)

func GoVersionOverride(goVersion string) string {
	warning := `Using $GOVERSION override.
    $GOVERSION = %s

If this isn't what you want please run:
    cf unset-env <app> GOVERSION`

	return fmt.Sprintf(warning, goVersion)
}

func PackageSpecOverride(goPackageSpec []string) string {
	warning := `Using $GO_INSTALL_PACKAGE_SPEC override.
    $GO_INSTALL_PACKAGE_SPEC = %s

If this isn't what you want please run:
    cf unset-env <app> GO_INSTALL_PACKAGE_SPEC`

	return fmt.Sprintf(warning, strings.Join(goPackageSpec, " "))
}

func GodirError() string {
	errorMessage := `Deprecated, .godir file found! Please update to supported Godep or Glide dependency managers.
See https://github.com/tools/godep or https://github.com/Masterminds/glide for usage information.`

	return errorMessage
}

func GBError() string {
	errorMessage := `Cloud Foundry does not support the GB package manager.
We currently only support the Godep, Glide and dep package managers for go apps.
For support please file an issue: https://github.com/cloudfoundry/go-buildpack/issues`

	return errorMessage
}

func NoGOPACKAGENAMEerror() string {
	errorMessage := `To use go native vendoring set the $GOPACKAGENAME
environment variable to your app's package name`

	return errorMessage
}

func UnsupportedGO15VENDOREXPERIMENTerror() string {
	errorMessage := `GO15VENDOREXPERIMENT is set, but is not supported by go1.7 and later.
Run 'cf unset-env <app> GO15VENDOREXPERIMENT' before pushing again.`

	return errorMessage
}

func GodepsWorkspaceWarning() string {
	errorMessage := `Godeps/_workspace/src and vendor/ exist
code may not compile. Please convert all deps to vendor/`

	return errorMessage
}

func MustUseVendorError() string {
	errorMessage := `$GO15VENDOREXPERIMENT=0. To vendor your packages in vendor/
with go 1.6 this environment variable must unset or set to 1.`

	return errorMessage
}
