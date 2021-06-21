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
package authorinomanifests

import (
	"embed"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
)

// Content holds authorino manifests
//go:embed authorino.yaml
var content embed.FS

func Content() ([]byte, error) {
	logf.Log.Info("Resource file", "name", "authorino.yaml")
	return content.ReadFile("authorino.yaml")
}
