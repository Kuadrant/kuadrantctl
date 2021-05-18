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
package cmd

import (
	"fmt"

	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v2"

	"github.com/kuadrant/kuadrantctl/pkg/kuadrantapi"
	"github.com/kuadrant/kuadrantctl/pkg/utils"
)

// apiGenerateCmd represents the generate command
var apiGenerateCmd = &cobra.Command{
	Use:   "generate",
	Short: "Generates a Kuadrant API manifest",
	Long: `The generate subcommand generates a Kuadrant API manifest from a OAS 3.0 document.
For example:

kuadrantctl api generate oas3-resource (/path/to/your/spec/file.[json|yaml|yml] OR
    http[s]://domain/resource/path.[json|yaml|yml] OR '-')

Outputs to the console by default.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		data, err := utils.ReadExternalResource(args[0])
		if err != nil {
			return err
		}

		loader := openapi3.NewLoader()
		doc, err := loader.LoadFromData(data)
		if err != nil {
			return err
		}

		// TODO(eastizle): optional flag for validation
		err = doc.Validate(loader.Context)
		if err != nil {
			return err
		}

		apiLoader := kuadrantapi.NewLoader()
		api, err := apiLoader.LoadFromDoc(doc)
		if err != nil {
			return err
		}

		serializedAPI, err := yaml.Marshal(api)
		if err != nil {
			return err
		}

		fmt.Println(string(serializedAPI))

		return nil
	},
}

func init() {
	apiCmd.AddCommand(apiGenerateCmd)
}
