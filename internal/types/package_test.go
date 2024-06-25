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
	"compress/gzip"
	"io"
	"os"
	"testing"

	"github.com/dpeckett/debby/internal/deb822"
	"github.com/dpeckett/debby/internal/types"
	"github.com/dpeckett/debby/internal/types/arch"
	"github.com/dpeckett/debby/internal/types/dependency"
	"github.com/dpeckett/debby/internal/types/version"
	"github.com/stretchr/testify/require"
)

func TestPackage(t *testing.T) {
	f, err := os.Open("../../testdata/Packages.gz")
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, f.Close())
	})

	gzReader, err := gzip.NewReader(f)
	require.NoError(t, err)

	decoder, err := deb822.NewDecoder(io.LimitReader(gzReader, 1000000), nil)
	require.NoError(t, err)

	var packageList []types.Package
	require.NoError(t, decoder.Decode(&packageList))

	require.Len(t, packageList, 1324)

	expected := types.Package{
		Package:       "0ad",
		Version:       version.MustParse("0.0.26-3"),
		InstalledSize: 28591,
		Maintainer:    "Debian Games Team <pkg-games-devel@lists.alioth.debian.org>",
		Architecture:  arch.MustParse("amd64"),
		Depends:       dependency.MustParse("0ad-data (>= 0.0.26), 0ad-data (<= 0.0.26-3), 0ad-data-common (>= 0.0.26), 0ad-data-common (<= 0.0.26-3), libboost-filesystem1.74.0 (>= 1.74.0), libc6 (>= 2.34), libcurl3-gnutls (>= 7.32.0), libenet7, libfmt9 (>= 9.1.0+ds1), libfreetype6 (>= 2.2.1), libgcc-s1 (>= 3.4), libgloox18 (>= 1.0.24), libicu72 (>= 72.1~rc-1~), libminiupnpc17 (>= 1.9.20140610), libopenal1 (>= 1.14), libpng16-16 (>= 1.6.2-1), libsdl2-2.0-0 (>= 2.0.12), libsodium23 (>= 1.0.14), libstdc++6 (>= 12), libvorbisfile3 (>= 1.1.2), libwxbase3.2-1 (>= 3.2.1+dfsg), libwxgtk-gl3.2-1 (>= 3.2.1+dfsg), libwxgtk3.2-1 (>= 3.2.1+dfsg-2), libx11-6, libxml2 (>= 2.9.0), zlib1g (>= 1:1.2.0)"),
		PreDepends:    dependency.MustParse("dpkg (>= 1.15.6~)"),
		Description:   "Real-time strategy game of ancient warfare",
		Homepage:      "https://play0ad.com/",
		Tag:           []string{"game::strategy", "interface::graphical", "interface::x11", "role::program", "uitoolkit::sdl", "uitoolkit::wxwidgets", "use::gameplaying", "x11::application"},
		Section:       "games",
		Priority:      "optional",
		Filename:      "pool/main/0/0ad/0ad_0.0.26-3_amd64.deb",
		Size:          7891488,
		SHA256:        "3a2118df47bf3f04285649f0455c2fc6fe2dc7f0b237073038aa00af41f0d5f2",
	}

	require.Equal(t, expected, packageList[0])
}
