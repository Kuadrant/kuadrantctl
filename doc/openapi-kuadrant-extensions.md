## OpenAPI 3.0.X Kuadrant Extensions

### Info level kuadrant extension

Kuadrant extension that can be added at the info level of the OpenAPI spec.

```yaml
info:
  x-kuadrant:
    route:  ## HTTPRoute metadata
      name: "petstore"
      namespace: "petstore"
      hostnames:  ## []gateway.networking.k8s.io/v1beta1.Hostname
        - example.com
      parentRefs:  ## []gateway.networking.k8s.io/v1beta1.ParentReference
        - name: apiGateway
          namespace: gateways
```

### Path level kuadrant extension

Kuadrant extension that can be added at the path level of the OpenAPI spec.
This configuration at the path level
is the default when there is no operation level configuration.

```yaml
paths:
  /cat:
    x-kuadrant:  ## Path level Kuadrant Extension
      enable: true  ## Add to the HTTPRoute. Optional. Default: false
      backendRefs:  ## Backend references to be included in the HTTPRoute. []gateway.networking.k8s.io/v1beta1.HTTPBackendRef. Optional.
        - name: petstore
          port: 80
          namespace: petstore
      rate_limit:  ## Rate limit config. Optional.
        rates:   ## Kuadrant API []github.com/kuadrant/kuadrant-operator/api/v1beta2.Rate
          - limit: 1
            duration: 10
            unit: second
        counters:   ## Kuadrant API []github.com/kuadrant/kuadrant-operator/api/v1beta2.CountextSelector
          - auth.identity.username
        when:   ## Kuadrant API []github.com/kuadrant/kuadrant-operator/api/v1beta2.WhenCondition
          - selector: metadata.filter_metadata.envoy\.filters\.http\.ext_authz.identity.userid
            operator: eq
            value: alice
```

### Operation level kuadrant extension

Kuadrant extension that can be added at the operation level of the OpenAPI spec.
Same schema as path level kuadrant extension.

```yaml
paths:
  /cat:
    get:
      x-kuadrant:  ## Path level Kuadrant Extension
        enable: true  ## Add to the HTTPRoute. Optional. Default: false
        backendRefs:  ## Backend references to be included in the HTTPRoute. Optional.
          - name: petstore
            port: 80
            namespace: petstore
        rate_limit:  ## Rate limit config. Optional.
          rates:   ## Kuadrant API github.com/kuadrant/kuadrant-operator/api/v1beta2.Rate
            - limit: 1
              duration: 10
              unit: second
          counters:   ## Kuadrant API github.com/kuadrant/kuadrant-operator/api/v1beta2.CountextSelector
            - auth.identity.username
          when:   ## Kuadrant API github.com/kuadrant/kuadrant-operator/api/v1beta2.WhenCondition
            - selector: metadata.filter_metadata.envoy\.filters\.http\.ext_authz.identity.userid
              operator: eq
              value: alice
```
