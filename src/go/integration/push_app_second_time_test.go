package integration_test

import (
	"path/filepath"

	"github.com/cloudfoundry/libbuildpack/cutlass"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("pushing an app a second time", func() {
	var app *cutlass.App
	AfterEach(func() {
		if app != nil {
			app.Destroy()
		}
		app = nil
	})

	BeforeEach(func() {
		if cutlass.Cached {
			Skip("but running cached tests")
		}

		app = cutlass.New(filepath.Join(bpDir, "fixtures", "single_file"))
		app.SetEnv("BP_DEBUG", "true")
	})

	Regexp := `\[.*\/go[\d\.]+\.linux-amd64-(cflinuxfs.*-)?[\da-f]+\.tar\.gz\]`
	DownloadRegexp := "Download " + Regexp
	CopyRegexp := "Copy " + Regexp

	It("uses the cache for manifest dependencies", func() {
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(MatchRegexp(DownloadRegexp))
		Expect(app.Stdout.String()).ToNot(MatchRegexp(CopyRegexp))

		app.Stdout.Reset()
		PushAppAndConfirm(app)
		Expect(app.Stdout.String()).To(MatchRegexp(CopyRegexp))
		Expect(app.Stdout.String()).ToNot(MatchRegexp(DownloadRegexp))
	})
})
