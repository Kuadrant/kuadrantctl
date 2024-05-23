# kuadrantctl
[![License](https://img.shields.io/badge/license-Apache--2.0-blue.svg)](http://www.apache.org/licenses/LICENSE-2.0)

`kuadrantctl` is a CLI tool for managing [Kuadrant](https://kuadrant.io/) configurations and resources.

## Installing

`kuadrantctl` can be installed either by downloading pre-compiled binaries or by compiling from source. For most users, downloading the binary is the easiest and recommended method.

### Installing Pre-compiled Binaries

1. Download the latest binary for your platform from the [`kuadrantctl` Releases](https://github.com/Kuadrant/kuadrantctl/releases) page.
2. Unpack the binary.
3. Move it to a directory in your `$PATH` so that it can be executed from anywhere.

### Compiling from Source

If you prefer to compile from source or are contributing to the project, you can install `kuadrantctl` using `go install` or `make install`. This method requires Golang 1.21 or newer.

```bash
go install github.com/kuadrant/kuadrantctl@latest
```

This command will compile `kuadrantctl` and install the binary executable in `$GOBIN` (defaulting to `$GOPATH/bin`).

It is also possible to use the make target `install` to compile from source. From root of the repository, run 

```bash
make install
```

This will compile `kuadrantctl` and install it in the `bin` directory at root of directory. It will also ensure the correct version of the binary is displayed, rather than `v0.0.0` . It can be ran using `./bin/kuadrantctl` .  

## Usage

 Below is a high-level overview of its commands, along with links to detailed documentation for more complex commands.

### General Syntax

```bash
kuadrantctl [command] [subcommand] [flags]
```


### Commands Overview

| Command      | Description                                                |
| ------------ | ---------------------------------------------------------- |
| `completion` | Generate autocompletion scripts for the specified shell    |
| `generate`   | Commands related to Kubernetes Gateway API and Kuadrant resource generation from OpenAPI 3.x specifications          |
| `help`       | Help about any command                                     |
| `version`    | Print the version number of `kuadrantctl`                    |

### Flags

| Flag               | Description           |
| ------------------ | --------------------- |
| `-h`, `--help`     | Help for `kuadrantctl`  |
| `-v`, `--verbose`  | Enable verbose output |

### Commands Detail

#### `completion`

Generate an autocompletion script for the specified shell.

| Subcommand   | Description                                 |
| ------------ | ------------------------------------------- |
| `bash`       | Generate script for Bash                    |
| `fish`       | Generate script for Fish                    |
| `powershell` | Generate script for PowerShell              |
| `zsh`        | Generate script for Zsh                     |

#### `generate`

Commands related to Kubernetes Gateway API and Kuadrant resource generation from OpenAPI 3.x specifications.

| Subcommand   | Description                                   |
| ------------ | --------------------------------------------- |
| `gatewayapi` | Generate Gateway API resources                |
| `kuadrant`   | Generate Kuadrant resources                   |

##### `generate gatewayapi`

Generate Gateway API resources from an OpenAPI 3.x specification

| Subcommand | Description                                      | Flags                             |
| ---------- | ------------------------------------------------ | --------------------------------- |
| `httproute`| Generate Gateway API HTTPRoute from OpenAPI 3.0.X| `--oas string` Path to OpenAPI spec file (in JSON or YAML format), URL, or '-' to read from standard input (required). `-o` Output format: 'yaml' or 'json'. (default "yaml") |

##### `generate kuadrant`

Generate Kuadrant resources from an OpenAPI 3.x specification

| Subcommand       | Description                                       | Flags                             |
| ---------------- | ------------------------------------------------- | --------------------------------- |
| `authpolicy`     | Generate a [Kuadrant AuthPolicy](https://docs.kuadrant.io/kuadrant-operator/doc/auth/) from an OpenAPI 3.0.x specification   | `--oas string` Path to OpenAPI spec file (in JSON or YAML format), URL, or '-' to read from standard input (required). `-o` Output format: 'yaml' or 'json'. (default "yaml") |
| `ratelimitpolicy`| Generate [Kuadrant RateLimitPolicy](https://docs.kuadrant.io/kuadrant-operator/doc/rate-limiting/) from an OpenAPI 3.0.x specification | `--oas string` Path to OpenAPI spec file (in JSON or YAML format), URL, or '-' to read from standard input (required). `-o` Output format: 'yaml' or 'json'. (default "yaml") |


#### `version`

Print the version number of `kuadrantctl`.

No additional flags or subcommands.

### Additional Guides

#### Generating Gateway API HTTPRoute Objects

- Generates [Gateway API HTTPRoute](https://gateway-api.sigs.k8s.io/v1alpha2/guides/http-routing/) objects from an OpenAPI Specification (OAS) 3.x.
- Supports reading from a file, URL, or stdin.
- Example usages and more information can be found in the [detailed guide](doc/generate-gateway-api-httproute.md).

#### Generating Kuadrant AuthPolicy Objects

- Generates [Kuadrant AuthPolicy](https://github.com/Kuadrant/kuadrant-operator/blob/v0.4.1/doc/auth.md) objects for managing API authentication.
- Supports `openIdConnect` and `apiKey` types from the OpenAPI Security Scheme Object.
- Example usages and more information can be found in the [detailed guide](doc/generate-kuadrant-auth-policy.md).

#### Generating Kuadrant RateLimitPolicy Objects

- Generates [Kuadrant RateLimitPolicy](https://github.com/Kuadrant/kuadrant-operator/blob/v0.4.1/doc/rate-limiting.md) objects for managing API rate limiting.
- Supports reading from a file, URL, or stdin.
- Example usages and more information can be found in the [detailed guide](doc/generate-kuadrant-rate-limit-policy.md).

For more detailed information about each command, including options and usage examples, use `kuadrantctl [command] --help`.


## Using with GitHub Actions

```yaml
- name: Install kuadrantctl
  uses: jaxxstorm/action-install-gh-release@v1.10.0
  with: # Grab the latest version
    repo: Kuadrant/kuadrantctl
```

## Commands
* [Generate Gateway API HTTPRoute objects from OpenAPI 3.X](doc/generate-gateway-api-httproute.md)
* [Generate Kuadrant RateLimitPolicy from OpenAPI 3.X](doc/generate-kuadrant-rate-limit-policy.md)
* [Generate Kuadrant AuthPolicy from OpenAPI 3.X](doc/generate-kuadrant-auth-policy.md)

## Contributing
The [Development guide](doc/development.md) describes how to build the kuadrantctl CLI and how to test your changes before submitting a patch or opening a PR.

## Licensing

This software is licensed under the [Apache 2.0 license](https://www.apache.org/licenses/LICENSE-2.0).

See the LICENSE and NOTICE files that should have been provided along with this software for details.
