// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package util

import (
	"crypto/sha256"
	"hash"
	"io"
)

// HashReader is a wrapper around an io.Reader that calculates the SHA-256 hash of the read data.
type HashReader struct {
	reader io.Reader
	hasher hash.Hash
}

// NewHashReader creates a new HashReader.
func NewHashReader(r io.Reader) *HashReader {
	hasher := sha256.New()
	return &HashReader{
		reader: io.TeeReader(r, hasher),
		hasher: hasher,
	}
}

// Read reads from the underlying reader and updates the hash.
func (hr *HashReader) Read(p []byte) (int, error) {
	return hr.reader.Read(p)
}

// Sum returns the SHA-256 checksum of the read data.
func (hr *HashReader) Sum() []byte {
	return hr.hasher.Sum(nil)
}
