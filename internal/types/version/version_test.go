// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * Portions of this file are based on code originally from: github.com/paultag/go-debian
 *
 * Copyright (c) 2012 Michael Stapelberg and contributors
 * All rights reserved.
 *
 * Redistribution and use in source and binary forms, with or without
 * modification, are permitted provided that the following conditions are met:
 *
 *     * Redistributions of source code must retain the above copyright
 *       notice, this list of conditions and the following disclaimer.
 *
 *     * Redistributions in binary form must reproduce the above copyright
 *       notice, this list of conditions and the following disclaimer in the
 *       documentation and/or other materials provided with the distribution.
 *
 *     * Neither the name of Michael Stapelberg nor the
 *       names of contributors may be used to endorse or promote products
 *       derived from this software without specific prior written permission.
 *
 * THIS SOFTWARE IS PROVIDED BY Michael Stapelberg ''AS IS'' AND ANY
 * EXPRESS OR IMPLIED WARRANTIES, INCLUDING, BUT NOT LIMITED TO, THE IMPLIED
 * WARRANTIES OF MERCHANTABILITY AND FITNESS FOR A PARTICULAR PURPOSE ARE
 * DISCLAIMED. IN NO EVENT SHALL Michael Stapelberg BE LIABLE FOR ANY
 * DIRECT, INDIRECT, INCIDENTAL, SPECIAL, EXEMPLARY, OR CONSEQUENTIAL DAMAGES
 * (INCLUDING, BUT NOT LIMITED TO, PROCUREMENT OF SUBSTITUTE GOODS OR SERVICES;
 * LOSS OF USE, DATA, OR PROFITS; OR BUSINESS INTERRUPTION) HOWEVER CAUSED AND
 * ON ANY THEORY OF LIABILITY, WHETHER IN CONTRACT, STRICT LIABILITY, OR TORT
 * (INCLUDING NEGLIGENCE OR OTHERWISE) ARISING IN ANY WAY OUT OF THE USE OF THIS
 * SOFTWARE, EVEN IF ADVISED OF THE POSSIBILITY OF SUCH DAMAGE.
 */

package version_test

import (
	"testing"

	"github.com/dpeckett/debby/internal/types/version"
	"github.com/stretchr/testify/require"
)

// Abbreviation for creating a new Version object.
func v(epoch uint, versionNumber string, revision string) version.Version {
	return version.Version{Epoch: epoch, Version: versionNumber, Revision: revision}
}

