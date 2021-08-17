module github.com/cloudfoundry/go-buildpack

go 1.16

require (
	github.com/Dynatrace/libbuildpack-dynatrace v1.3.0
	github.com/Masterminds/semver v1.5.0
	github.com/ZiCog/shiny-thing v0.0.0-20121130081921-e9e19444ccf5
	github.com/cloudfoundry/libbuildpack v0.0.0-20210726164432-80929621d448
	github.com/golang/mock v1.6.0
	github.com/kr/go-heroku-example v0.0.0-20150601175414-712a6d2f98f1
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.14.0
	github.com/sclevine/spec v1.4.0
	github.com/vendorlib v0.0.0-00010101000000-000000000000
)

replace github.com/vendorlib => ./fixtures/default/install_package_spec/vendored/vendor/github.com/vendorlib
