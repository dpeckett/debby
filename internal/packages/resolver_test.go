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
	"compress/gzip"
	"fmt"
	"os"
	"testing"

	"github.com/dpeckett/debby/internal/deb822"
	"github.com/dpeckett/debby/internal/packages"
	"github.com/dpeckett/debby/internal/types"
	"github.com/neilotoole/slogt"
	"github.com/stretchr/testify/require"
)

func TestResolver(t *testing.T) {
	logger := slogt.New(t)

	f, err := os.Open("../../testdata/Packages.gz")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	gzReader, err := gzip.NewReader(f)
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, gzReader.Close())
	})

	decoder, err := deb822.NewDecoder(gzReader, nil)
	require.NoError(t, err)

	var packageList []types.Package
	require.NoError(t, decoder.Decode(&packageList))

	packageDB := packages.NewPackageDB()
	packageDB.AddAll(packageList)

	res := packages.NewResolver(logger, packageDB)

	selectedDB, err := res.Resolve([]string{"bash=5.2.15-2+b2"})
	require.NoError(t, err)

	var selectedPackageNameVersions []string
	selectedDB.ForEach(func(pkg types.Package) error {
		if !pkg.IsVirtual {
			selectedPackageNameVersions = append(selectedPackageNameVersions,
				fmt.Sprintf("%s=%s", pkg.Package, pkg.Version))
		}

		return nil
	})

	expectedPackageNameVersions := []string{
		"base-files=12.4+deb12u5",
		"bash=5.2.15-2+b2",
		"debianutils=5.7-0.5~deb12u1",
		"gcc-12-base=12.2.0-14",
		"libc6=2.36-9+deb12u4",
		"libgcc-s1=12.2.0-14",
		"libtinfo6=6.4-4",
		"mawk=1.3.4.20200120-3.1",
	}

	require.ElementsMatch(t, expectedPackageNameVersions, selectedPackageNameVersions)
}
