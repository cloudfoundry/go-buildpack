module github.com/cloudfoundry/go-buildpack

go 1.16

require (
	github.com/Dynatrace/libbuildpack-dynatrace v1.4.1
	github.com/Masterminds/semver v1.5.0
	github.com/Microsoft/go-winio v0.5.0 // indirect
	github.com/ZiCog/shiny-thing v0.0.0-20121130081921-e9e19444ccf5
	github.com/cloudfoundry/libbuildpack v0.0.0-20211012151659-ed6fb71f47bf
	github.com/cloudfoundry/switchblade v0.0.3
	github.com/containerd/containerd v1.5.5 // indirect
	github.com/fsnotify/fsnotify v1.5.1 // indirect
	github.com/golang/mock v1.6.0
	github.com/kr/go-heroku-example v0.0.0-20150601175414-712a6d2f98f1
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.16.0
	github.com/paketo-buildpacks/packit v1.0.1 // indirect
	github.com/sclevine/spec v1.4.0
	github.com/vendorlib v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20210916014120-12bc252f5db8 // indirect
	golang.org/x/sys v0.0.0-20211007075335-d3039528d8ac // indirect
	google.golang.org/genproto v0.0.0-20210916144049-3192f974c780 // indirect
)

replace github.com/vendorlib => ./fixtures/default/install_package_spec/vendored/vendor/github.com/vendorlib
