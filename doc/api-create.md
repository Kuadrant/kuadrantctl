## Create Kuadrant API manifest

Convert a valid [OpenAPI Specification (OAS) 3.x](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.2.md)
into a Kuadrant [api.networking.kuadrant.io/v1beta1 API](https://github.com/Kuadrant/kuadrant-controller/blob/v0.0.1-pre/apis/networking/v1beta1/api_types.go) kubernetes resource and install it in your cluster.  

OpenAPI definition resource can be provided by one of the following channels:
* Filename in the available path.
* URL format (supported schemes are HTTP and HTTPS). The CLI will try to download from the given address.
* Read from stdin standard input stream.

### Limitations
* Supported security schemes: **apiKey**, **openIdConnect**.

### Usage :

```shell
$ kuadrantctl api create -h
The create command generates a Kuadrant API manifest and applies it to a cluster.
For example:

kuadrantctl api create oas3-resource -n ns (/path/to/your/spec/file.[json|yaml|yml] OR
    http[s]://domain/resource/path.[json|yaml|yml] OR '-')

Usage:
  kuadrantctl api create [flags]

Flags:
  -h, --help                help for create
      --kubeconfig string   Kubernetes configuration file
  -n, --namespace string    Cluster namespace (required)

Global Flags:
      --config string   config file (default is $HOME/.kuadrantctl.yaml)
```
