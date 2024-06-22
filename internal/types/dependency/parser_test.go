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

package dependency_test

import (
	"testing"

	"github.com/dpeckett/debby/internal/types/dependency"
	"github.com/stretchr/testify/require"
)

func TestSingleParse(t *testing.T) {
	dep, err := dependency.Parse("foo")
	require.NoError(t, err)

	if dep.Relations[0].Possibilities[0].Name != "foo" {
		t.Fail()
	}
}

func TestMultiarchParse(t *testing.T) {
	dep, err := dependency.Parse("foo:amd64")
	require.NoError(t, err)

	require.Equal(t, "foo", dep.Relations[0].Possibilities[0].Name)
	require.Equal(t, "amd64", dep.Relations[0].Possibilities[0].Arch.CPU)

	dep, err = dependency.Parse("foo:amd64 [amd64 sparc]")
	require.NoError(t, err)

	require.Equal(t, "foo", dep.Relations[0].Possibilities[0].Name)
	require.Equal(t, "amd64", dep.Relations[0].Possibilities[0].Arch.CPU)

	require.Equal(t, "amd64", dep.Relations[0].Possibilities[0].Architectures.Architectures[0].CPU)
	require.Equal(t, "sparc", dep.Relations[0].Possibilities[0].Architectures.Architectures[1].CPU)
}

func TestTwoRelations(t *testing.T) {
	dep, err := dependency.Parse("foo, bar")
	require.NoError(t, err)

	require.Len(t, dep.Relations, 2)
}

func TestTwoPossibilities(t *testing.T) {
	dep, err := dependency.Parse("foo, bar | baz")
	require.NoError(t, err)

	require.Len(t, dep.Relations, 2)

	possi := dep.Relations[1].Possibilities

	require.Len(t, possi, 2)

	require.Equal(t, "bar", possi[0].Name)
	require.Equal(t, "baz", possi[1].Name)
}

func TestVersioning(t *testing.T) {
	dep, err := dependency.Parse("foo (>= 1.0)")
	require.NoError(t, err)

	require.Len(t, dep.Relations, 1)

	possi := dep.Relations[0].Possibilities[0]
	version := possi.Version

	require.Equal(t, ">=", version.Operator)
	require.Equal(t, "1.0", version.Version.String())
}

func TestVersioningSkippedSpace(t *testing.T) {
	dep, err := dependency.Parse("foo(>= 1.0)")
	require.NoError(t, err)

	require.Len(t, dep.Relations, 1)

	possi := dep.Relations[0].Possibilities[0]
	version := possi.Version

	require.Equal(t, ">=", version.Operator)
	require.Equal(t, "1.0", version.Version.String())
}

func TestSingleArch(t *testing.T) {
	dep, err := dependency.Parse("foo [arch]")
	require.NoError(t, err)

	require.Len(t, dep.Relations, 1)

	possi := dep.Relations[0].Possibilities[0]
	arches := possi.Architectures.Architectures

	require.Len(t, arches, 1)
	require.Equal(t, "arch", arches[0].CPU)
}

func TestSingleNotArch(t *testing.T) {
	dep, err := dependency.Parse("foo [!arch]")
	require.NoError(t, err)

	require.Len(t, dep.Relations, 1)

	possi := dep.Relations[0].Possibilities[0]
	arches := possi.Architectures.Architectures

	require.Len(t, arches, 1)
	require.Equal(t, "arch", arches[0].CPU)
	require.True(t, possi.Architectures.Not)
}

func TestDoubleInvalidNotArch(t *testing.T) {
	_, err := dependency.Parse("foo [arch !foo]")
	require.Error(t, err)

	_, err = dependency.Parse("foo [!arch foo]")
	require.Error(t, err)

	_, err = dependency.Parse("foo [arch!foo]")
	require.Error(t, err)

	_, err = dependency.Parse("foo [arch!]")
	require.Error(t, err)
}

func TestDoubleArch(t *testing.T) {
	for depStr, not := range map[string]bool{
		"foo [arch arch2]":   false,
		"foo [!arch !arch2]": true,
	} {
		dep, err := dependency.Parse(depStr)
		require.NoError(t, err)

		require.Len(t, dep.Relations, 1)

		possi := dep.Relations[0].Possibilities[0]
		arches := possi.Architectures.Architectures

		require.Equal(t, not, possi.Architectures.Not)

		require.Len(t, arches, 2)
		require.Equal(t, "arch", arches[0].CPU)
		require.Equal(t, "arch2", arches[1].CPU)
	}
}

