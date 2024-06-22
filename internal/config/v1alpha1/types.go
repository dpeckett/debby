// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package v1alpha1

import (
	"fmt"

	"github.com/dpeckett/debby/internal/config/types"
)

const APIVersion = "debby/v1alpha1"

type Config struct {
	types.TypeMeta `yaml:",inline"`
	// Contents is the contents of the base system.
	Contents ContentsConfig `yaml:"contents"`
}

type ContentsConfig struct {
	// Sources is a list of apt repositories to use for downloading packages.
	Sources []SourceConfig `yaml:"sources"`
	// Keyring is a list of public key URLs or files to use for verifying packages.
	Keyring []string `yaml:"keyring,omitempty"`
	// Packages is a list of packages to install.
	Packages []string `yaml:"packages"`
}

// SourceConfig is the configuration for an apt repository.
type SourceConfig struct {
	// URL is the URL of the repository.
	URL string `yaml:"url"`
	// Distribution specifies the Debian distribution name (e.g., bullseye, buster)
	// or class (e.g., stable, testing). If not specified, defaults to "stable".
	Distribution string `yaml:"distribution,omitempty"`
	// Components is a list of components to use from the repository.
	// If not specified, defaults to ["main"].
	Components []string `yaml:"components,omitempty"`
}

func (c *Config) GetAPIVersion() string {
	return APIVersion
}

func (c *Config) GetKind() string {
	return "Config"
}

func (c *Config) PopulateTypeMeta() {
	c.TypeMeta = types.TypeMeta{
		APIVersion: APIVersion,
		Kind:       "Config",
	}
}

func GetConfigByKind(kind string) (types.Config, error) {
	switch kind {
	case "Config":
		return &Config{}, nil
	default:
		return nil, fmt.Errorf("unsupported kind: %s", kind)
	}
}
