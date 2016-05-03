# Cloud Foundry Go(Lang) Buildpack
[![CF Slack](https://s3.amazonaws.com/buildpacks-assets/buildpacks-slack.svg)](http://slack.cloudfoundry.org)

A Cloud Foundry [buildpack](http://docs.cloudfoundry.org/buildpacks/) for Go(lang) based apps.

This is based on the [Heroku buildpack] (https://github.com/heroku/heroku-buildpack-go).

## Using the Buildpack

For information on deploying Go applications visit [CloudFoundry.org](http://docs.cloudfoundry.org/buildpacks/go/index.html).

## Building the Buildpack

1. Make sure you have fetched submodules

  ```bash
  git submodule update --init
  ```

1. Get latest buildpack dependencies

  ```shell
  BUNDLE_GEMFILE=cf.Gemfile bundle
  ```

1. Build the buildpack

  ```shell
  BUNDLE_GEMFILE=cf.Gemfile bundle exec buildpack-packager [ --uncached | --cached ]
  ```

1. Use in Cloud Foundry

  Upload the buildpack to your Cloud Foundry and optionally specify it by name

  ```bash
  cf create-buildpack custom_go_buildpack go_buildpack-cached-custom.zip 1
  cf push my_app -b custom_go_buildpack
  ```

## Testing
Buildpacks use the [Machete](https://github.com/cloudfoundry/machete) framework for running integration tests.

To test a buildpack, run the following command from the buildpack's directory:

```
BUNDLE_GEMFILE=cf.Gemfile bundle exec buildpack-build
```

More options can be found on Machete's [Github page.](https://github.com/cloudfoundry/machete)

## Contributing

Find our guidelines [here](./CONTRIBUTING.md).

## Help and Support

Join the #buildpacks channel in our [Slack community] (http://slack.cloudfoundry.org/) if you need any further assistance.

## Reporting Issues

Please fill out the issue template fully if you'd like to start an issue for the buildpack.

## Active Development

The project backlog is on [Pivotal Tracker](https://www.pivotaltracker.com/projects/1042066)

## Hacking on this Buildpack

To change this buildpack, fork it on GitHub. Push changes to your fork, then
create a test app with `--buildpack YOUR_GITHUB_GIT_URL` and push to it. If you
already have an existing app you may use `heroku config:add
BUILDPACK_URL=YOUR_GITHUB_GIT_URL` instead of `--buildpack`.


### Tests

Requires docker.

```console
$ make test
```

## Using with cgo

This buildpack supports building with C dependencies via [cgo][cgo]. You can set
config vars to specify CGO flags to specify paths for vendored dependencies. For
example, if you added C headers to and `includes` directory, add the config var
`CGO_CFLAGS` with the value `-I/app/code/includes` and include the relevant
header files in `includes/` in your app.

## Using a development version of Go

The buildpack can install and use any specific commit of the Go compiler when
the specified go version is `devel-<short sha>`. The version can be set either
via the appropriate vendoring tools config file or via the `$GOVERSION`
environment variable. The specific sha is downloaded from Github w/o git
history. Builds may fail if GitHub is down, but the compiled go version is
cached.

When this is used the buildpack also downloads and installs the buildpack's
current default Go version for use in bootstrapping the compiler.

Build tests are NOT RUN. Go compilation failures will fail a build.

No official support is provided for unreleased versions of Go.

## Passing a symbol (and optional string) to the linker

This buildpack supports the go [linker's][go-linker] ability (`-X symbol value`)
to set the value of a string at link time. This can be done by setting
`GO_LINKER_SYMBOL` and `GO_LINKER_VALUE` in the application's config before
pushing code. If `GO_LINKER_SYMBOL` is set, but `GO_LINKER_VALUE` isn't set then
`GO_LINKER_VALUE` defaults to [`$SOURCE_VERSION`][source-version].

This can be used to embed the commit sha, or other build specific data directly
into the compiled executable.

[go]: http://golang.org/
[buildpack]: http://devcenter.heroku.com/articles/buildpacks
[go-linker]: https://golang.org/cmd/ld/
[godep]: https://github.com/tools/godep
[govendor]: https://github.com/kardianos/govendor
[gb]: https://getgb.io/
[quickstart]: http://mmcgrana.github.com/2012/09/getting-started-with-go-on-heroku.html
[build-constraint]: http://golang.org/pkg/go/build/
[app-engine-build-constraints]: http://blog.golang.org/2013/01/the-app-engine-sdk-and-workspaces-gopath.html
[source-version]: https://devcenter.heroku.com/articles/buildpack-api#bin-compile
[cgo]: http://golang.org/cmd/cgo/
[vendor.json]: https://github.com/kardianos/vendor-spec
[gopgsqldriver]: https://github.com/jbarham/gopgsqldriver
[grp]: https://github.com/kardianos/govendor/commit/81ca4f23cab56f287e1d5be5ab920746fd6fb834
