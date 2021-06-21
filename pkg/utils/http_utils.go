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
	"io/ioutil"
	"net/http"
	"net/url"
)

// ParseURL returns true when valid HTTP[S] url is found
func ParseURL(str string) (*url.URL, bool) {
	u, err := url.Parse(str)
	return u, err == nil && u.Scheme != "" && u.Host != ""
}

func ReadURL(location *url.URL) ([]byte, error) {
	resp, err := http.Get(location.String())
	if err != nil {
		return nil, err
	}
	data, err := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	if err != nil {
		return nil, err
	}
	return data, nil
}
