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
	"crypto/hmac"
	"crypto/sha256"
	"errors"
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

// Verify returns true if the calculated hash matches the expected hash.
func (hr *HashReader) Verify(expected []byte) error {
	if !hmac.Equal(hr.hasher.Sum(nil), expected) {
		return errors.New("hash mismatch")
	}

	return nil
}
