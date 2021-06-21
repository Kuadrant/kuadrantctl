## Install Kuadrant

The install command applies kuadrant manifest bundle and applies it to a cluster.

Install command does:

* Creates kuadrant-system namespace (currently namespace name is hardcoded)
* Deploy ingress controller (currently istio 1.9.4)
* Deploy auth provider (currently Authorino)
* Deploy kuadrant manifests and controller
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
