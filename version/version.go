/*
Copyright 2021 Red Hat, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package version

var (
	// This variable shows the commit hash of the repo so they can confirm they are on latest or specific version of the branch.
	GitHash string
	// This variable is dependent on what the current release is e.g. if it is v0.2.3 then this variable, outside of releases, will be v0.2.4-dev .
	Version = "v0.2.4-alpha1"
)
