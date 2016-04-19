# Cloud Foundry Go(Lang) Buildpack
[![CF Slack](https://s3.amazonaws.com/buildpacks-assets/buildpacks-slack.svg)](http://slack.cloudfoundry.org)

A Cloud Foundry [buildpack](http://docs.cloudfoundry.org/buildpacks/) for Go(lang) based apps.

This is based on the [Heroku buildpack] (https://github.com/heroku/heroku-buildpack-go).

Additional documentation can be found at [CloudFoundry.org](http://docs.cloudfoundry.org/buildpacks/).

## Usage

The Go buildpack will be automatically detected if your app has been packaged with [godep](https://github.com/tools/godep) using `godep save`. It will also be automatically detected if your app has a `vendor/` directory and your app has any files ending with `.go`.

If your Cloud Foundry deployment does not have the Go Buildpack installed, or the installed version is out of date, you can use the latest version with the command:

```bash
cf push my_app -b https://github.com/cloudfoundry/go-buildpack.git
```

Your app must specify a start command. The start command can be placed in the file `Procfile` in your app's root directory. For example, if your Go app's package name is `my-go-server`, the following `Procfile` would provide the correct start command:

```
web: my-go-server
```

You can also specify your app's start command in the `manifest.yml` file in the root directory, for example:

```yaml
---
applications:
  - name: my-app-name
    command: my-go-server
```

### Fixing a Go Version

If a version of Go is unspecified, the Go buildpack will default to the most recent version of Go it has available.

If you are using godep, you can fix your Go version in  `GoVersion` key of the `Godeps/Godeps.json` file.

If you are using the `vendor/` directory for dependencies, you can set the Go version with the `GOVERSION` environment variable. For example, a `manifest.yml` to request the most recent Go 1.5:

```yaml
---
applications:
  - name: my-app-name
    env:
      GOVERSION: go1.5
      GOPACKAGENAME: app-package-name
      GO15VENDOREXPERIMENT: 1
```

If you are fixing a Go version, make sure it is supported by the Go buildpack. You can see a list of supported versions in the [release notes](https://github.com/cloudfoundry/go-buildpack/releases).

## Dependency management

Please note that this buildpack only supports the [godep](https://github.com/tools/godep) package manager if you are not packaging your dependencies in the `vendor/` directory.

However, as long as your package manager supports the new Go standard and places dependencies correctly into `vendor/`, your app should stage if you follow the instructions below.

### Vendoring with godep

If you are using [godep](https://github.com/tools/godep) to package your dependencies, make sure that you have created a valid `Godeps/Godeps.json` file in the root directory of your app by running `godep save`. See this [test app](https://github.com/cloudfoundry/go-buildpack/tree/master/cf_spec/fixtures/go_app_with_dependencies/src/go_app_with_dependencies) for an example app that uses godep.

**NOTE**: if you are using godep with Go 1.6, you must set the `GO15VENDOREXPERIMENT` environment variable to 0, otherwise your app will not stage. 

### Go native vendoring

If you are using the native Go vendoring system, which packages all local dependencies in the `vendor/` directory, you must specify your app's package name in the `GOPACKAGENAME` environment variable. An example `manifest.yml`:

```yaml
---
applications:
 - name: my-app-name
   command: example-project
   env:
     GOPACKAGENAME: github.com/example-org/example-project
```

**NOTE**: For Go 1.5, since native vendoring is turned off by default, you must set the environment variable `GO15VENDOREXPERIMENT` to 1 in your `manifest.yml` to use this feature.

For further help, please see this [fixture app](https://github.com/cloudfoundry/go-buildpack/tree/develop/cf_spec/fixtures/go_with_native_vendoring/src/go_app) for a sample app using `vendor/`.

### C dependencies

This buildpack supports building with C dependencies via
[cgo](https://golang.org/cmd/cgo/). You can set config vars to specify CGO flags
to, e.g., specify paths for vendored dependencies. E.g., to build
[gopgsqldriver](https://github.com/jbarham/gopgsqldriver), add the config var
`CGO_CFLAGS` with the value `-I/app/code/vendor/include/postgresql` and include
the relevant Postgres header files in `vendor/include/postgresql/` in your app.

## Disconnected environments
To use this buildpack on Cloud Foundry, where the Cloud Foundry instance limits some or all internet activity, please read the [Disconnected Environments documentation](https://github.com/cf-buildpacks/buildpack-packager/blob/master/doc/disconnected_environments.md).

## Proxy Support

If you need to use a proxy to download dependencies during staging, you can set
the `http_proxy` and/or `https_proxy` environment variables. For more information, see
the [Proxy Usage Docs](http://docs.cloudfoundry.org/buildpacks/proxy-usage.html).

### Vendoring app dependencies
As stated in the [Disconnected Environments documentation](https://github.com/cf-buildpacks/buildpack-packager/blob/master/doc/disconnected_environments.md), your application must vendor its dependencies locally.

As described in the "Dependency Management" section, you can use [Godep](https://github.com/tools/godep) or the Go vendoring system using the `vendor/` directory.

```cf push``` uploads your vendored dependencies. The buildpack will compile any dependencies requiring compilation while staging your application.

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

## Supported binary dependencies

The buildpack only supports the stable patches for each dependency listed in the [manifest.yml](manifest.yml) and [releases page](https://github.com/cloudfoundry/go-buildpack/releases).


If you try to use a binary that is not currently supported, staging your app will fail and you will see the following error message:

```
       Could not get translated url, exited with: DEPENDENCY_MISSING_IN_MANIFEST: ...
 !
 !     exit
 !
Staging failed: Buildpack compilation step failed
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