func TestVersioningOperators(t *testing.T) {
	opers := map[string]string{
		">=": "foo (>= 1.0)",
		"<=": "foo (<= 1.0)",
		">>": "foo (>> 1.0)",
		"<<": "foo (<< 1.0)",
		"=":  "foo (= 1.0)",
	}

	for operator, vstring := range opers {
		dep, err := dependency.Parse(vstring)
		require.NoError(t, err)

		require.Len(t, dep.Relations, 1)

		possi := dep.Relations[0].Possibilities[0]
		version := possi.Version
		require.Equal(t, operator, version.Operator)
		require.Equal(t, "1.0", version.Version.String())
	}
}

func TestNoComma(t *testing.T) {
	_, err := dependency.Parse("foo bar")
	require.Error(t, err)
}

func TestTwoVersions(t *testing.T) {
	_, err := dependency.Parse("foo (>= 1.0) (<= 2.0)")
	require.Error(t, err)
}

func TestTwoArchitectures(t *testing.T) {
	_, err := dependency.Parse("foo [amd64] [sparc]")
	require.Error(t, err)
}

func TestTwoStages(t *testing.T) {
	dep, err := dependency.Parse("foo <stage1 !cross> <!stage1 cross>")
	require.NoError(t, err)

	possi := dep.Relations[0].Possibilities[0]

	require.Len(t, possi.StageSets, 2)

	// <stage1 !cross>
	require.Len(t, possi.StageSets[0].Stages, 2)
	require.False(t, possi.StageSets[0].Stages[0].Not)
	require.Equal(t, "stage1", possi.StageSets[0].Stages[0].Name)
	require.True(t, possi.StageSets[0].Stages[1].Not)
	require.Equal(t, "cross", possi.StageSets[0].Stages[1].Name)

	// <!stage1 cross>
	require.Len(t, possi.StageSets[1].Stages, 2)
	require.True(t, possi.StageSets[1].Stages[0].Not)
	require.Equal(t, "stage1", possi.StageSets[1].Stages[0].Name)
	require.False(t, possi.StageSets[1].Stages[1].Not)
	require.Equal(t, "cross", possi.StageSets[1].Stages[1].Name)
}

func TestBadVersion(t *testing.T) {
	vers := []string{
		"foo (>= 1.0",
		"foo (>= 1",
		"foo (>= ",
		"foo (>=",
		"foo (>",
		"foo (",
	}

	for _, ver := range vers {
		_, err := dependency.Parse(ver)
		require.Error(t, err)
	}
}

func TestBadArch(t *testing.T) {
	vers := []string{
		"foo [amd64",
		"foo [amd6",
		"foo [amd",
		"foo [am",
		"foo [a",
		"foo [",
	}

	for _, ver := range vers {
		_, err := dependency.Parse(ver)
		require.Error(t, err)
	}
}

func TestBadStages(t *testing.T) {
	vers := []string{
		"foo <stage1> <!cross",
		"foo <stage1> <!cros",
		"foo <stage1> <!cro",
		"foo <stage1> <!cr",
		"foo <stage1> <!c",
		"foo <stage1> <!",
		"foo <stage1> <",
		"foo <stage1",
		"foo <stag",
		"foo <sta",
		"foo <st",
		"foo <s",
		"foo <",
	}

	for _, ver := range vers {
		_, err := dependency.Parse(ver)
		require.Error(t, err)
	}
}

func TestSingleSubstvar(t *testing.T) {
	dep, err := dependency.Parse("${foo:Depends}, bar, baz")
	require.NoError(t, err)

	require.Len(t, dep.Relations, 3)

	require.Equal(t, "foo:Depends", dep.Relations[0].Possibilities[0].Name)
	require.Equal(t, "bar", dep.Relations[1].Possibilities[0].Name)
	require.Equal(t, "baz", dep.Relations[2].Possibilities[0].Name)

	require.True(t, dep.Relations[0].Possibilities[0].Substvar)
	require.False(t, dep.Relations[1].Possibilities[0].Substvar)
	require.False(t, dep.Relations[2].Possibilities[0].Substvar)
}

func TestInsaneRoundTrip(t *testing.T) {
	dep, err := dependency.Parse("foo:armhf <stage1 !cross> [amd64 i386] (>= 12:3.4~5.6-7.8~9.0) <!stage1 cross>")
	require.NoError(t, err)

	require.Equal(t, "foo:armhf [amd64 i386] (>= 12:3.4~5.6-7.8~9.0) <stage1 !cross> <!stage1 cross>", dep.String())

	rtDep, err := dependency.Parse(dep.String())
	require.NoError(t, err)
	require.Equal(t, dep.String(), rtDep.String())

	dep.Relations[0].Possibilities[0].Architectures.Not = true
	require.Equal(t, "foo:armhf [!amd64 !i386] (>= 12:3.4~5.6-7.8~9.0) <stage1 !cross> <!stage1 cross>", dep.String())

	rtDep, err = dependency.Parse(dep.String())
	require.NoError(t, err)

	require.Equal(t, dep.String(), rtDep.String())
}
