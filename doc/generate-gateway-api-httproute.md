## Generate Gateway API HTTPRoute object from OpenAPI 3

The `kuadrantctl generate gatewayapi httproute` command generates an [Gateway API HTTPRoute](https://gateway-api.sigs.k8s.io/v1alpha2/guides/http-routing/)
from your [OpenAPI Specification (OAS) 3.x](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.2.md) powered with [kuadrant extensions](openapi-kuadrant-extensions.md).

### OpenAPI specification

[OpenAPI `v3.0`](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md)

OpenAPI document resource can be provided by one of the following channels:
* Filename in the available path.
* URL format (supported schemes are HTTP and HTTPS). The CLI will try to download from the given address.
* Read from stdin standard input stream.

### Usage

```shell
$ kuadrantctl generate gatewayapi httproute -h
Generate Gateway API HTTPRoute from OpenAPI 3.0.X

Usage:
  kuadrantctl generate gatewayapi httproute [flags]

Flags:
  -h, --help         help for httproute
      --oas string   /path/to/file.[json|yaml|yml] OR http[s]://domain/resource/path.[json|yaml|yml] OR @ (required)

Global Flags:
  -v, --verbose   verbose output
```

> Under the example folder there are examples of OAS 3 that can be used to generate the resources
