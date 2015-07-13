# CloudFoundry build pack: Go(Lang)

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

The buildpack only supports the two most stable patches for each dependency in the [manifest.yml](manifest.yml).

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
  BUNDLE_GEMFILE=cf.Gemfile bundle exec buildpack-packager [ uncached | cached ]
  ```

1. Use in Cloud Foundry

    Upload the buildpack to your Cloud Foundry and optionally specify it by name
        
    ```bash
    cf create-buildpack custom_go_buildpack go_buildpack-cached-custom.zip 1
    cf push my_app -b custom_go_buildpack
    ```  

### Deprecated Versions

If you would like to build the buildpack with previously supported dependency versions, provide the `--use-custom-manifest=manifest-including-unsupported.yml` option to `buildpack-packager`.

## Contributing

Find our guidelines [here](./CONTRIBUTING.md).

## .godir and Godeps

Early versions of this buildpack required users to
create a `.godir` file in the root of the project,
containing the application name in order to build the
project. While using a `.godir` file is still supported,
it has been deprecated in favor of using
[godep](https://github.com/kr/godep) in your project to
manage dependencies, and including the generated `Godep`
directory in your git repository.

This buildpack supports building with C dependencies via
[cgo][cgo]. You can set config vars to specify CGO flags
to, e.g., specify paths for vendored dependencies. E.g., to build
[gopgsqldriver](https://github.com/jbarham/gopgsqldriver), add the config var
`CGO_CFLAGS` with the value `-I/app/code/vendor/include/postgresql` and include
the relevant Postgres header files in `vendor/include/postgresql/` in your app.

## Reporting Issues

Open an issue on this project

## Active Development

The project backlog is on [Pivotal Tracker](https://www.pivotaltracker.com/projects/1042066)
