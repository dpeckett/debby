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
	"bytes"

	"github.com/dpeckett/debby/internal/control"
	"github.com/dpeckett/debby/internal/types/arch"
	"github.com/dpeckett/debby/internal/types/dependency"
	"github.com/dpeckett/debby/internal/types/version"
)

// Package represents a Debian package.
type Package struct {
	// Package is the name of the package.
	Package string
	// Source is the source package name.
	Source string `control:",omitempty"`
	// Version is the version of the package.
	Version version.Version
	// InstalledSize is the installed size of the package, in kilobytes.
	InstalledSize int `control:"Installed-Size,omitempty"`
	// Maintainer is the person or organization responsible for the package.
	Maintainer string `control:",omitempty"`
	// Architecture is the architecture the package is built for.
	Architecture arch.Arch
	// Replaces lists packages that this package replaces.
	Replaces dependency.Dependency `control:",omitempty"`
	// Breaks lists packages that this package breaks.
	Breaks dependency.Dependency `control:",omitempty"`
	// Provides lists virtual packages that this package provides.
	Provides dependency.Dependency `control:",omitempty"`
	// Conflicts lists packages that conflict with this package.
	Conflicts dependency.Dependency `control:",omitempty"`
	// Enhances lists packages that this package enhances.
	Enhances dependency.Dependency `control:",omitempty"`
	// Depends lists packages that this package depends on.
	Depends dependency.Dependency `control:",omitempty"`
	// Recommends lists packages that are recommended to be installed with this package.
	Recommends dependency.Dependency `control:",omitempty"`
	// Suggests lists packages that are suggested to be installed with this package.
	Suggests dependency.Dependency `control:",omitempty"`
	// PreDepends lists packages that must be installed and configured before this package.
	PreDepends dependency.Dependency `control:"Pre-Depends,omitempty"`
	// Description provides a short description of the package.
	Description string
	// Homepage is the URL of the package's homepage.
	Homepage string `control:",omitempty"`
	// Tag lists tags associated with the package.
	Tag []string `control:",omitempty"`
	// Section categorizes the package within the archive.
	Section string `control:",omitempty"`
	// Priority defines the importance of the package.
	Priority string `control:",omitempty"`
	// Filename is the name of the package file.
	Filename string `control:",omitempty"`
	// Size is the size of the package file, in bytes.
	Size int `control:",omitempty"`
	// SHA256 is the SHA-256 checksum of the package file.
	SHA256 string `control:",omitempty"`
	// Additional fields that are not part of the control file.
	// The full URL to the package file.
	URL string `control:"-"`
}

// ControlFieldOrder returns the order of fields in the control file.
func (p Package) ControlFieldOrder() []string {
	return []string{
		"Package",
		"Source",
		"Version",
		"Installed-Size",
		"Maintainer",
		"Architecture",
		"Replaces",
		"Breaks",
		"Provides",
		"Conflicts",
		"Enhances",
		"Depends",
		"Recommends",
		"Suggests",
		"Pre-Depends",
		"Description",
		"Homepage",
		"Tag",
		"Section",
		"Priority",
		"Filename",
		"Size",
		"SHA256",
	}
}

func (p Package) String() string {
	text, err := p.MarshalText()
	if err != nil {
		panic(err)
	}
	return string(text)
}

func (p Package) MarshalText() ([]byte, error) {
	var buf bytes.Buffer
	if err := control.Marshal(&buf, p); err != nil {
		panic(err)
	}
	return buf.Bytes(), nil
}

func (p *Package) UnmarshalText(text []byte) error {
	return control.Unmarshal(text, p)
}

// ID returns a unique identifier for the package.
func (p Package) ID() string {
	return p.Package + "_" + p.Version.String() + "_" + p.Architecture.String()
}
