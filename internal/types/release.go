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
	"encoding/hex"
	"strings"

	"github.com/dpeckett/debby/internal/deb822"
	"github.com/dpeckett/debby/internal/types/filehash"
	"github.com/dpeckett/debby/internal/types/list"
	"github.com/dpeckett/debby/internal/types/time"
)

// Release represents a Debian release.
type Release struct {
	// Origin is the origin of the release.
	Origin string
	// Label is the label of the release.
	Label string
	// Suite is the suite of the release.
	Suite string
	// Version is the version of the release.
	Version string
	// Codename is the codename of the release.
	Codename string
	// Changelogs is the URL to the changelogs for the release.
	Changelogs string
	// Date is the date the release was published.
	Date time.Time
	// ValidUntil is the date the release is valid until.
	ValidUntil time.Time `json:"Valid-Until"`
	// Architectures lists the architectures supported by the release.
	Architectures list.SpaceDelimited[string]
	// Components lists the components available in the release.
	Components list.SpaceDelimited[string]
	// Description is a description of the release.
	Description string
	// SHA256 lists SHA-256 checksums for files in the release.
	SHA256 list.NewLineDelimited[filehash.FileHash]
}

func (r Release) String() string {
	var sb strings.Builder
	if err := deb822.Marshal(&sb, r); err != nil {
		panic(err)
	}
	return sb.String()
}

// SHA256Sums returns a map of SHA-256 checksums for files in the release.
func (r *Release) SHA256Sums() (map[string][]byte, error) {
	ret := make(map[string][]byte)
	for _, hash := range r.SHA256 {
		var err error
		ret[hash.Filename], err = hex.DecodeString(hash.Hash)
		if err != nil {
			return nil, err
		}
	}
	return ret, nil
}
