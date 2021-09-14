# Cloud Foundry Go(Lang) Buildpack

[![CF Slack](https://www.google.com/s2/favicons?domain=www.slack.com) Join us on Slack](https://cloudfoundry.slack.com/messages/buildpacks/)

A Cloud Foundry [buildpack](http://docs.cloudfoundry.org/buildpacks/) for Go(lang) based apps.

### Buildpack User Documentation

Official buildpack documentation can be found at [go buildpack docs](http://docs.cloudfoundry.org/buildpacks/go/index.html).

### Building the Buildpack

To build this buildpack, run the following command from the buildpack's directory:

```bash
./scripts/package.sh --stack cflinuxfs3 --version <version>
```

You can then find the built artifact in `./build/buildpack.zip`.

### Use in Cloud Foundry

Upload the buildpack to your Cloud Foundry and optionally specify it by name

```bash
cf create-buildpack [BUILDPACK_NAME] [BUILDPACK_ZIP_FILE_PATH] 1
cf push my_app [-b BUILDPACK_NAME]
```

### Testing

Buildpacks use the [Switchblade](https://github.com/cloudfoundry/switchblade) framework for running integration tests.

To test this buildpack, run the following command from the buildpack's directory:

1. Run unit tests

    ```bash
    ./scripts/unit.sh
    ```

1. Run integration tests

    ```bash
    ./scripts/integration.sh --github-token <token> --platform <cf|docker>
    ```

More information can be found on the [switchblade repo](https://github.com/cloudfoundry/switchblade).

### Contributing

Find our guidelines [here](./CONTRIBUTING.md).

### Help and Support

Join the #buildpacks channel in our [Slack community](http://slack.cloudfoundry.org/) if you need any further assistance.

### Reporting Issues

Please fill out the issue template fully if you'd like to start an issue for the buildpack.

### Active Development

The project backlog is on [Pivotal Tracker](https://www.pivotaltracker.com/projects/1042066)

### Acknowledgements

Inspired by the [Heroku buildpack](https://github.com/heroku/heroku-buildpack-go).
