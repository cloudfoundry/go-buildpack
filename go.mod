module github.com/cloudfoundry/go-buildpack

require (
	github.com/Dynatrace/libbuildpack-dynatrace v1.2.2
	github.com/Masterminds/semver v1.5.0
	github.com/ZiCog/shiny-thing v0.0.0-20121130081921-e9e19444ccf5
	github.com/cloudfoundry/libbuildpack v0.0.0-20200409131146-f0ac4f2efc24
	github.com/golang/mock v1.4.3
	github.com/kr/go-heroku-example v0.0.0-20150601175414-712a6d2f98f1
	github.com/onsi/ginkgo v1.12.0
	github.com/onsi/gomega v1.9.0
	github.com/vendorlib v0.0.0-00010101000000-000000000000
	golang.org/x/sys v0.0.0-20200420163511-1957bb5e6d1f // indirect
)

replace github.com/vendorlib => ./fixtures/vendored_custom_install_spec/vendor/github.com/vendorlib

go 1.13
