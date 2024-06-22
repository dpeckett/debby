// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package filehash

import (
	"fmt"
)

// FileHash is an entry found in a Debian Release file
type FileHash struct {
	Hash     string
	Size     int64
	Filename string
}

func (h FileHash) String() string {
	return fmt.Sprintf("%s %d %s", h.Hash, h.Size, h.Filename)
}

func (h FileHash) MarshalText() ([]byte, error) {
	return []byte(h.String()), nil
}

func (h *FileHash) UnmarshalText(text []byte) error {
	_, err := fmt.Sscanf(string(text), "%s %d %s", &h.Hash, &h.Size, &h.Filename)
	return err
}
