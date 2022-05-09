module github.com/cloudfoundry/go-buildpack

go 1.16

require (
	github.com/Dynatrace/libbuildpack-dynatrace v1.5.0
	github.com/Masterminds/semver v1.5.0
	github.com/Microsoft/go-winio v0.5.2 // indirect
	github.com/ZiCog/shiny-thing v0.0.0-20121130081921-e9e19444ccf5
	github.com/cloudfoundry/libbuildpack v0.0.0-20220509111721-05ef1d6ca1f1
	github.com/cloudfoundry/switchblade v0.0.3
	github.com/containerd/containerd v1.6.4 // indirect
	github.com/docker/distribution v2.8.1+incompatible // indirect
	github.com/docker/docker v20.10.15+incompatible // indirect
	github.com/elazarl/goproxy v0.0.0-20220417044921-416226498f94 // indirect
	github.com/fsnotify/fsnotify v1.5.4 // indirect
	github.com/golang/mock v1.6.0
	github.com/kr/go-heroku-example v0.0.0-20150601175414-712a6d2f98f1
	github.com/onsi/ginkgo v1.16.5
	github.com/onsi/gomega v1.19.0
	github.com/sclevine/spec v1.4.0
	github.com/tidwall/gjson v1.14.1 // indirect
	github.com/vendorlib v0.0.0-00010101000000-000000000000
	golang.org/x/net v0.0.0-20220425223048-2871e0cb64e4 // indirect
	golang.org/x/sys v0.0.0-20220503163025-988cb79eb6c6 // indirect
	google.golang.org/genproto v0.0.0-20220505152158-f39f71e6c8f3 // indirect
)

replace github.com/vendorlib => ./fixtures/default/install_package_spec/vendored/vendor/github.com/vendorlib
