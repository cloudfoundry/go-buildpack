# Cloud Foundry Go(Lang) Buildpack
[![CF Slack](https://s3.amazonaws.com/buildpacks-assets/buildpacks-slack.svg)](http://slack.cloudfoundry.org)

A Cloud Foundry [buildpack](http://docs.cloudfoundry.org/buildpacks/) for Go(lang) based apps.

This is based on the [Heroku buildpack] (https://github.com/heroku/heroku-buildpack-go).

Additional documentation can be found at [CloudFoundry.org](http://docs.cloudfoundry.org/buildpacks/).

## Usage
=======

This buildpack will get used if you have any files with the `.go` extension in your repository.

```bash
cf push my_app -b https://github.com/cloudfoundry/go-buildpack.git
```
## Disconnected environments
To use this buildpack on Cloud Foundry, where the Cloud Foundry instance limits some or all internet activity, please read the [Disconnected Environments documentation](https://github.com/cf-buildpacks/buildpack-packager/blob/master/doc/disconnected_environments.md).

### Vendoring app dependencies
As stated in the [Disconnected Environments documentation](https://github.com/cf-buildpacks/buildpack-packager/blob/master/doc/disconnected_environments.md), your application must 'vendor' it's dependencies.

For the Go buildpack, use [Godep](https://github.com/tools/godep):

```cf push``` uploads your vendored dependencies. The buildpack will compile any dependencies requiring compilation while staging your application.

## Building

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


## Dependency managment

### .godir

If you use `.godir` your app will no longer stage.

`.godir` has been retired in favor of using
[godep](https://github.com/tools/godep) in your project to
manage dependencies, and including the generated `Godep`
directory in your git repository.

### Godeps

[Godeps](https://github.com/tools/godep) is the buildpack's only supported
package manager. The buildpack will run `godep` to install your dependencies at
staging.

### C dependencies

This buildpack supports building with C dependencies via
[cgo](https://golang.org/cmd/cgo/). You can set config vars to specify CGO flags
to, e.g., specify paths for vendored dependencies. E.g., to build
[gopgsqldriver](https://github.com/jbarham/gopgsqldriver), add the config var
`CGO_CFLAGS` with the value `-I/app/code/vendor/include/postgresql` and include
the relevant Postgres header files in `vendor/include/postgresql/` in your app.

## Help and Support

Join the #buildpacks channel in our [Slack community] (http://slack.cloudfoundry.org/) 

## Reporting Issues

Open an issue on this project

## Active Development

The project backlog is on [Pivotal Tracker](https://www.pivotaltracker.com/projects/1042066)
