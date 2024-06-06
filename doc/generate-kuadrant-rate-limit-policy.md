## Generate Kuadrant RateLimitPolicy object from OpenAPI 3

The `kuadrantctl generate kuadrant ratelimitpolicy` command generates a [Kuadrant RateLimitPolicy](https://docs.kuadrant.io/kuadrant-operator/doc/rate-limiting/)
from your [OpenAPI Specification (OAS) 3.x document](https://spec.openapis.org/oas/latest.html) powered with [Kuadrant extensions](openapi-kuadrant-extensions.md).

### OpenAPI specification

An OpenAPI document resource can be provided to the Kuadrant CLI in one of the following ways:

* Filename in the available path.
* URL format (supported schemes are HTTP and HTTPS). The CLI will try to download from the given address.
* Read from `stdin` standard input stream.

### Usage

```shell
Generate Kuadrant RateLimitPolicy from OpenAPI 3.0.x

Usage:
  kuadrantctl generate kuadrant ratelimitpolicy [flags]

Flags:
  -h, --help         help for ratelimitpolicy
  --oas string        Path to OpenAPI spec file (in JSON or YAML format), URL, or '-' to read from standard input (required)
  -o Output format:   'yaml' or 'json'. (default "yaml")

Global Flags:
  -v, --verbose   verbose output
```

> **Note**: The `kuadrantctl/examples` directory in GitHub includes sample OAS 3 files that you can use to generate the resources.

### Procedure

1. Clone the Git repository as follows: 
```bash
git clone https://github.com/Kuadrant/kuadrantctl.git
cd kuadrantctl
 ```
2. Set up a cluster, Istio and Gateway API CRDs, and Kuadrant as follows: 

* Use the single-cluster quick start script to install Kuadrant in a local `kind` cluster: https://docs.kuadrant.io/getting-started-single-cluster/.

3. Build and install the CLI in `bin/kuadrantctl` path as follows:
```bash
make install
```

4. Deploy the Petstore backend API as follows:
```bash
kubectl create namespace petstore
kubectl apply -n petstore -f examples/petstore/petstore.yaml
```

5. Create the Petstore OpenAPI definition as follows:
<details>

```yaml
cat <<EOF >petstore-openapi.yaml
---
openapi: "3.0.3"
info:
  title: "Pet Store API"
  version: "1.0.0"
x-kuadrant:
  route:
    name: "petstore"
    namespace: "petstore"
    hostnames:
      - example.com
    parentRefs:
      - name: istio-ingressgateway
        namespace: istio-system
servers:
  - url: https://example.io/v1
paths:
  /cat:
    x-kuadrant:  ## Path level Kuadrant Extension
      backendRefs:
        - name: petstore
          port: 80
          namespace: petstore
      rate_limit:
        rates:
          - limit: 1
            duration: 10
            unit: second
        counters:
          - request.headers.x-forwarded-for
    get:  # Added to the route and rate limited
      operationId: "getCat"
      responses:
        405:
          description: "invalid input"
    post:  # NOT added to the route
      x-kuadrant: 
        disable: true
      operationId: "postCat"
      responses:
        405:
          description: "invalid input"
  /dog:
    get:  # Added to the route and rate limited
      x-kuadrant:  ## Operation level Kuadrant Extension
        backendRefs:
          - name: petstore
            port: 80
            namespace: petstore
        rate_limit:
          rates:
            - limit: 3
              duration: 10
              unit: second
          counters:
            - request.headers.x-forwarded-for
      operationId: "getDog"
      responses:
        405:
          description: "invalid input"
    post:  # Added to the route and NOT rate limited
      x-kuadrant:  ## Operation level Kuadrant Extension
        backendRefs:
          - name: petstore
            port: 80
            namespace: petstore
      operationId: "postDog"
      responses:
        405:
          description: "invalid input"
EOF
```
</details>

> **Note**: The `servers` base path is not included. WIP in following up PRs.

| Operation | Applied configuration |
| --- | --- |
| `GET /cat` | Should return 200 OK and be rate limited (1 req / 10 seconds). |
| `POST /cat`  | Not added to the HTTPRoute. Should return 404 Not Found. |
| `GET /dog`  | Should return 200 OK and be rate limited (3 req / 10 seconds). |
| `POST /dog`   | Should return 200 OK and NOT rate limited. |


6. Create the HTTPRoute by using the CLI as follows:
```bash
bin/kuadrantctl generate gatewayapi httproute --oas petstore-openapi.yaml | kubectl apply -n petstore -f -
```

7. Create the rate limit policy as follows:
```bash
bin/kuadrantctl generate kuadrant ratelimitpolicy --oas petstore-openapi.yaml | kubectl apply -n petstore -f -
```

8. Test the OpenAPI endpoints as follows:

  * `GET /cat` - Should return 200 OK and be rate limited (1 req / 10 seconds).
```bash
curl --resolve example.com:9080:127.0.0.1 -v "http://example.com:9080/cat"
```
  *   `POST /cat` - Not added to the HTTPRoute. Should return 404 Not Found.
```bash
curl --resolve example.com:9080:127.0.0.1 -v -X POST "http://example.com:9080/cat"
```
  * `GET /dog` - Should return 200 OK and be rate limited (3 req / 10 seconds).

```bash
curl --resolve example.com:9080:127.0.0.1 -v "http://example.com:9080/dog"
```
  *  `POST /dog` - Should return 200 OK and NOT rate limited.

```bash
curl --resolve example.com:9080:127.0.0.1 -v -X POST "http://example.com:9080/dog"
```
