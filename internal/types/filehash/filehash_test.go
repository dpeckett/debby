// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package filehash_test

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/dpeckett/debby/internal/deb822"
	"github.com/dpeckett/debby/internal/keyring"
	"github.com/dpeckett/debby/internal/types/filehash"
	"github.com/dpeckett/debby/internal/types/list"
	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/require"
)

func TestFileHash(t *testing.T) {
	slog.SetDefault(slogt.New(t))

	f, err := os.Open("../../../testdata/InRelease")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	ctx := context.Background()
	keyring, err := keyring.Load(ctx, http.DefaultClient, "../../../testdata/archive-key-12.asc")
	require.NoError(t, err)

	decoder, err := deb822.NewDecoder(f, keyring)
	require.NoError(t, err)

	type TestStruct struct {
		MD5Sum list.NewLineDelimited[filehash.FileHash]
		SHA256 list.NewLineDelimited[filehash.FileHash]
	}

	var foo TestStruct
	require.NoError(t, decoder.Decode(&foo))

	require.Len(t, foo.MD5Sum, 772)
	require.Len(t, foo.SHA256, 772)

	require.Equal(t, "0ed6d4c8891eb86358b94bb35d9e4da4", foo.MD5Sum[0].Hash)
	require.Equal(t, int64(1484322), foo.MD5Sum[0].Size)
	require.Equal(t, "contrib/Contents-all", foo.MD5Sum[0].Filename)
}

func TestFileHash_MarshalText(t *testing.T) {
	hashes := list.NewLineDelimited[filehash.FileHash]([]filehash.FileHash{{
		Hash:     "0ed6d4c8891eb86358b94bb35d9e4da4",
		Size:     1484322,
		Filename: "contrib/Contents-all",
	}, {
		Hash:     "d0a0325a97c42fd5f66a8c3e29bcea64",
		Size:     98581,
		Filename: "contrib/Contents-all.gz",
	}})

	text, err := hashes.MarshalText()
	require.NoError(t, err)

	expected := ` 0ed6d4c8891eb86358b94bb35d9e4da4 1484322 contrib/Contents-all
 d0a0325a97c42fd5f66a8c3e29bcea64 98581 contrib/Contents-all.gz`

	require.Equal(t, expected, string(text))
}
