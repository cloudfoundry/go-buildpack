package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("running supply ruby buildpack before the go buildpack", func() {
	var app *cutlass.App

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	Context("the app is pushed", func() {
		BeforeEach(func() {
			if ok, err := cutlass.ApiGreaterThan("2.65.1"); err != nil || !ok {
				Skip("API version does not have multi-buildpack support")
			}

			app = cutlass.New(filepath.Join(bpDir, "fixtures", "go_calls_ruby"))
			app.Buildpacks = []string{
				"https://github.com/cloudfoundry/ruby-buildpack#master",
				"go_buildpack",
			}
		})

		It("finds the supplied dependency in the runtime container", func() {
			PushAppAndConfirm(app)

			Expect(app.Stdout.String()).To(MatchRegexp("Installing ruby \\d+\\.\\d+\\.\\d+"))
			Expect(app.Stdout.String()).To(ContainSubstring("Go Buildpack version"))

			Expect(app.GetBody("/")).To(MatchRegexp("Bundler version \\d+\\.\\d+\\.\\d+"))
		})
	})
})
