# kuadrantctl
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)

Kuadrant configuration command line utility

## Installing
Use `go get` to install the latest version of the library. This command will install the `kuadrantctl` binary executable in `$GOBIN` (defaulting to `$GOPATH/bin`).

```
go install github.com/kuadrant/kuadrantctl@latest
```

## Commands

* Kuadrant API manifest subcommands `kuadrantctl api <subcommand>`
    * [Generate Kuadrant API manifest](doc/api-generate.md)
    * [Create Kuadrant API manifest](doc/api-create.md)
    * [Install Kuadrant](doc/install.md)


## Contributing
The [Development guide](doc/development.md) describes how to build the kuadrantctl CLI and how to test your changes before submitting a patch or opening a PR.

## Licensing

This software is licensed under the [Apache 2.0 license](https://www.apache.org/licenses/LICENSE-2.0).

See the LICENSE and NOTICE files that should have been provided along with this software for details.
