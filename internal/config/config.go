// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package config

import (
	"fmt"
	"io"

	configtypes "github.com/dpeckett/debby/internal/config/types"
	latestconfig "github.com/dpeckett/debby/internal/config/v1alpha1"
	"gopkg.in/yaml.v3"
)

// FromYAML reads the given reader and returns a config object.
func FromYAML(r io.Reader) (*latestconfig.Config, error) {
	confBytes, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read config from reader: %w", err)
	}

	var typeMeta configtypes.TypeMeta
	if err := yaml.Unmarshal(confBytes, &typeMeta); err != nil {
		return nil, fmt.Errorf("failed to unmarshal type meta from config file: %w", err)
	}

	var versionedConf configtypes.Config
	switch typeMeta.APIVersion {
	case latestconfig.APIVersion:
		versionedConf, err = latestconfig.GetConfigByKind(typeMeta.Kind)
	default:
		return nil, fmt.Errorf("unsupported api version: %s", typeMeta.APIVersion)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get config by kind %q: %w", typeMeta.Kind, err)
	}

	if err := yaml.Unmarshal(confBytes, versionedConf); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config from config file: %w", err)
	}

	versionedConf, err = MigrateToLatest(versionedConf)
	if err != nil {
		return nil, fmt.Errorf("failed to migrate config: %w", err)
	}

	return versionedConf.(*latestconfig.Config), nil
}

// ToYAML writes the given config object to the given writer.
func ToYAML(w io.Writer, versionedConf configtypes.Config) error {
	versionedConf.PopulateTypeMeta()

	if err := yaml.NewEncoder(w).Encode(versionedConf); err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	return nil
}

// MigrateToLatest migrates the given config object to the latest version.
func MigrateToLatest(versionedConf configtypes.Config) (configtypes.Config, error) {
	switch conf := versionedConf.(type) {
	case *latestconfig.Config:
		// Nothing to do, already at the latest version.
		return conf, nil
	default:
		return nil, fmt.Errorf("unsupported config version: %s", conf.GetAPIVersion())
	}
}
