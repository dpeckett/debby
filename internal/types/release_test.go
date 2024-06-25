// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package types_test

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"testing"

	"github.com/dpeckett/debby/internal/deb822"
	"github.com/dpeckett/debby/internal/keyring"
	"github.com/dpeckett/debby/internal/types"
	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/require"
)

func TestRelease(t *testing.T) {
	slog.SetDefault(slogt.New(t))

	f, err := os.Open("../../testdata/InRelease")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	ctx := context.Background()
	keyring, err := keyring.Load(ctx, http.DefaultClient, "../../testdata/archive-key-12.asc")
	require.NoError(t, err)

	decoder, err := deb822.NewDecoder(f, keyring)
	require.NoError(t, err)

	var release types.Release
	require.NoError(t, decoder.Decode(&release))

	require.Equal(t, "Debian", release.Origin)
}
