## Generate Istio AuthorizationPolicy objects

The `kuadrantctl generate istio authorizationpolicy` command generates an [Istio AuthorizationPolicy](https://istio.io/latest/docs/reference/config/security/authorization-policy/)
from your [OpenAPI Specification (OAS) 3.x](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.2.md) and kubernetes service information.

### OpenAPI specification

OpenAPI document resource can be provided by one of the following channels:
* Filename in the available path.
* URL format (supported schemes are HTTP and HTTPS). The CLI will try to download from the given address.
* Read from stdin standard input stream.

### Usage :

```shell
$ kuadrantctl generate istio authorizationpolicy -h
Generate Istio AuthorizationPolicy

Usage:
  kuadrantctl generate istio authorizationpolicy [flags]

Flags:
      --gateway-label strings   Gateway label (required)
  -h, --help                    help for authorizationpolicy
      --oas string              /path/to/file.[json|yaml|yml] OR http[s]://domain/resource/path.[json|yaml|yml] OR - (required)
      --public-host string      The address used by a client when attempting to connect to a service (required)

Global Flags:
  -v, --verbose   verbose output
```
