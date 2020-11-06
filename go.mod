module github.com/cloudfoundry/go-buildpack

require (
	github.com/Dynatrace/libbuildpack-dynatrace v1.3.0
	github.com/Masterminds/semver v1.5.0
	github.com/ZiCog/shiny-thing v0.0.0-20121130081921-e9e19444ccf5
	github.com/cloudfoundry/libbuildpack v0.0.0-20201106010926-a2f840faf4eb
	github.com/golang/mock v1.4.4
	github.com/kr/go-heroku-example v0.0.0-20150601175414-712a6d2f98f1
	github.com/nxadm/tail v1.4.5 // indirect
	github.com/onsi/ginkgo v1.14.2
	github.com/onsi/gomega v1.10.3
	github.com/vendorlib v0.0.0-00010101000000-000000000000
	golang.org/x/sys v0.0.0-20201106081118-db71ae66460a // indirect
)

replace github.com/vendorlib => ./fixtures/vendored_custom_install_spec/vendor/github.com/vendorlib

go 1.13
