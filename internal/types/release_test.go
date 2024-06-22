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
	"os"
	"testing"

	"github.com/dpeckett/debby/internal/control"
	"github.com/dpeckett/debby/internal/types"
	"github.com/stretchr/testify/require"
)

func TestRelease(t *testing.T) {
	f, err := os.Open("testdata/Release")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	decoder, err := control.NewDecoder(f, nil)
	require.NoError(t, err)

	var release types.Release
	require.NoError(t, decoder.Decode(&release))

	require.Equal(t, "Debian", release.Origin)
}
