## Release

The release process follows a streamlined approach, no release branches involved.
New releases can be major, minor or patch based releases, but always incrementing digits
regarding the latest release version.

### New Major.Minor.Patch version

1. Create a new minor release branch from the HEAD of main:
```sh
git checkout -b release-vX.Y.Z
```
2. Update version (prefixed with **"v"**):
```sh
make prepare-release VERSION=vX.Y.Z
```
3. Verify local changes:
```sh
make install
bin/kuadrantctl version
```
The output should be the new version, for example :
```
kuadrantctl v0.3.0 (ff779a1-dirty)
```
4. Commit and push:
```sh
git add .
git commit -m "prepare-release: release-vX.Y.Z"
git push origin release-vX.Y.Z
```
5. Create git tag:
```sh
git tag -s -m vX.Y.Z vX.Y.Z
git push origin vX.Y.Z
```
6. In Github, [create release](https://github.com/Kuadrant/kuadrantctl/releases/new).

* Pick recently pushed git tag
* Automatically generate release notes from previous released tag
* Set as the latest release

7. Verify that the build [Release workflow](https://github.com/Kuadrant/kuadrantctl/actions/workflows/release.yaml) is triggered and completes for the new tag

### Verify new release is available

1. Download the latest binary for your platform from the [`kuadrantctl` Latest Releases](https://github.com/Kuadrant/kuadrantctl/releases/latest) page.
2. Unpack the binary.
3. Move it to a directory in your `$PATH` so that it can be executed from anywhere.
4. Check the version:
```sh
kuadrantctl version
```
The output should be the new version, for example :
```
kuadrantctl v0.3.0 (eec318b2e11e7ea5add5e550ff872bde64555d8f)
```
