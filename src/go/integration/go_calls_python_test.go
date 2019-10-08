package integration_test

import (
	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("running supply python buildpack before the go buildpack", func() {
	var app *cutlass.App

	BeforeEach(func() {
		if ok, err := cutlass.ApiGreaterThan("2.65.1"); err != nil || !ok {
			Skip("API version does not have multi-buildpack support")
		}
	})

	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	pushApp := func(fixture string) {
		app = cutlass.New(Fixtures(fixture))
		app.Buildpacks = []string{
			"https://github.com/cloudfoundry/python-buildpack#master",
			"go_buildpack",
		}
		app.Disk = "1G"
		PushAppAndConfirm(app)
	}

	It("an app is pushed which uses pip dependencies", func() {
		pushApp("go_calls_python")

		Expect(app.Stdout.String()).To(ContainSubstring("Installing python"))

		Expect(app.GetBody("/")).To(ContainSubstring(`[{"hello":"world"}]`))
	})

	It("an app is pushed which uses miniconda", func() {
		pushApp("go_calls_python_miniconda")

		Expect(app.Stdout.String()).To(ContainSubstring("Installing Miniconda"))

		Expect(app.GetBody("/")).To(ContainSubstring(`[{"hello":"world"}]`))
	})

	It("an app is pushed which uses NLTK corpus", func() {
		pushApp("go_calls_python_nltk")

		Expect(app.Stdout.String()).To(ContainSubstring("Downloading NLTK corpora..."))

		Expect(app.GetBody("/")).To(ContainSubstring("The Fulton County Grand Jury said Friday an investigation of Atlanta's recent primary election produced"))
	})
})
