## Generate Kuadrant AuthConfig objects

The `kuadrantctl generate kuadrant authconfig` command generates an [Authorino AuthConfig](https://github.com/Kuadrant/authorino/blob/v0.7.0/docs/architecture.md#the-authorino-authconfig-custom-resource-definition-crd)
from your [OpenAPI Specification (OAS) 3.x](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.2.md) and kubernetes service information.

### OpenAPI specification

OpenAPI document resource can be provided by one of the following channels:
* Filename in the available path.
* URL format (supported schemes are HTTP and HTTPS). The CLI will try to download from the given address.
* Read from stdin standard input stream.

### Usage :

```shell
$ kuadrantctl generate kuadrant authconfig -h
Generate kuadrant authconfig from OpenAPI 3.x

Usage:
  kuadrantctl generate kuadrant authconfig [flags]

Flags:
  -h, --help                 help for authconfig
      --oas string           /path/to/file.[json|yaml|yml] OR http[s]://domain/resource/path.[json|yaml|yml] OR - (required)
      --public-host string   The address used by a client when attempting to connect to a service (required)

Global Flags:
  -v, --verbose   verbose output
```
