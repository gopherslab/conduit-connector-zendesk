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
	"time"
)

const (
	ConfigKeyDomain            = "domain"
	ConfigKeyUserName          = "username"
	ConfigKeyAPIToken          = "apitoken"
	ConfigKeyIterationInterval = "iterationinterval"
	DefaulIterationInterval    = "2m"
)

type Config struct {
	Domain            string
	UserName          string
	APIToken          string
	IterationInterval time.Duration
}

func Parse(cfg map[string]string) (Config, error) {
	userDomain, ok := cfg[ConfigKeyDomain]
	if !ok {
		return Config{}, requiredConfigErr(ConfigKeyDomain)
	}

	userName, ok := cfg[ConfigKeyUserName]
	if !ok {
		return Config{}, requiredConfigErr(ConfigKeyUserName)
	}

	userAPIToken, ok := cfg[ConfigKeyAPIToken]
	if !ok {
		return Config{}, requiredConfigErr(ConfigKeyAPIToken)
	}

	itertionTime := cfg[ConfigKeyIterationInterval]
	if itertionTime == "" {
		itertionTime = DefaulIterationInterval
	}
	interval, err := time.ParseDuration(itertionTime)
	if err != nil {
		return Config{}, fmt.Errorf("%q can't parse time interval", itertionTime)
	}

	config := Config{
		Domain:            userDomain,
		UserName:          userName,
		APIToken:          userAPIToken,
		IterationInterval: interval,
	}

	return config, nil
}

func requiredConfigErr(name string) error {
	return fmt.Errorf("%q config value must be set", name)
}
