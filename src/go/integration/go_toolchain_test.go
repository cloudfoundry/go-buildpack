package integration_test

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testGoToolchain(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect     = NewWithT(t).Expect
		Eventually = NewWithT(t).Eventually

		app *cutlass.App
	)

	it.After(func() {
		app = DestroyApp(t, app)
	})

	context("when GO_SETUP_GOPATH_IN_IMAGE is specified", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "go_toolchain", "gopath"))
			app.SetEnv("GO_SETUP_GOPATH_IN_IMAGE", "true")
		})

		it("sets up the $HOME as the $GOPATH", func() {
			PushAppAndConfirm(t, app)

			Expect(app.GetBody("/")).To(ContainSubstring("GOPATH: /home/vcap/app"))
		})
	})

	context("when GO_INSTALL_TOOLS_IN_IMAGE is specified", func() {
		it.Before(func() {
			app = cutlass.New(filepath.Join(settings.FixturesPath, "go_toolchain", "toolchain"))
			app.SetEnv("GO_INSTALL_TOOLS_IN_IMAGE", "true")
			app.Disk = "1G"
		})

		it("keeps the go toolchain in the droplet", func() {
			PushAppAndConfirm(t, app)

			Expect(app.GetBody("/")).To(MatchRegexp(`go version go1\.\d+(\.\d+)? linux/amd64`))
		})

		context("when running a task", func() {
			it("can execute the go toolchain", func() {
				PushAppAndConfirm(t, app)

				_, err := app.RunTask(`echo "RUNNING A TASK: $(go version)"`)
				Expect(err).ToNot(HaveOccurred())

				Eventually(func() string { return app.Stdout.String() }, 1*time.Minute).Should(MatchRegexp(`RUNNING A TASK: go version go1\.\d+(\.\d+)? linux/amd64`))
			})
		})
	})
}
