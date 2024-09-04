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
		err error
	)

	BeforeEach(func() {
		Expect(err).NotTo(HaveOccurred())
	})

	bratshelper.UnbuiltBuildpack(DEP, CopyBrats)
	bratshelper.DeployAppWithExecutableProfileScript(DEP, CopyBrats)

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
