# How to cut a new release

## Process

To release a version _“v0.W.Z”_ of the `kuadrantctl` in GitHub, follow these steps:

1. Create annotated and GPG-signed tag

```sh
git tag -s v0.W.Z -m "v0.W.Z"
git push origin v0.W.Z
```

2. In Github, [create release](https://github.com/Kuadrant/kuadrantctl/releases/new).

* Pick recently pushed git tag
* Automatically generate release notes from previous released tag
* Set as the latest release

3. Verify that the build [Release workflow](https://github.com/Kuadrant/kuadrantctl/actions/workflows/release.yaml) is triggered and completes for the new tag

### Verify new release is available

Download `kuadrantctl` binary from [releases](https://github.com/Kuadrant/kuadrantctl/releases) page.
The binary is available in multiple `OS` and `arch`. Pick your option.

```sh
wget https://github.com/Kuadrant/kuadrantctl/releases/download/v0.W.Z/kuadrantctl-v0.W.Z-{OS}-{arch}.tar.gz

tar -zxf kuadrantctl-v0.W.Z-{OS}-{arch}.tar.gz
```

2. Verify version, it should be:

```sh
./kuadrantctl version
```

The output should be the expected v0.W.Z and commitID. For example

```
kuadrantctl v0.3.0 (eec318b2e11e7ea5add5e550ff872bde64555d8f)
```
