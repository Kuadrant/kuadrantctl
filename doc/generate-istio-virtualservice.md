## Generate Istio VirtualService objects

The `kuadrantctl generate istio virtualservice` command generates an [Istio VirtualService](https://istio.io/latest/docs/reference/config/networking/virtual-service/)
from your [OpenAPI Specification (OAS) 3.x](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.2.md) and kubernetes service information.

### OpenAPI specification

OpenAPI document resource can be provided by one of the following channels:
* Filename in the available path.
* URL format (supported schemes are HTTP and HTTPS). The CLI will try to download from the given address.
* Read from stdin standard input stream.

### Usage :

```shell
$ kuadrantctl generate istio virtualservice -h
Generate Istio VirtualService from OpenAPI 3.x

Usage:
  kuadrantctl generate istio virtualservice [flags]

Flags:
      --gateway strings            Gateways (required)
  -h, --help                       help for virtualservice
      --oas string                 /path/to/file.[json|yaml|yml] OR http[s]://domain/resource/path.[json|yaml|yml] OR - (required)
      --path-prefix-match          Path match type (defaults to exact match type)
      --public-host string         The address used by a client when attempting to connect to a service (required)
      --service-name string        Service name (required)
      --service-namespace string   Service namespace (required)
  -p, --service-port int32         Service port (default 80)

Global Flags:
  -v, --verbose   verbose output
```
