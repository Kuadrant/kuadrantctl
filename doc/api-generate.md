## Generate Kuadrant API manifest

Convert a valid [OpenAPI Specification (OAS) 3.x](https://github.com/OAI/OpenAPI-Specification/blob/main/versions/3.0.2.md)
into a Kuadrant [api.networking.kuadrant.io/v1beta1 API](https://github.com/Kuadrant/kuadrant-controller/blob/v0.0.1-pre/apis/networking/v1beta1/api_types.go) kubernetes resource. 

Deploy the [api.networking.kuadrant.io/v1beta1 API](https://github.com/Kuadrant/kuadrant-controller/blob/v0.0.1-pre/apis/networking/v1beta1/api_types.go) kubernetes resource easily:

```
$ kuadrantctl api generate /path/to/oas3.yaml | kubectl apply -f -
```

OpenAPI definition resource can be provided by one of the following channels:
* Filename in the available path.
* URL format (supported schemes are HTTP and HTTPS). The CLI will try to download from the given address.
* Read from stdin standard input stream.

More options:

```shell
kuadrantctl api generate -h
The generate subcommand generates a Kuadrant API manifest from a OAS 3.0 document.
For example:

kuadrantctl api generate oas3-resource (/path/to/your/spec/file.[json|yaml|yml] OR
    http[s]://domain/resource/path.[json|yaml|yml] OR '-')

Outputs to the console by default.

Usage:
  kuadrantctl api generate [flags]

Flags:
  -h, --help            help for generate
  -o, --output string   Write output to <file> instead of stdout

Global Flags:
      --config string   config file (default is $HOME/.kuadrantctl.yaml)
```
