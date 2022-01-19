## Install Kuadrant

The install command applies kuadrant manifest bundle and applies it to a cluster.

Components being installed:

* `kuadrant-system` namespace
* [istio](https://istio.io/) 1.12.1
* Authentication/Authorization Service
  * [Authorino Operator](https://github.com/kuadrant/authorino-operator) v0.1.0
  * [Authorino](https://github.com/Kuadrant/authorino) v0.7.0
* Rate Limit Service
  * [Limitador Operator](https://github.com/kuadrant/limitador-operator) v0.2.0
  * [Limitador](https://github.com/kuadrant/limitador) v0.4.0
* [kuadrant controller](https://github.com/Kuadrant/kuadrant-controller) v0.2.0

### Usage :

```shell
$ kuadrantctl install --help
The install command applies kuadrant manifest bundle and applies it to a cluster.

Usage:
  kuadrantctl install [flags]

Flags:
  -h, --help                help for install
      --kubeconfig string   Kubernetes configuration file

Global Flags:
      --config string   config file (default is $HOME/.kuadrantctl.yaml)
```
