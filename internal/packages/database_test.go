// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package packages_test

import (
	"testing"

	"github.com/dpeckett/debby/internal/packages"
	"github.com/dpeckett/debby/internal/types"
	"github.com/dpeckett/debby/internal/types/dependency"
	"github.com/dpeckett/debby/internal/types/version"
	"github.com/stretchr/testify/require"
)

func TestPackageDB(t *testing.T) {
	db := packages.NewPackageDB()

	db.AddAll([]types.Package{
		{
			Package: "foo",
			Version: version.MustParse("1.0"),
		},
		{
			Package: "foo",
			Version: version.MustParse("1.1"),
		},
		{
			Package: "bar",
			Version: version.MustParse("2.0"),
		},
	})

	require.Equal(t, 3, db.Len())

	t.Run("Get", func(t *testing.T) {
		t.Run("All", func(t *testing.T) {
			packages := db.Get("foo")

			require.Len(t, packages, 2)
		})

		t.Run("Strictly Earlier", func(t *testing.T) {
			packages := db.StrictlyEarlier("foo", version.MustParse("1.1"))

			require.Len(t, packages, 1)
			require.Equal(t, "foo", packages[0].Package)
			require.Equal(t, version.MustParse("1.0"), packages[0].Version)
		})

		t.Run("Earlier or Equal", func(t *testing.T) {
			packages := db.EarlierOrEqual("foo", version.MustParse("1.1"))

			require.Len(t, packages, 2)
		})

		t.Run("Exact Version", func(t *testing.T) {
			pkg, exists := db.ExactlyEqual("foo", version.MustParse("1.0"))

			require.True(t, exists)
			require.Equal(t, "foo", pkg.Package)
			require.Equal(t, version.MustParse("1.0"), pkg.Version)
		})

		t.Run("Exact Version (Missing)", func(t *testing.T) {
			_, exists := db.ExactlyEqual("foo", version.MustParse("1.2"))

			require.False(t, exists)
		})

		t.Run("Later or Equal", func(t *testing.T) {
			packages := db.LaterOrEqual("foo", version.MustParse("1.0"))

			require.Len(t, packages, 2)
			require.Equal(t, "foo", packages[0].Package)
			require.Equal(t, version.MustParse("1.0"), packages[0].Version)
			require.Equal(t, version.MustParse("1.1"), packages[1].Version)
		})

		t.Run("Strictly Later", func(t *testing.T) {
			packages := db.StrictlyLater("foo", version.MustParse("1.0"))

			require.Len(t, packages, 1)
			require.Equal(t, "foo", packages[0].Package)
			require.Equal(t, version.MustParse("1.1"), packages[0].Version)
		})
	})

	t.Run("Add and Remove", func(t *testing.T) {
		pkg := types.Package{
			Package: "baz",
			Version: version.MustParse("3.0"),
		}

		db.Add(pkg)

		require.Equal(t, 4, db.Len())

		db.Remove(pkg)

		require.Equal(t, 3, db.Len())
	})

	t.Run("Virtual Packages", func(t *testing.T) {
		pkg := types.Package{
			Package: "baz",
			Version: version.MustParse("3.0"),
			Provides: dependency.Dependency{
				Relations: []dependency.Relation{
					{
						Possibilities: []dependency.Possibility{{Name: "bazz"}},
					},
				},
			},
		}

		db.Add(pkg)

		packages := db.Get("bazz")

		require.Len(t, packages, 1)
		require.Equal(t, "bazz", packages[0].Package)
		require.True(t, packages[0].IsVirtual)
		require.Equal(t, "baz", packages[0].Providers[0].Package)
		require.Equal(t, version.MustParse("3.0"), packages[0].Providers[0].Version)
	})
}
