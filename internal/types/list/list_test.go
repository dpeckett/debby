// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package list_test

import (
	"math/big"
	"testing"

	"github.com/dpeckett/debby/internal/types/list"
	"github.com/stretchr/testify/require"
)

func TestCommaDelimited(t *testing.T) {
	t.Parallel()

	t.Run("Marshaler", func(t *testing.T) {
		t.Run("MarshalText", func(t *testing.T) {
			l := list.CommaDelimited[big.Int]{*big.NewInt(1), *big.NewInt(2), *big.NewInt(3)}

			text, err := l.MarshalText()
			require.NoError(t, err)

			require.Equal(t, "1, 2, 3", string(text))
		})

		t.Run("UnmarshalText", func(t *testing.T) {
			var l list.CommaDelimited[big.Int]
			err := l.UnmarshalText([]byte("1, 2, 3"))
			require.NoError(t, err)

			require.Equal(t, list.CommaDelimited[big.Int]{*big.NewInt(1), *big.NewInt(2), *big.NewInt(3)}, l)
		})
	})

	t.Run("String", func(t *testing.T) {
		t.Run("MarshalText", func(t *testing.T) {
			l := list.CommaDelimited[string]{"a", "b", "c"}

			text, err := l.MarshalText()
			require.NoError(t, err)

			require.Equal(t, "a, b, c", string(text))
		})

		t.Run("UnmarshalText", func(t *testing.T) {
			var l list.CommaDelimited[string]
			err := l.UnmarshalText([]byte("a, b, c"))
			require.NoError(t, err)

			require.Equal(t, list.CommaDelimited[string]{"a", "b", "c"}, l)
		})
	})

	t.Run("Int", func(t *testing.T) {
		t.Run("MarshalText", func(t *testing.T) {
			l := list.CommaDelimited[int]{1, 2, 3}

			text, err := l.MarshalText()
			require.NoError(t, err)

			require.Equal(t, "1, 2, 3", string(text))
		})

		t.Run("UnmarshalText", func(t *testing.T) {
			var l list.CommaDelimited[int]
			err := l.UnmarshalText([]byte("1, 2, 3"))
			require.NoError(t, err)

			require.Equal(t, list.CommaDelimited[int]{1, 2, 3}, l)
		})
	})
}

func TestNewLineDelimited(t *testing.T) {
	t.Parallel()

	t.Run("Marshaler", func(t *testing.T) {
		t.Run("MarshalText", func(t *testing.T) {
			l := list.NewLineDelimited[big.Int]{*big.NewInt(1), *big.NewInt(2), *big.NewInt(3)}

			text, err := l.MarshalText()
			require.NoError(t, err)

			require.Equal(t, " 1\n 2\n 3", string(text))
		})

		t.Run("UnmarshalText", func(t *testing.T) {
			var l list.NewLineDelimited[big.Int]
			err := l.UnmarshalText([]byte("1\n2\n3"))
			require.NoError(t, err)

			require.Equal(t, list.NewLineDelimited[big.Int]{*big.NewInt(1), *big.NewInt(2), *big.NewInt(3)}, l)
		})
	})

	t.Run("String", func(t *testing.T) {
		t.Run("MarshalText", func(t *testing.T) {
			l := list.NewLineDelimited[string]{"a", "b", "c"}

			text, err := l.MarshalText()
			require.NoError(t, err)

			require.Equal(t, "a\nb\nc", string(text))
		})

		t.Run("UnmarshalText", func(t *testing.T) {
			var l list.NewLineDelimited[string]
			err := l.UnmarshalText([]byte("a\nb\nc"))
			require.NoError(t, err)

			require.Equal(t, list.NewLineDelimited[string]{"a", "b", "c"}, l)
		})
	})

	t.Run("Int", func(t *testing.T) {
		t.Run("MarshalText", func(t *testing.T) {
			l := list.NewLineDelimited[int]{1, 2, 3}

			text, err := l.MarshalText()
			require.NoError(t, err)

			require.Equal(t, " 1\n 2\n 3", string(text))
		})

		t.Run("UnmarshalText", func(t *testing.T) {
			var l list.NewLineDelimited[int]
			err := l.UnmarshalText([]byte("1\n2\n3"))
			require.NoError(t, err)

			require.Equal(t, list.NewLineDelimited[int]{1, 2, 3}, l)
		})
	})
}

func TestSpaceDelimited(t *testing.T) {
	t.Parallel()

	t.Run("Marshaler", func(t *testing.T) {
		t.Run("MarshalText", func(t *testing.T) {
			l := list.SpaceDelimited[big.Int]{*big.NewInt(1), *big.NewInt(2), *big.NewInt(3)}

			text, err := l.MarshalText()
			require.NoError(t, err)

			require.Equal(t, "1 2 3", string(text))
		})

		t.Run("UnmarshalText", func(t *testing.T) {
			var l list.SpaceDelimited[big.Int]

			err := l.UnmarshalText([]byte("1 2 3"))
			require.NoError(t, err)

			require.Equal(t, list.SpaceDelimited[big.Int]{*big.NewInt(1), *big.NewInt(2), *big.NewInt(3)}, l)
		})
	})

	t.Run("String", func(t *testing.T) {
		t.Run("MarshalText", func(t *testing.T) {
			l := list.SpaceDelimited[string]{"a", "b", "c"}

			text, err := l.MarshalText()
			require.NoError(t, err)

			require.Equal(t, "a b c", string(text))
		})

		t.Run("UnmarshalText", func(t *testing.T) {
			var l list.SpaceDelimited[string]

			err := l.UnmarshalText([]byte("a b c"))
			require.NoError(t, err)

			require.Equal(t, list.SpaceDelimited[string]{"a", "b", "c"}, l)
		})
	})

	t.Run("Int", func(t *testing.T) {
		t.Run("MarshalText", func(t *testing.T) {
			l := list.SpaceDelimited[int]{1, 2, 3}

			text, err := l.MarshalText()
			require.NoError(t, err)

			require.Equal(t, "1 2 3", string(text))
		})

		t.Run("UnmarshalText", func(t *testing.T) {
			var l list.SpaceDelimited[int]

			err := l.UnmarshalText([]byte("1 2 3"))
			require.NoError(t, err)

			require.Equal(t, list.SpaceDelimited[int]{1, 2, 3}, l)
		})
	})
}
