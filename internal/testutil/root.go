// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package testutil

import (
	"os"
	"path/filepath"
)

// Root finds the root directory of the module by looking for the go.mod file.
func Root() string {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	// Look for go.mod file by walking up the directory structure.
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}

		// Move to the parent directory.
		parentDir := filepath.Dir(dir)
		if parentDir == dir {
			panic("could not find root directory")
		}
		dir = parentDir
	}
}
