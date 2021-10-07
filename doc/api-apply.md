## Apply Kuadrant API objects

The [kuadrant API CRD](https://github.com/Kuadrant/kuadrant-controller/blob/v0.1.1/apis/networking/v1beta1/api_types.go) represent internal APIs bundled in a API product.
The kuadrant API CRD grant API Providers the freedom to map their internal API organization structure to kuadrant.

The `kuadrantctl api apply` command allows easily to create and update existing *kuadrant API* custom resources.

A prior condition before the *Kuadrant API custom resource* can be created is that the API Provider
must have a [kubernetes service](https://kubernetes.io/docs/concepts/services-networking/service/) created for the API being protected.
This command will create (or update) one [kuadrant API](https://github.com/Kuadrant/kuadrant-controller/blob/v0.1.1/apis/networking/v1beta1/api_types.go)
custom resource in the same namespace as the referenced service.

The `kuadrantctl api apply` command needs to have some info about the API being exposed. It can be either:
* A valid [OpenAPI Specification (OAS) 3.x](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.2.md)
* A match path. It can be a single specific path, prefix or regex

### OpenAPI specification

OpenAPI document resource can be provided by one of the following channels:
* Filename in the available path.
* URL format (supported schemes are HTTP and HTTPS). The CLI will try to download from the given address.
* Read from stdin standard input stream.

### Usage :

```shell
$ kuadrantctl api apply -h
The apply command allows easily to create and update existing *kuadrant API* custom resources

Usage:
  kuadrantctl api apply [flags]

Flags:
      --api-name string          If not set, the name of the API can be matched with the service name
  -h, --help                     help for apply
      --kubeconfig string        Kubernetes configuration file
      --match-path string        Define a single specific path, prefix or regex (default "/")
      --match-path-type string   Specifies how to match against the matchpath value. Accepted values are Exact, Prefix and RegularExpression. Defaults to Prefix (default "Prefix")
  -n, --namespace string         Service namespace (required)
      --oas string               /path/to/file.[json|yaml|yml] OR http[s]://domain/resource/path.[json|yaml|yml] OR -
      --port string              Only required if there are multiple ports in the service. Either the Name of the port or the Number
      --scheme string            Either HTTP or HTTPS specifies how the kuadrant gateway will connect to this API (default "http")
      --service-name string      Service name (required)
      --tag string               A special tag used to distinguish this deployment between several instances of the API
      --to-stdout                Serialize the kuadrant API object in stdout instead of applying to the cluster

Global Flags:
      --config string   config file (default is $HOME/.kuadrantctl.yaml)
  -v, --verbose         verbose output
```
