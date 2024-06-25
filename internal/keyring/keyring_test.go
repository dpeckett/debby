// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package keyring_test

import (
	"context"
	"log/slog"
	"net/http"
	"testing"

	"github.com/dpeckett/debby/internal/keyring"
	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/require"
)

func TestKeyringRead(t *testing.T) {
	ctx := context.Background()
	slog.SetDefault(slogt.New(t))

	t.Run("Web", func(t *testing.T) {
		keyring, err := keyring.Load(ctx, http.DefaultClient, "https://ftp-master.debian.org/keys/archive-key-12.asc")
		require.NoError(t, err)

		require.NotEmpty(t, keyring)
	})

	t.Run("File", func(t *testing.T) {
		keyring, err := keyring.Load(ctx, http.DefaultClient, "../../testdata/archive-key-12.asc")
		require.NoError(t, err)

		require.NotEmpty(t, keyring)
	})
}
