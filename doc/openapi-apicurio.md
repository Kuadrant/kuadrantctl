## Using Apicurio with Kuadrant OAS extensions

[OpenAPI Extensions](https://swagger.io/docs/specification/openapi-extensions/) can
be used to describe extra functionality beyond what is covered by the standard OpenAPI
specification. They typically start with `x-`. Kuadrant OpenAPI extensions start with
`x-kuadrant`, and allow you to configure Kuadrant policy information along side
your API.

Apicurio Studio is a UI tool for visualising and editing OpenAPI specifications.
It has support for visualising security and extensions defined in your spec.

This guide assumes you have Apicurio Studio already running.
See https://www.apicur.io/studio/ for info on how to install Apicurio Studio.

### Adding extensions to the spec

Open or import your OpenAPI spec in the Apicurio Studio UI.
You can modify the source of the spec from the UI.
There are a few different configuration and extension points supported by Apicurio Studio, and also supported by the `kuadrantctl` cli.

To generate a [HTTPRoute](https://gateway-api.sigs.k8s.io/api-types/httproute/) for the API, add the following `x-kuadrant` block to your spec in the UI, replacing values to match your APIs details and the location of your Gateway.

```yaml
info:
    x-kuadrant:
        route:
            name: petstore
            namespace: petstore
            hostnames:
                - 'petstore.example.com'
            parentRefs:
                -   name: prod-web
                    namespace: kuadrant-multi-cluster-gateways
                    kind: Gateway
```

See [this guide](./generate-gateway-api-httproute.md) for more info on generating a HTTPRoute.

To generate an [AuthPolicy](https://docs.kuadrant.io/kuadrant-operator/doc/auth/), add a `securityScheme` to the components block.
This `securityScheme` requires that an API key header is set.
Although securityScheme is not an OpenAPI extension, it is used by `kuadrantctl` like the other extensions mentioned here.

```yaml
    securitySchemes:
        api_key:
            type: apiKey
            name: api_key
            in: header
```

When added, the UI will display this in the security requirements section:

![Apicurio security requirements](./images/apicurio-security-scheme-apikey.png)

See [this guide](./generate-kuadrant-auth-policy.md) for more info on generating an AuthPolicy.

To generate a [RateLimitPolicy](https://docs.kuadrant.io/kuadrant-operator/doc/rate-limiting/) for the API, add the following `x-kuadrant` block to a path in your spec,
replacing values to match your APIs details.

```yaml
paths:
    /:
        x-kuadrant:
            backendRefs:
                -
                    name: petstore
                    namespace: petstore
                    port: 8080
            rate_limit:
                rates:
                    -
                        limit: 2
                        duration: 10
                        unit: second
```

When added, the UI will show this in Vendor Extensions section for that specific path:

![Apicurio RateLimitPolicy Vendor Extension](./images/apicurio-vendor-extension-backend-rate-limit.png)

See [this guide](./generate-kuadrant-rate-limit-policy.md) for more info on generating a RateLimitPoliicy.
There is also the full [kuadrantctl guide](./openapi-kuadrant-extensions.md).
