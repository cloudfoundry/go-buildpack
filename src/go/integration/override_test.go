package integration_test

import (
	"path/filepath"
	"testing"

	"github.com/cloudfoundry/switchblade"
	"github.com/sclevine/spec"

	. "github.com/cloudfoundry/switchblade/matchers"
	. "github.com/onsi/gomega"
)

func testOverride(platform switchblade.Platform, fixtures string) func(*testing.T, spec.G, spec.S) {
	return func(t *testing.T, context spec.G, it spec.S) {
		var (
			Expect = NewWithT(t).Expect

			name string
		)

		it.Before(func() {
			var err error
			name, err = switchblade.RandomName()
			Expect(err).NotTo(HaveOccurred())
		})

		it.After(func() {
			Expect(platform.Delete.Execute(name)).To(Succeed())
		})

		it("forces go from override buildpack", func() {
			_, logs, err := platform.Deploy.
				WithBuildpacks("override_buildpack", "go_buildpack").
				WithEnv(map[string]string{
					"GOPACKAGENAME": "example.com/user/go-online",
				}).
				Execute(name, filepath.Join(fixtures, "default", "vendored"))
			Expect(err).To(MatchError(ContainSubstring("App staging failed")))

			Expect(logs).To(ContainLines(ContainSubstring("-----> OverrideYML Buildpack")))
			Expect(logs).To(ContainLines(ContainSubstring("-----> Installing go")))
			Expect(logs).To(ContainLines(MatchRegexp("Copy .*/go.tgz")))
			Expect(logs).To(ContainLines(ContainSubstring("Error installing Go: dependency sha256 mismatch: expected sha256 062d906c87839d03b243e2821e10653c89b4c92878bfe2bf995dec231e117bfc, actual sha256 b56b58ac21f9f42d032e1e4b8bf8b8823e69af5411caa15aee2b140bc756962f")))
		})
	}
}
