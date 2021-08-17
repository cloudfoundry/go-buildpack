package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/libbuildpack/cutlass"
	"github.com/sclevine/spec"

	. "github.com/onsi/gomega"
)

func testOverride(t *testing.T, context spec.G, it spec.S) {
	var (
		Expect = NewWithT(t).Expect

		app *cutlass.App
	)

	it.Before(func() {
		app = cutlass.New(filepath.Join(settings.FixturesPath, "default", "vendored"))
		app.Buildpacks = []string{"override_buildpack", "go_buildpack"}
	})

	it.After(func() {
		app = DestroyApp(t, app)
	})

	it("Forces go from override buildpack", func() {
		Expect(app.V3Push()).ToNot(Succeed())
		Expect(app.Stdout.String()).To(ContainSubstring("-----> OverrideYML Buildpack"))
		Expect(app.ConfirmBuildpack(settings.Buildpack.Version)).To(Succeed())

		Expect(app.Stdout.String()).To(ContainSubstring("-----> Installing go"))
		Expect(app.Stdout.String()).To(MatchRegexp("Copy .*/go.tgz"))
		Expect(app.Stdout.String()).To(ContainSubstring("Error installing Go: dependency sha256 mismatch: expected sha256 062d906c87839d03b243e2821e10653c89b4c92878bfe2bf995dec231e117bfc, actual sha256 b56b58ac21f9f42d032e1e4b8bf8b8823e69af5411caa15aee2b140bc756962f"))
	})
}