func TestVersion(t *testing.T) {
	t.Run("Parse", func(t *testing.T) {
		t.Run("Zero", func(t *testing.T) {
			b := v(0, "0", "")

			a, err := version.Parse("0")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))

			a, err = version.Parse("0:0")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))

			a, err = version.Parse("0:0-")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))

			b = v(0, "0", "0")
			a, err = version.Parse("0:0-0")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))

			b = v(0, "0.0", "0.0")
			a, err = version.Parse("0:0.0-0.0")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))
		})

		t.Run("Epoched", func(t *testing.T) {
			b := v(1, "0", "")

			a, err := version.Parse("1:0")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))

			b = v(5, "1", "")
			a, err = version.Parse("5:1")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))
		})

		t.Run("MultipleHyphens", func(t *testing.T) {
			b := v(0, "0-0", "0")

			a, err := version.Parse("0:0-0-0")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))

			b = v(0, "0-0-0", "0")
			a, err = version.Parse("0:0-0-0-0")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))
		})

		t.Run("MultipleColons", func(t *testing.T) {
			b := v(0, "0:0", "0")

			a, err := version.Parse("0:0:0-0")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))

			b = v(0, "0:0:0", "0")
			a, err = version.Parse("0:0:0:0-0")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))
		})

		t.Run("MultipleHyphensAndColons", func(t *testing.T) {
			b := v(0, "0:0-0", "0")

			a, err := version.Parse("0:0:0-0-0")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))

			b = v(0, "0-0:0", "0")
			a, err = version.Parse("0:0-0:0-0")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))
		})

		t.Run("ValidUpstreamVersionCharacters", func(t *testing.T) {
			b := v(0, "09azAZ.-+~:", "0")

			a, err := version.Parse("0:09azAZ.-+~:-0")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))
		})

		t.Run("ValidRevisionCharacters", func(t *testing.T) {
			b := v(0, "0", "09azAZ.+~")

			a, err := version.Parse("0:0-09azAZ.+~")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))
		})

		t.Run("LeadingTrailingSpaces", func(t *testing.T) {
			b := v(0, "0", "1")

			a, err := version.Parse("    0:0-1")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))

			a, err = version.Parse("0:0-1     ")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))

			a, err = version.Parse("      0:0-1     ")
			require.NoError(t, err)

			require.Zero(t, a.Compare(b))
		})

		t.Run("EmptyVersion", func(t *testing.T) {
			_, err := version.Parse("")
			require.Error(t, err)

			_, err = version.Parse("  ")
			require.Error(t, err)
		})

		t.Run("EmptyUpstreamVersionAfterEpoch", func(t *testing.T) {
			_, err := version.Parse("0:")
			require.Error(t, err)
		})

		t.Run("VersionWithEmbeddedSpaces", func(t *testing.T) {
			_, err := version.Parse("0:0 0-1")
			require.Error(t, err)
		})

		t.Run("VersionWithNegativeEpoch", func(t *testing.T) {
			_, err := version.Parse("-1:0-1")
			require.Error(t, err)
		})

		t.Run("VersionWithHugeEpoch", func(t *testing.T) {
			_, err := version.Parse("999999999999999999999999:0-1")
			require.Error(t, err)
		})

		t.Run("InvalidCharactersInEpoch", func(t *testing.T) {
			_, err := version.Parse("a:0-0")
			require.Error(t, err)

			_, err = version.Parse("A:0-0")
			require.Error(t, err)
		})

		t.Run("UpstreamVersionNotStartingWithADigit", func(t *testing.T) {
			_, err := version.Parse("0:abc3-0")
			require.Error(t, err)
		})

		t.Run("InvalidCharactersInUpstreamVersion", func(t *testing.T) {
			chars := "!#@$%&/|\\<>()[]{};,_=*^'"
			for i := 0; i < len(chars); i++ {
				verstr := "0:0" + chars[i:i+1] + "-0"
				_, err := version.Parse(verstr)
				require.Error(t, err)
			}
		})

		t.Run("InvalidCharactersInRevision", func(t *testing.T) {
			_, err := version.Parse("0:0-0:0")
			require.Error(t, err)

			chars := "!#@$%&/|\\<>()[]{}:;,_=*^'"
			for i := 0; i < len(chars); i++ {
				verstr := "0:0-" + chars[i:i+1]
				_, err := version.Parse(verstr)
				require.Error(t, err)
			}
		})

		t.Run("Compare", func(t *testing.T) {
			t.Run("Equal", func(t *testing.T) {
				a, b := v(0, "0", "0"), v(0, "0", "0")
				require.Zero(t, a.Compare(b))

				a, b = v(0, "0", "00"), v(0, "00", "0")
				require.Zero(t, a.Compare(b))

				a, b = v(1, "2", "3"), v(1, "2", "3")
				require.Zero(t, a.Compare(b))
			})

			t.Run("Epoch", func(t *testing.T) {
				require.NotZero(t, (version.Version{Epoch: 1}).Compare(version.Version{Epoch: 2}))

				a, b := v(0, "1", "1"), v(0, "2", "1")
				require.NotZero(t, a.Compare(b))

				a, b = v(0, "1", "1"), v(0, "1", "2")
				require.NotZero(t, a.Compare(b))

				a, b = v(0, "0", "0"), v(1, "0", "0")
				require.Less(t, a.Compare(b), 0)
				require.Greater(t, b.Compare(a), 0)
			})

			t.Run("Version", func(t *testing.T) {
				a, b := v(0, "a", "0"), v(0, "b", "0")
				require.Less(t, a.Compare(b), 0)
				require.Greater(t, b.Compare(a), 0)
			})

			t.Run("Revision", func(t *testing.T) {
				a, b := v(0, "0", "a"), v(0, "0", "b")
				require.Less(t, a.Compare(b), 0)
				require.Greater(t, b.Compare(a), 0)
			})

			t.Run("CodeSearch", func(t *testing.T) {
				a, b := v(0, "1.8.6", "2"), v(0, "1.8.6", "2.1")
				require.Less(t, a.Compare(b), 0)
			})
		})

		t.Run("TestString", func(t *testing.T) {
			require.Equal(t, "1.0-1", version.Version{
				Version:  "1.0",
				Revision: "1",
			}.String(), "String() returned malformed Version")

			require.Equal(t, "1:1.0-1", version.Version{
				Epoch:    1,
				Version:  "1.0",
				Revision: "1",
			}.String(), "String() returned malformed Version with Epoch")

			require.Equal(t, "1.0-1", version.Version{
				Epoch:    1,
				Version:  "1.0",
				Revision: "1",
			}.StringWithoutEpoch(), "StringWithoutEpoch() returned malformed Version with Epoch")
		})
	})
}
