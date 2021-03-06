# kuadrantctl
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)

Kuadrant configuration command line utility

## Installing
Use `go install` to install the latest version of the library. This command will install the `kuadrantctl` binary executable in `$GOBIN` (defaulting to `$GOPATH/bin`).

```
go install github.com/kuadrant/kuadrantctl@latest
```

## Commands
* [Install Kuadrant](doc/install.md)
* [Uninstall Kuadrant](doc/uninstall.md)
* [Generate Istio virtualservice objects](doc/generate-istio-virtualservice.md)
* [Generate Istio authenticationpolicy objects](doc/generate-istio-authorizationpolicy.md)
* [Generate kuadrat authconfig objects](doc/generate-kuadrant-authconfig.md)
* [Generate Gateway API HTTPRoute objects](doc/generate-gateway-api-httproute.md)


## Contributing
The [Development guide](doc/development.md) describes how to build the kuadrantctl CLI and how to test your changes before submitting a patch or opening a PR.

## Licensing

This software is licensed under the [Apache 2.0 license](https://www.apache.org/licenses/LICENSE-2.0).

See the LICENSE and NOTICE files that should have been provided along with this software for details.
