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
	"bufio"
	"bytes"
	"compress/gzip"
	"io"

	"github.com/ulikunitz/xz"
)

// Decompress returns a reader that decompresses the input stream if it is compressed.
func Decompress(r io.Reader) (io.ReadCloser, error) {
	bufioReader := bufio.NewReader(r)

	buf, err := bufioReader.Peek(8)
	if err != nil {
		return nil, err
	}

	switch {
	case bytes.HasPrefix(buf, []byte{0x1F, 0x8B}): // GZIP
		return gzip.NewReader(bufioReader)
	case bytes.HasPrefix(buf, []byte{0xFD, 0x37, 0x7A, 0x58, 0x5A, 0x00}): // XZ
		xzReader, err := xz.NewReader(bufioReader)
		if err != nil {
			return nil, err
		}

		return io.NopCloser(xzReader), nil
	default:
		return io.NopCloser(bufioReader), nil
	}
}
