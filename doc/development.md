# Development Guide

## Technology stack required for development

* [git][git_tool]
* [go] version 1.20+

## Build the CLI
```
$ git clone https://github.com/kuadrant/kuadrantctl.git
$ cd kuadrantctl && make install
$ bin/kuadrantctl version
{"level":"info","ts":1699437585.7809818,"msg":"kuadrantctl version: latest"}
{"level":"info","ts":1699437585.7941456,"msg":"Istio version: docker.io/istio/pilot:1.12.1"}
{"level":"info","ts":1699437585.798138,"msg":"Authorino operator version: quay.io/3scale/authorino-operator:latest"}
{"level":"info","ts":1699437585.798147,"msg":"Authorino version: quay.io/3scale/authorino:v0.7.0"}
{"level":"info","ts":1699437585.7990465,"msg":"Limitador operator version: quay.io/kuadrant/limitador-operator:main"}
{"level":"info","ts":1699437585.799057,"msg":"Limitador version: 0.4.0"}
{"level":"info","ts":1699437585.8007147,"msg":"Kuadrant controller version: quay.io/kuadrant/kuadrant-controller:main"}
```

## Quick steps to contribute

* Fork the project.
* Download your fork to your PC (`git clone https://github.com/your_username/kuadrantctl && cd kuadrantctl`)
* Create your feature branch (`git checkout -b my-new-feature`)
* Make changes and run tests (`make test`)
* Add them to staging (`git add .`)
* Commit your changes (`git commit -m 'Add some feature'`)
* Push to the branch (`git push origin my-new-feature`)
* Create new pull request

[git_tool]:https://git-scm.com/downloads
[go]:https://golang.org/
