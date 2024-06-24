// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package types

import (
	"strings"

	"github.com/dpeckett/debby/internal/types/arch"
	"github.com/dpeckett/debby/internal/types/dependency"
	"github.com/dpeckett/debby/internal/types/list"
	"github.com/dpeckett/debby/internal/types/version"
)

// Package represents a Debian package.
type Package struct {
	// Package is the name of the package.
	Package string
	// Source is the source package name.
	Source string
	// Version is the version of the package.
	Version version.Version
	// InstalledSize is the installed size of the package, in kilobytes.
	InstalledSize int `json:"Installed-Size,string"`
	// Maintainer is the person or organization responsible for the package.
	Maintainer string
	// Architecture is the architecture the package is built for.
	Architecture arch.Arch
	// Replaces lists packages that this package replaces.
	Replaces dependency.Dependency
	// Breaks lists packages that this package breaks.
	Breaks dependency.Dependency
	// Provides lists virtual packages that this package provides.
	Provides dependency.Dependency
	// Conflicts lists packages that conflict with this package.
	Conflicts dependency.Dependency
	// Enhances lists packages that this package enhances.
	Enhances dependency.Dependency
	// Depends lists packages that this package depends on.
	Depends dependency.Dependency
	// Recommends lists packages that are recommended to be installed with this package.
	Recommends dependency.Dependency
	// Suggests lists packages that are suggested to be installed with this package.
	Suggests dependency.Dependency
	// PreDepends lists packages that must be installed and configured before this package.
	PreDepends dependency.Dependency `json:"Pre-Depends"`
	// Description provides a short description of the package.
	Description string
	// Homepage is the URL of the package's homepage.
	Homepage string
	// Tag lists tags associated with the package.
	Tag list.CommaDelimited[string]
	// Section categorizes the package within the archive.
	Section string
	// Priority defines the importance of the package.
	Priority string
	// Filename is the name of the package file.
	Filename string
	// Size is the size of the package file, in bytes.
	Size int `json:",string"`
	// SHA256 is the SHA-256 checksum of the package file.
	SHA256 string

	// Additional fields that are not part of the standard control file but are
	// used internally by debby.

	// URLs is a list of URLs that the package can be downloaded from.
	URLs []string `json:"X-URLs"`
}

// ID returns a unique identifier for the package.
func (p Package) ID() string {
	return p.Package + "_" + p.Version.String() + "_" + p.Architecture.String()
}

func (a Package) Compare(b Package) int {
	// Compare package names.
	if cmp := strings.Compare(a.Package, b.Package); cmp != 0 {
		return cmp
	}

	// Compare package versions.
	if cmp := a.Version.Compare(b.Version); cmp != 0 {
		return cmp
	}

	// Compare architectures.
	if a.Architecture.Is(&b.Architecture) || b.Architecture.Is(&a.Architecture) {
		return 0
	}
	if cmp := strings.Compare(a.Architecture.String(), b.Architecture.String()); cmp != 0 {
		return cmp
	}

	return 0
}
