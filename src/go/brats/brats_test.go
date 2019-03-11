package brats_test

import (
	"github.com/cloudfoundry/libbuildpack/bratshelper"
	"github.com/cloudfoundry/libbuildpack/cutlass"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Go buildpack", func() {
	const (
		DEP = "go"
	)

	var (
		bpDir				string
		err					error
	)

	BeforeEach(func() {
		bpDir, err = cutlass.FindRoot()
		Expect(err).NotTo(HaveOccurred())
	})

	bratshelper.UnbuiltBuildpack(DEP, CopyBrats)
	bratshelper.DeployingAnAppWithAnUpdatedVersionOfTheSameBuildpack(CopyBrats)
	bratshelper.StagingWithBuildpackThatSetsEOL(DEP, func(_ string) *cutlass.App {
		return CopyBrats(GetOldestVersion(DEP, bpDir))
	})
	bratshelper.StagingWithADepThatIsNotTheLatest(DEP, func(_ string) *cutlass.App {
		return CopyBrats(GetOlderDepThatDiffersInPatch(DEP, bpDir))
	})

	bratshelper.StagingWithCustomBuildpackWithCredentialsInDependencies(CopyBrats)
	bratshelper.DeployAppWithExecutableProfileScript(DEP, CopyBrats)
	bratshelper.DeployAnAppWithSensitiveEnvironmentVariables(CopyBrats)

	bratshelper.ForAllSupportedVersions(DEP, CopyBrats, func(goVersion string, app *cutlass.App) {
		PushApp(app)

		By("installs the correct go version", func() {
			Expect(app.Stdout.String()).To(ContainSubstring("Installing go " + goVersion))
		})
		By("runs a simple webserver", func() {
			Expect(app.GetBody("/")).To(ContainSubstring("Hello World!"))
		})
	})
})
