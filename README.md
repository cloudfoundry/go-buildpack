# CloudFoundry build pack: Go(Lang)

A Cloud Foundry [buildpack](http://docs.cloudfoundry.org/buildpacks/) for Go(lang) based apps.

This is based on the [Heroku buildpack] (https://github.com/kr/heroku-buildpack-go).

Additional documentation can be found at [CloudFoundry.org](http://docs.cloudfoundry.org/buildpacks/).

## Usage

This buildpack will get used if you have any files with the `.go` extension in your repository.

```bash
cf push my_app -b https://github.com/cloudfoundry/go-buildpack.git
```

## Cloud Foundry Extensions - Cached Dependencies

The primary purpose of extending the heroku buildpack is to cache system dependencies for partially or fully disconnected environments.
Historically, this was called 'offline' mode.
It is now called 'Cached dependencies'.

Cached buildpacks can be used in any environment where you would prefer the dependencies to be cached instead of fetched from the internet.

The list of what is cached is maintained in [the manifest](manifest.yml). For a description of the manifest file, see the [buildpack packager documentation](https://github.com/cf-buildpacks/buildpack-packager/blob/master/README.md#manifest)

The buildpack consumes cached system dependencies during staging by translating remote urls. Search for 'translate_dependency_url' in this repo to see examples.

### App Dependencies in Cached Mode
Cached (offline) mode expects each app to use [Godep](https://github.com/tools/godep) to manage dependencies. The Godep folder should be populated before pushing your app.

_Deprecated_

A .godir file containing the name of your application can be used to build the project.

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
  BUNDLE_GEMFILE=cf.Gemfile bundle exec buildpack-packager [ online | offline ]
  ```
    
1. Use in Cloud Foundry

    Upload the buildpack to your Cloud Foundry and optionally specify it by name
        
    ```bash
    cf create-buildpack custom_go_buildpack go_buildpack-offline-custom.zip 1
    cf push my_app -b custom_go_buildpack
    ```  

## Contributing

### Run the tests

See the [Machete](https://github.com/cf-buildpacks/machete) CF buildpack test framework for more information.


### Pull Requests

1. Fork the project
1. Submit a pull request

## .godir and Godeps

Early versions of this buildpack required users to
create a `.godir` file in the root of the project,
containing the application name in order to build the
project. While using a `.godir` file is still supported,
it has been deprecated in favor of using
[godep](https://github.com/kr/godep) in your project to
manage dependencies, and including the generated `Godep`
directory in your git repository.

## Reporting Issues

Open an issue on this project

## Active Development

The project backlog is on [Pivotal Tracker](https://www.pivotaltracker.com/projects/1042066)
