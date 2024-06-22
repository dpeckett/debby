// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package util_test

import (
	"bytes"
	"encoding/hex"
	"io"
	"testing"

	"github.com/dpeckett/debby/internal/util"
	"github.com/stretchr/testify/require"
)

func TestHashReader(t *testing.T) {
	data := []byte("The quick brown fox jumps over the lazy dog")

	expectedHash, err := hex.DecodeString("d7a8fbb307d7809469ca9abcb0082e4f8d5651e46d3cdb762d02d0bf37c9e592")
	require.NoError(t, err)

	// Create a HashReader
	reader := bytes.NewReader(data)
	hashReader := util.NewHashReader(reader)

	// Read the data
	readData, err := io.ReadAll(hashReader)
	require.NoError(t, err)
	require.Equal(t, data, readData)

	// Verify the checksum
	computedHash := hashReader.Sum()
	require.Equal(t, expectedHash, computedHash)
}
