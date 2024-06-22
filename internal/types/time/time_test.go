// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package time_test

import (
	"testing"

	stdtime "time"

	"github.com/dpeckett/debby/internal/types/time"

	"github.com/stretchr/testify/require"
)

func TestTime(t *testing.T) {
	t.Run("MarshalText", func(t *testing.T) {
		tm := time.Time(stdtime.Date(2024, stdtime.February, 10, 11, 7, 25, 0, stdtime.UTC))

		text, err := tm.MarshalText()
		require.NoError(t, err)

		require.Equal(t, "Sat, 10 Feb 2024 11:07:25 UTC", string(text))
	})

	t.Run("UnmarshalText", func(t *testing.T) {
		text := "Sat, 10 Feb 2024 11:07:25 UTC"

		var tm time.Time
		require.NoError(t, tm.UnmarshalText([]byte(text)))

		require.Zero(t, stdtime.Date(2024, stdtime.February, 10, 11, 7, 25, 0, stdtime.UTC).Compare(stdtime.Time(tm)))
	})
}
