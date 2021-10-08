## Install Kuadrant

The install command applies kuadrant manifest bundle and applies it to a cluster.

Install command does:

* Creates kuadrant-system namespace (currently namespace name is hardcoded)
* Deploy ingress controller (currently [istio](https://istio.io/) 1.9.4)
* Deploy auth provider (currently [Authorino](https://github.com/Kuadrant/authorino) v0.4.0)
* Deploy rate limit provider (currently [Limitador Operator](https://github.com/kuadrant/limitador-operator) v0.2.0)
* Deploy [kuadrant controller](https://github.com/Kuadrant/kuadrant-controller) (currently v0.1.1)
* Waits for deployment availabilty

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
