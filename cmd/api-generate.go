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
	"encoding/json"
	"os"

	kctlrv1beta1 "github.com/kuadrant/kuadrant-controller/apis/networking/v1beta1"
	"github.com/spf13/cobra"
)

var apiGenerateOutputFlag string

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
		api := &kctlrv1beta1.API{}

		// Using json serializer because
		// currently (github.com/kuadrant/kuadrant-controller#0.0.1-pre) API data type
		// does not have yaml tags, only json tags
		serializedAPI, err := json.MarshalIndent(api, "", "  ")
		if err != nil {
			return err
		}

		writer := cmd.OutOrStdout()
		if apiGenerateOutputFlag != "" {
			file, err := os.Create(apiGenerateOutputFlag)
			if err != nil {
				return err
			}
			defer file.Close()
			writer = file
		}

		_, err = writer.Write(serializedAPI)
		if err != nil {
			return err
		}

		return nil
	},
}

func init() {
	apiGenerateCmd.PersistentFlags().StringVarP(&apiGenerateOutputFlag, "output", "o", "", "Write output to <file> instead of stdout")
	apiCmd.AddCommand(apiGenerateCmd)
}
