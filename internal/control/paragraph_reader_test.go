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
 * Copyright (c) Paul R. Tagliamonte <paultag@debian.org>, 2015
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package control_test

import (
	"os"
	"strings"
	"testing"

	"github.com/ProtonMail/go-crypto/openpgp"
	"github.com/dpeckett/debby/internal/control"
	"github.com/stretchr/testify/require"
)

func TestBasicParagraphReader(t *testing.T) {
	reader, err := control.NewParagraphReader(strings.NewReader(`Para: one

Para: two

Para: three
`), nil)
	require.NoError(t, err)

	blocks, err := reader.All()
	require.NoError(t, err)

	require.Len(t, blocks, 3)
}

func TestMultipleNewlines(t *testing.T) {
	reader, err := control.NewParagraphReader(strings.NewReader(`Para: one


Para: two

Para: three
 `), nil)
	require.NoError(t, err)

	blocks, err := reader.All()
	require.NoError(t, err)

	require.Len(t, blocks, 3)
}

func TestWhitespacePrefixedLines(t *testing.T) {
	reader, err := control.NewParagraphReader(strings.NewReader(`Key1: one
	 continuation
Key2: two
	 tabbed continuation
 `), nil)
	require.NoError(t, err)

	blocks, err := reader.All()
	require.NoError(t, err)

	require.Len(t, blocks, 1)
	require.Equal(t, "one\n continuation\n", blocks[0]["Key1"])
	require.Equal(t, "two\n tabbed continuation\n", blocks[0]["Key2"])
}

func TestCommentLines(t *testing.T) {
	reader, err := control.NewParagraphReader(strings.NewReader(`Key1: one
# comment
Key2: two
 `), nil)
	require.NoError(t, err)

	blocks, err := reader.All()
	require.NoError(t, err)

	require.Len(t, blocks, 1)
	require.Equal(t, "one", blocks[0]["Key1"])
	require.Equal(t, "two", blocks[0]["Key2"])
}

func TestTrailingTwoCharacterNewlines(t *testing.T) {
	reader, err := control.NewDecoder(strings.NewReader("Key1: one\r\nKey2: two\r\n\r\n"), nil)
	require.NoError(t, err)

	type TestStruct struct {
		Key1 string
		Key2 string
	}

	testStruct := TestStruct{}
	require.NoError(t, reader.Decode(&testStruct))

	require.Equal(t, "one", testStruct.Key1)
	require.Equal(t, "two", testStruct.Key2)
}

func TestOpenPGPParagraphReader(t *testing.T) {
	f, err := os.Open("testdata/0ad_0.0.26-3.dsc")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	pubKeyFile, err := os.Open("testdata/d53a815a3cb7659af882e3958eedcc1baa1f32ff.asc")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, pubKeyFile.Close())
	})

	entity, err := openpgp.ReadArmoredKeyRing(pubKeyFile)
	require.NoError(t, err)

	reader, err := control.NewParagraphReader(f, openpgp.EntityList{entity[0]})
	require.NoError(t, err)

	blocks, err := reader.All()
	require.NoError(t, err)

	require.Len(t, blocks, 1)
}

func TestEmptyKeyringOpenPGPParagraphReader(t *testing.T) {
	keyring := openpgp.EntityList{}

	f, err := os.Open("testdata/0ad_0.0.26-3.dsc")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	_, err = control.NewParagraphReader(f, keyring)
	require.Error(t, err)
}
