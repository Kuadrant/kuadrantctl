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
package utils

import (
	"fmt"
	"testing"
)

func TestIsURL(t *testing.T) {
	cases := []struct {
		url      string
		expected bool
	}{
		{"", false},
		{"https", false},
		{"https://", false},
		{"http://www", true},
		{"http://www.example.com/resources/a.yaml", true},
		{"https://www.example.com:443/resources/a.yaml", true},
		{"/home/testing-path.yaml", false},
		{"testing-path.yaml", false},
		{"alskjff#?asf//dfas", false},
	}

	for i, tc := range cases {
		t.Run(fmt.Sprintf("test-%d", i), func(subT *testing.T) {
			_, res := ParseURL(tc.url)
			if res != tc.expected {
				subT.Errorf("\"%s\": Expecting from IsURL %t and got %t", tc.url, tc.expected, res)
			}
		})
	}
}
