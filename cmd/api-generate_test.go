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
	"io/ioutil"
	"os"
	"testing"

	kctlrv1beta1 "github.com/kuadrant/kuadrant-controller/apis/networking/v1beta1"
)

func TestAPIGenerateBasicType(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "testsample")
	ok(t, err)
	defer os.Remove(tempFile.Name())

	rootCmd.SetArgs([]string{"api", "generate", "-o", tempFile.Name(), "testdata/petstore.yaml"})
	ok(t, rootCmd.Execute())

	serializedCommandOut, err := ioutil.ReadFile(tempFile.Name())
	ok(t, err)

	api := &kctlrv1beta1.API{}
	ok(t, json.Unmarshal(serializedCommandOut, api))

	equals(t, "API", api.TypeMeta.Kind)
	equals(t, kctlrv1beta1.GroupVersion.String(), api.TypeMeta.APIVersion)
}

func TestAPIGenerateBasicHosts(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "testsample")
	ok(t, err)
	defer os.Remove(tempFile.Name())

	rootCmd.SetArgs([]string{"api", "generate", "-o", tempFile.Name(), "testdata/petstore.yaml"})
	ok(t, rootCmd.Execute())

	serializedCommandOut, err := ioutil.ReadFile(tempFile.Name())
	ok(t, err)

	api := &kctlrv1beta1.API{}
	ok(t, json.Unmarshal(serializedCommandOut, api))

	equals(t, 1, len(api.Spec.Hosts))
	equals(t, "petstore.swagger.io", api.Spec.Hosts[0])
}

func TestAPIGenerateBasicOperations(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "testsample")
	ok(t, err)
	defer os.Remove(tempFile.Name())

	rootCmd.SetArgs([]string{"api", "generate", "-o", tempFile.Name(), "testdata/petstore.yaml"})
	ok(t, rootCmd.Execute())

	serializedCommandOut, err := ioutil.ReadFile(tempFile.Name())
	ok(t, err)

	api := &kctlrv1beta1.API{}
	ok(t, json.Unmarshal(serializedCommandOut, api))

	equals(t, 3, len(api.Spec.Operations))

	findOperation := func(ops []*kctlrv1beta1.Operation, name string) *kctlrv1beta1.Operation {
		for idx := range ops {
			if ops[idx].Name == name {
				return ops[idx]
			}
		}
		return nil
	}

	operations := []struct {
		name   string
		path   string
		method string
	}{
		{"listpets", "/v1/pets", "GET"},
		{"createpets", "/v1/pets", "POST"},
		{"showpetbyid", "/v1/pets/{petId}", "GET"},
	}

	for _, expectedOp := range operations {
		op := findOperation(api.Spec.Operations, expectedOp.name)
		if op == nil {
			t.Fatalf("operation {%s} not found", expectedOp.name)
		}
		equals(t, expectedOp.path, op.Path)
		equals(t, expectedOp.method, op.Method)
	}
}

func TestAPIGenerateBasicSecSchemes(t *testing.T) {
	tempFile, err := ioutil.TempFile("", "testsample")
	ok(t, err)
	defer os.Remove(tempFile.Name())

	rootCmd.SetArgs([]string{"api", "generate", "-o", tempFile.Name(), "testdata/petstore.yaml"})
	ok(t, rootCmd.Execute())

	serializedCommandOut, err := ioutil.ReadFile(tempFile.Name())
	ok(t, err)

	api := &kctlrv1beta1.API{}
	ok(t, json.Unmarshal(serializedCommandOut, api))

	equals(t, 3, len(api.Spec.SecurityScheme))

	findSecScheme := func(schs []*kctlrv1beta1.SecurityScheme, name string) *kctlrv1beta1.SecurityScheme {
		for idx := range schs {
			if schs[idx].Name == name {
				return schs[idx]
			}
		}
		return nil
	}

	secSchemes := []struct {
		name             string
		schType          string
		apiKeyLocation   string
		apiKeyName       string
		openIDConnectURL string
	}{
		{"apiKey", "apiKey", "header", "X-API-KEY", ""},
		{"appId", "apiKey", "header", "X-APP-ID", ""},
		{"openIdConnect", "openIdConnect", "", "", "https://example.com/.well-known/openid-configuration"},
	}

	for _, expectedSch := range secSchemes {
		secScheme := findSecScheme(api.Spec.SecurityScheme, expectedSch.name)
		assert(t, secScheme != nil, "secScheme {%s} not found", expectedSch.name)
		switch expectedSch.schType {
		case "apiKey":
			assert(t, secScheme.APIKeyAuth != nil, "secScheme {%s} apiKey scheme expected and not parsed", expectedSch.name)
			equals(t, expectedSch.apiKeyLocation, secScheme.APIKeyAuth.Location)
			equals(t, expectedSch.apiKeyName, secScheme.APIKeyAuth.Name)
		case "openIdConnect":
			assert(t, secScheme.OpenIDConnectAuth != nil, "secScheme {%s} openIdConnect scheme expected and not parsed", expectedSch.name)
			equals(t, expectedSch.openIDConnectURL, secScheme.OpenIDConnectAuth.URL)
		default:
			assert(t, false, "secScheme {%s} has unknown type %s", expectedSch.name, expectedSch.schType)
		}
	}
}
