/*
Copyright © 2022 Meroxa, Inc.

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

package config

import (
	"fmt"
)

const (
	KeyDomain   = "zendesk.domain"
	KeyUserName = "zendesk.userName"
	KeyAPIToken = "zendesk.apiToken" //nolint:gosec //we are not hard coding the credentials
)

type Config struct {
	Domain   string
	UserName string
	APIToken string
}

// Parse validate zendesk basic token authentication
func Parse(cfg map[string]string) (Config, error) {
	userDomain, ok := cfg[KeyDomain]
	if !ok {
		return Config{}, requiredConfigErr(KeyDomain)
	}

	userName, ok := cfg[KeyUserName]
	if !ok {
		return Config{}, requiredConfigErr(KeyUserName)
	}

	userAPIToken, ok := cfg[KeyAPIToken]
	if !ok {
		return Config{}, requiredConfigErr(KeyAPIToken)
	}

	config := Config{
		Domain:   userDomain,
		UserName: userName,
		APIToken: userAPIToken,
	}

	return config, nil
}

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}
