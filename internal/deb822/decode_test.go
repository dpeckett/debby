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

package deb822_test

import (
	"testing"

	"github.com/dpeckett/debby/internal/deb822"
	"github.com/dpeckett/debby/internal/types/arch"
	"github.com/dpeckett/debby/internal/types/boolean"
	"github.com/dpeckett/debby/internal/types/dependency"
	"github.com/dpeckett/debby/internal/types/list"
	"github.com/dpeckett/debby/internal/types/version"
	"github.com/stretchr/testify/require"
)

type Fnord struct {
	FooBar string `json:"Fnord-Foo-Bar"`
}

type TestStruct struct {
	Fnord           `json:",inline"`
	Value           string
	ValueTwo        string `json:"Value-Two"`
	ValueThree      list.SpaceDelimited[string]
	Depends         dependency.Dependency
	Version         version.Version
	Arch            arch.Arch
	Arches          list.SpaceDelimited[arch.Arch]
	ExtraSourceOnly boolean.Boolean `json:"Extra-Source-Only"`
}

func TestBasicUnmarshal(t *testing.T) {
	foo := TestStruct{}

	require.NoError(t, deb822.Unmarshal([]byte(`Value: foo
Foo-Bar: baz
`), &foo))
	require.Equal(t, "foo", foo.Value)
}

func TestBasicArrayUnmarshal(t *testing.T) {
	var foo []TestStruct
	require.NoError(t, deb822.Unmarshal([]byte(`Value: foo
Foo-Bar: baz

Value: Bar

Value: Baz
`), &foo))
	require.Len(t, foo, 3)
	require.Equal(t, "foo", foo[0].Value)
}

func TestTagUnmarshal(t *testing.T) {
	var foo TestStruct
	require.NoError(t, deb822.Unmarshal([]byte(`Value: foo
Value-Two: baz
`), &foo))
	require.Equal(t, "foo", foo.Value)
	require.Equal(t, "baz", foo.ValueTwo)
}

func TestDependsUnmarshal(t *testing.T) {
	foo := TestStruct{}
	require.NoError(t, deb822.Unmarshal([]byte(`Value: foo
Depends: foo, bar
`), &foo))
	require.Equal(t, "foo", foo.Value)
	require.Len(t, foo.Depends.Relations, 2)
	require.Equal(t, "foo", foo.Depends.Relations[0].Possibilities[0].Name)
	require.Equal(t, "bar", foo.Depends.Relations[1].Possibilities[0].Name)

	// Actually invalid below
	require.Error(t, deb822.Unmarshal([]byte(`foo (>= 1.0) (<= 1.0)`), &foo))
}

func TestVersionUnmarshal(t *testing.T) {
	var foo TestStruct
	require.NoError(t, deb822.Unmarshal([]byte(`Value: foo
Version: 1.0-1
`), &foo))
	require.Equal(t, "foo", foo.Value)
	require.Equal(t, "1", foo.Version.Revision)
}

func TestArchUnmarshal(t *testing.T) {
	foo := TestStruct{}
	require.NoError(t, deb822.Unmarshal([]byte(`Value: foo
Arch: amd64
`), &foo))
	require.Equal(t, "foo", foo.Value)
	require.Equal(t, "amd64", foo.Arch.CPU)

	foo = TestStruct{}
	require.NoError(t, deb822.Unmarshal([]byte(`Value: foo
Arches: amd64 sparc any
`), &foo))
	require.Equal(t, "foo", foo.Value)
	require.Len(t, foo.Arches, 3)
	require.Equal(t, "amd64", foo.Arches[0].CPU)
	require.Equal(t, "sparc", foo.Arches[1].CPU)
	require.Equal(t, "any", foo.Arches[2].CPU)
}

func TestNestedUnmarshal(t *testing.T) {
	var foo TestStruct
	require.NoError(t, deb822.Unmarshal([]byte(`Value: foo
Fnord-Foo-Bar: Thing
`), &foo))
	require.Equal(t, "foo", foo.Value)
	require.Equal(t, "Thing", foo.Fnord.FooBar)
}

func TestListUnmarshal(t *testing.T) {
	foo := TestStruct{}
	require.NoError(t, deb822.Unmarshal([]byte(`Value: foo
ValueThree: foo bar baz
`), &foo))
	require.Equal(t, "foo", foo.Value)
	require.Len(t, foo.ValueThree, 3)
	require.Equal(t, "foo", foo.ValueThree[0])
}

func TestBoolUnmarshal(t *testing.T) {
	var foo TestStruct
	require.NoError(t, deb822.Unmarshal([]byte(`Value: foo
`), &foo))
	require.False(t, bool(foo.ExtraSourceOnly))

	require.NoError(t, deb822.Unmarshal([]byte(`Value: foo
Extra-Source-Only: no
`), &foo))
	require.False(t, bool(foo.ExtraSourceOnly))

	require.NoError(t, deb822.Unmarshal([]byte(`Value: foo
Extra-Source-Only: yes
`), &foo))
	require.True(t, bool(foo.ExtraSourceOnly))
}
