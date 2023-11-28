## Generate Kuadrant AuthPolicy object from OpenAPI 3

The `kuadrantctl generate kuadrant authpolicy` command generates an [Kuadrant AuthPolicy](https://github.com/Kuadrant/kuadrant-operator/blob/v0.4.1/doc/auth.md)
from your [OpenAPI Specification (OAS) 3.x](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.2.md) powered with [kuadrant extensions](openapi-kuadrant-extensions.md).

### OpenAPI specification

[OpenAPI `v3.0`](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md)

OpenAPI document resource can be provided by one of the following channels:
* Filename in the available path.
* URL format (supported schemes are HTTP and HTTPS). The CLI will try to download from the given address.
* Read from stdin standard input stream.

#### openIdConnect type
This initial version of the command only generates AuhPolicy when there is at least one security requirement referencing the
[Security Scheme Object](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.3.md#security-scheme-object) which type is `openIdConnect`.

### Description

The following OAS example has one protected endpoint `GET /dog` with OIDC sec scheme.

```yaml
paths:
  /dog:
    get:
      operationId: "getDog"
      security:
        - securedDog: []
      responses:
        405:
          description: "invalid input"
components:
  securitySchemes:
    securedDog:
      type: openIdConnect
      openIdConnectUrl: https://example.com/.well-known/openid-configuration
```

Running the command

```
kuadrantctl generate kuadrant authpolicy --oas ./petstore-openapi.yaml  | yq -P
```

The generated authpolicy (only relevan fields shown here):

```yaml
kind: AuthPolicy
apiVersion: kuadrant.io/v1beta2
metadata:
  name: petstore
  namespace: petstore
  creationTimestamp: null
spec:
  routeSelectors:
    - matches:
        - path:
            type: Exact
            value: /api/v1/dog
          method: GET
  rules:
    authentication:
      getDog:
        credentials: {}
        jwt:
          issuerUrl: https://example.com/.well-known/openid-configuration
        routeSelectors:
          - matches:
              - path:
                  type: Exact
                  value: /api/v1/dog
                method: GET
```

### Usage

```shell
Generate Kuadrant AuthPolicy from OpenAPI 3.0.X

Usage:
  kuadrantctl generate kuadrant authpolicy [flags]

Flags:
  -h, --help         help for authpolicy
      --oas string   /path/to/file.[json|yaml|yml] OR http[s]://domain/resource/path.[json|yaml|yml] OR @ (required)

Global Flags:
  -v, --verbose   verbose output
```

> Under the example folder there are examples of OAS 3 that can be used to generate the resources

### User Guide

* [Optional] Setup SSO service supporting OIDC. For this example, we will be using [keycloak](https://www.keycloak.org).
  * Create a new realm `petstore`
  * Create a client `petstore`. In the Client Protocol field, select `openid-connect`.
  * Configure client settings. Access Type to public. Direct Access Grants Enabled to ON (for this example password will be used directly to generate the token).
  * Add a user to the realm
    * Click the Users menu on the left side of the window.  Click Add user.
    * Type the username `bob`, set the Email Verified switch to ON, and click Save.
    * On the Credentials tab, set the password `p`. Enter the password in both the fields, set the Temporary switch to OFF to avoid the password reset at the next login, and click `Set Password`.

Now, let's run local cluster to test the kuadrantctl new command to generate authpolicy.

* Clone the repo

```bash
git clone https://github.com/Kuadrant/kuadrantctl.git
cd kuadrantctl
```

* Setup cluster, istio and Gateway API CRDs

```bash
make local-setup
```

* Build and install CLI in `bin/kuadrantctl` path

```bash
make install
```

* Install Kuadrant service protection. The CLI can be used to install kuadrant v0.4.1

```bash
bin/kuadrantctl install
```

* Deploy petstore backend API

```bash
kubectl create namespace petstore
kubectl apply -n petstore -f examples/petstore/petstore.yaml
```

* Let's create Petstore's OpenAPI spec

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
  - url: https://example.io/api/v1
paths:
  /cat:
    x-kuadrant:
      backendRefs:
        - name: petstore
          port: 80
          namespace: petstore
    get:  # public (not auth)
      operationId: "getCat"
      responses:
        405:
          description: "invalid input"
  /dog:
    x-kuadrant:
      backendRefs:
        - name: petstore
          port: 80
          namespace: petstore
    get:  # secured
      operationId: "getDog"
      security:
        - openIdConnect: []
      responses:
        405:
          description: "invalid input"
components:
  securitySchemes:
    openIdConnect:
      type: openIdConnect
      openIdConnectUrl: https://${KEYCLOAK_PUBLIC_DOMAIN}/auth/realms/petstore
EOF
```
</details>

> Replace `${KEYCLOAK_PUBLIC_DOMAIN}` with your SSO instance domain

| Operation | Applied config |
| --- | --- |
| `GET /api/v1/cat` | public (not auth) |
| `GET /api/v1/dog` | OIDC authenticatred  |

* Create the HTTPRoute using the CLI
```bash
bin/kuadrantctl generate gatewayapi httproute --oas petstore-openapi.yaml | kubectl apply -n petstore -f -
```

* Create Kuadrant's Auth Policy
```bash
bin/kuadrantctl generate kuadrant authpolicy --oas petstore-openapi.yaml | kubectl apply -n petstore -f -
```

Now, we are ready to test OpenAPI endpoints :exclamation:

- `GET /api/v1/cat` -> It's a public endpoint, hence should return 200 Ok
```bash
curl  -H "Host: example.com" -i "http://127.0.0.1:9080/api/v1/cat"
```
- `GET /api/v1/dog` -> It's a secured endpoint, hence, without credentials, it should return 401
```bash
curl -H "Host: example.com" -i "http://127.0.0.1:9080/api/v1/dog"
```
```
HTTP/1.1 401 Unauthorized
www-authenticate: Bearer realm="getDog"
x-ext-auth-reason: credential not found
date: Tue, 28 Nov 2023 09:38:26 GMT
server: istio-envoy
content-length: 0
```
- Get authentication token. This example is using Direct Access Grants oauth2 grant type (also known as Client Credentials grant type). When configuring the Keycloak (OIDC provider) client settings, we enabled Direct Access Grants to enable this procedure. We will be authenticating as `bob` user with `p` password. We previously created `bob` user in Keycloak in the `petstore` realm.
```
export ACCESS_TOKEN=$(curl -k -H "Content-Type: application/x-www-form-urlencoded" \
        -d 'grant_type=password' \
        -d 'client_id=petstore' \
        -d 'scope=openid' \
        -d 'username=bob' \
        -d 'password=p' "https://${KEYCLOAK_PUBLIC_DOMAIN}/auth/realms/petstore/protocol/openid-connect/token" | jq -r '.access_token')
```
> Replace `${KEYCLOAK_PUBLIC_DOMAIN}` with your SSO instance domain

With the access token in place, let's try to get those puppies

```bash
curl -H "Authorization: Bearer $ACCESS_TOKEN" -H 'Host: example.com' http://127.0.0.1:9080/api/v1/dog -i
```
should return 200 Ok

* Clean environment
```bash
make local-cleanup
```
