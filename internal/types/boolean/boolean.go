// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package boolean

import (
	"fmt"
	"strconv"
)

type Boolean bool

func (b Boolean) MarshalText() ([]byte, error) {
	if b {
		return []byte("yes"), nil
	}
	return []byte("no"), nil
}

func (b *Boolean) UnmarshalText(text []byte) error {
	switch string(text) {
	case "yes":
		*b = true
	case "no":
		*b = false
	default:
		val, err := strconv.ParseBool(string(text))
		if err != nil {
			return fmt.Errorf("failed to parse boolean: %w", err)
		}
		*b = Boolean(val)
	}
	return nil
}
