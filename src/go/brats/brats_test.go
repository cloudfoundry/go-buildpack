package brats_test

import (
	"github.com/cloudfoundry/libbuildpack/bratshelper"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Go buildpack", func() {
	bratshelper.UnbuiltBuildpack("go", CopyBrats)
	bratshelper.DeployingAnAppWithAnUpdatedVersionOfTheSameBuildpack(CopyBrats)
	bratshelper.StagingWithBuildpackThatSetsEOL("go", func(_ string) *cutlass.App {
		return CopyBrats("1.8.7")
	})
	bratshelper.StagingWithADepThatIsNotTheLatest("go", func(_ string) *cutlass.App {
		return CopyBrats("1.10.3")
	})

	bratshelper.StagingWithCustomBuildpackWithCredentialsInDependencies(CopyBrats)
	bratshelper.DeployAppWithExecutableProfileScript("go", CopyBrats)
	bratshelper.DeployAnAppWithSensitiveEnvironmentVariables(CopyBrats)

	bratshelper.ForAllSupportedVersions("go", CopyBrats, func(goVersion string, app *cutlass.App) {
		PushApp(app)

		By("installs the correct go version", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("Installing go " + goVersion))
		})
		By("runs a simple webserver", func() {
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})
	})
})
