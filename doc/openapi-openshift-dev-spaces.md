# Integrating Kuadrant OAS Extensions with Red Hat OpenShift Dev Spaces

[OpenAPI Extensions](https://swagger.io/docs/specification/openapi-extensions/) enhance the standard OpenAPI specification by adding custom functionality. Kuadrant OpenAPI extensions, identified by the prefix `x-kuadrant`, allow the integration of Kuadrant policies directly within your API specs.

[Red Hat OpenShift Dev Spaces](https://developers.redhat.com/developer-sandbox/ide) offers a browser-based, cloud-native IDE that supports rapid and decentralized development within container-based environments. This tutorial demonstrates how to use OpenShift Dev Spaces to modify an OpenAPI specification by incorporating Kuadrant policies and then employ `kuadrantctl` to create Kubernetes resources for both Gateway API and Kuadrant.


To follow along, you'll need access to a Dev Spaces instance. This can be either:

- A self-hosted instance.
- An instance provided through the [Red Hat Developer Sandbox](https://developers.redhat.com/developer-sandbox/ide).

## Setting up your Workspace

First, create a Workspace in Dev Spaces for your project:

1. Fork the following repository: [https://github.com/Kuadrant/blank-petstore](https://github.com/Kuadrant/blank-petstore).
2. In Dev Spaces, select `Create Workspace`, enter the URL of your forked repository (e.g., `https://github.com/<your-username>/blank-petstore.git`), and then click `Create & Open`.

## Configuring VSCode in Dev Spaces

For this tutorial, we'll:

- Install `kuadrantctl` within your workspace to demonstrate Kubernetes resource generation from your modified OpenAPI spec.
- (Optionally) Configure `git` with your username and email to enable pushing changes back to your repository.


### `kuadrantctl` installation

To install `kuadrantctl` in your Dev Spaces workspace, execute the following command:

```bash
curl -sL "https://github.com/kuadrant/kuadrantctl/releases/download/v0.2.3/kuadrantctl-v0.2.3-linux-amd64.tar.gz" | tar xz -C /home/user/.local/bin
```

This will place `kuadrantctl` in `/home/user/.local/bin`, which is included in the container's `$PATH` by default.

### Configuring Git (Optional)

If you plan to push changes back to your repository, configure your git username and email:

```bash
git config --global user.email "foo@example.com"
git config --global user.name "Foo Example"
```

## Editing Your OpenAPI Spec

Upon creating your workspace, Dev Spaces will launch VSCode loaded with your forked repository. Navigate to the `openapi.yaml` file within the sample app to begin modifications.

### Kuadrant Policies Introduction

We'll enhance our API spec by applying Kuadrant policies to the following endpoints:

`/pet/findByStatus`
`/user/login`
`/store/inventory`

In this tutorial, we're going to introduce some Kuadrant policies via this OAS. We will:

- Generate a `HTTPRoute` to expose these three routes for an existing Gateway API `Gateway`
- Add API key authentication for the `/user/login` route, using Kuadrant's `AuthPolicy` API and OAS' `securitySchemes`
- Add a Kuadrant `RateLimitPolicy` to the `/store/inventory` endpoint, to limit the amount of requests this endpoint can receive

### Defining a Gateway

Utilize the `x-kuadrant` extension in the `info` block to specify a `Gateway`. This information will be used to generate `HTTPRoute`s at the path level:

For example:

```yaml
info:
  x-kuadrant:
    route:  ## HTTPRoute metadata
      name: "petstore"
      namespace: "petstore"
      labels:  ## map[string]string
        deployment: petstore
      hostnames:  ## []gateway.networking.k8s.io/v1beta1.Hostname
        - example.com
      parentRefs:  ## []gateway.networking.k8s.io/v1beta1.ParentReference
        - name: apiGateway
          namespace: gateways
```

Add this extension to the `info` section.


### Specifing `HTTPRoute`'s for each Path

For each path, add an `x-kuadrant` extension with `backendRefs` to link our routes to our paths:


```yaml
  /pet/findByStatus:
    x-kuadrant:
      backendRefs:
      - name: petstore
        namespace: petstore
        port: 8080
    get:
      # ...
```

```yaml
  /user/login:
    x-kuadrant:
      backendRefs:
      - name: petstore
        namespace: petstore
        port: 8080
    get:
      # ...
```

```yaml
  /store/inventory:
    x-kuadrant:
      backendRefs:
      - name: petstore
        namespace: petstore
        port: 8080
    get:
      # ...
```

**Note:** The `x-kuadrant` extension at the path level applies to all HTTP methods defined within. For method-specific policies, move the extension inside the relevant HTTP method block (e.g., `get`, `post`).


### Implementing `AuthPolicy` and Security Schemes

To secure the `/user/login` endpoint with API key authentication, use the following configuration:

```yaml
  /user/login:
    # ...
    get:
      security:
      - api_key: []
```


```yaml
components:
  schemas:
    # ...
  securitySchemes:
    api_key:
      type: apiKey
      name: api_key
      in: header
```

This configuration generates an `AuthPolicy` that references an API key stored in a labeled `Secret`:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: petstore-api-key
  namespace: petstore
  labels:
    authorino.kuadrant.io/managed-by: authorino
    kuadrant.io/apikeys-by: api_key
stringData:
  api_key: secret
type: Opaque
```

We don't recommend using a simple, static API key for your app, but will do so for this tutorial for the sake of simplicity.


### Applying a `RateLimitPolicy` to an Endpoint


To enforce rate limiting on the `/store/inventory` endpoint, add the following `x-kuadrant` extension:

```yaml
  /store/inventory:
    get:
      # ...
      x-kuadrant:
        backendRefs:
          # ...
        rate_limit:
          rates:
          - limit: 10
            duration: 10
            unit: second
```

This limits requests to 10 every 10 seconds for the `/store/inventory` endpoint.

## `kuadrantctl` and Kubernetes resource generation

With our extensions in place, let's use `kuadrantctl` to generate some Kubernetes resources, including:

- An `HTTPRoute` for our petstore app for each of our endpoints
- An `AuthPolicy` with a simple, static API key from a secret for the `/user/login` endpoint
- A `RateLimitPolicy` with a rate limit of 10 requests every 10 seconds for the `/store/inventory` endpoint

Open a new terminal in Dev Spaces (`☰` > `Terminal` > `New Terminal`), and run the following:

```bash
kuadrantctl generate gatewayapi httproute --oas openapi.yaml
```

Outputs:

```yaml
kind: HTTPRoute
apiVersion: gateway.networking.k8s.io/v1beta1
metadata:
  name: petstore
  namespace: petstore
  creationTimestamp: null
  labels:
    deployment: petstore
spec:
  parentRefs:
    - namespace: gateways
      name: apiGateway
  hostnames:
    - example.com
  rules:
    - matches:
        - path:
            type: Exact
            value: /api/v3/pet/findByStatus
          method: GET
      backendRefs:
        - name: petstore
          namespace: petstore
          port: 8080
    - matches:
        - path:
            type: Exact
            value: /api/v3/store/inventory
          method: GET
      backendRefs:
        - name: petstore
          namespace: petstore
          port: 8080
    - matches:
        - path:
            type: Exact
            value: /api/v3/user/login
          method: GET
      backendRefs:
        - name: petstore
          namespace: petstore
          port: 8080
status:
  parents: null
```


```bash
kuadrantctl generate kuadrant authpolicy --oas openapi.yaml
```

Outputs:

```yaml
apiVersion: kuadrant.io/v1beta2
kind: AuthPolicy
metadata:
  name: petstore
  namespace: petstore
  creationTimestamp: null
  labels:
    deployment: petstore
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: petstore
    namespace: petstore
  routeSelectors:
    - matches:
        - path:
            type: Exact
            value: /api/v3/user/login
          method: GET
  rules:
    authentication:
      GETuserlogin_api_key:
        credentials:
          customHeader:
            name: api_key
        apiKey:
          selector:
            matchLabels:
              kuadrant.io/apikeys-by: api_key
        routeSelectors:
          - matches:
              - path:
                  type: Exact
                  value: /api/v3/user/login
                method: GET
status: {}
```


```bash
kuadrantctl generate kuadrant ratelimitpolicy --oas openapi.yaml
```

Outputs:

```yaml
apiVersion: kuadrant.io/v1beta2
kind: RateLimitPolicy
metadata:
  name: petstore
  namespace: petstore
  creationTimestamp: null
  labels:
    deployment: petstore
spec:
  targetRef:
    group: gateway.networking.k8s.io
    kind: HTTPRoute
    name: petstore
    namespace: petstore
  limits:
    GETstoreinventory:
      routeSelectors:
        - matches:
            - path:
                type: Exact
                value: /api/v3/store/inventory
              method: GET
      rates:
        - limit: 10
          duration: 10
          unit: second
status: {}
```

## Applying resources

> **Note:** by default, `oc` and `kubectl` in Dev Spaces will target the cluster running Dev Spaces. If you want to apply resources to another cluster, you will need to login with `oc` or `kubectl` to another cluster, and pass a different `--context` to these to apply resources to another cluster.

You can now apply these policies to a running app via `kubectl` or `oc`. If Dev Spaces is running on a cluster where Kuadrant is also installed, you can apply these resources:


```bash
kuadrantctl generate gatewayapi httproute --oas openapi.yaml | kubectl apply -f -
kuadrantctl generate kuadrant authpolicy --oas openapi.yaml | kubectl apply -f -
kuadrantctl generate kuadrant ratelimitpolicy --oas openapi.yaml | kubectl apply -f -
```

Alternatively, `kuadrantctl` can be used as part of a CI/CD pipeline. See the [kuadrantctl CI/CD guide](./kuadrantctl-ci-cd.md) for more details.

If you've completed the optional `git` configuration step above, you can now `git commit` the changes above and push these to your fork.

# Next

Here are some extra documentation on using `x-kuadrant` OAS extensions with `kuadrantctl`:

- [Guide to `kuadrantctl` and OAS extensions](./openapi-kuadrant-extensions.md)
- [Generating Gateway API HTTPRoutes with `kuadrantctl`](./generate-gateway-api-httproute.md)
- [Generating Kuadrant AuthPolicy with `kuadrantctl`](./generate-kuadrant-auth-policy.md)
- [Generate Kuadrant RateLimitPolicy with `kuadrantctl`](./generate-kuadrant-rate-limit-policy.md)
- [`kuadrantctl` CI/CD guide](./kuadrantctl-ci-cd.md)
