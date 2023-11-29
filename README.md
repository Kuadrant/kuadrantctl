# kuadrantctl
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)

Kuadrant configuration command line utility

## Installing
Use `go install` to install the latest version of the library. This command will install the `kuadrantctl` binary executable in `$GOBIN` (defaulting to `$GOPATH/bin`).

```
go install github.com/kuadrant/kuadrantctl@latest
```
> Golang 1.21+ required

## Using with GitHub Actions

```yaml
      - name: Install kuadrantctl
        uses: jaxxstorm/action-install-gh-release@v1.10.0
        with: # Grab the latest version
          repo: Kuadrant/kuadrantctl
```

## Commands
* [Install Kuadrant](doc/install.md)
* [Uninstall Kuadrant](doc/uninstall.md)
* [Generate Gateway API HTTPRoute objects from OpenAPI 3.X](doc/generate-gateway-api-httproute.md)
* [Generate Kuadrant RateLimitPolicy from OpenAPI 3.X](doc/generate-kuadrant-rate-limit-policy.md)
* [Generate Kuadrant AuthPolicy from OpenAPI 3.X](doc/generate-kuadrant-auth-policy.md)

## Contributing
The [Development guide](doc/development.md) describes how to build the kuadrantctl CLI and how to test your changes before submitting a patch or opening a PR.

## Licensing

This software is licensed under the [Apache 2.0 license](https://www.apache.org/licenses/LICENSE-2.0).

See the LICENSE and NOTICE files that should have been provided along with this software for details.
