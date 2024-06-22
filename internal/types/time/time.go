// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package time

import (
	stdtime "time"
)

// Time is an RFC2822 formatted time.
type Time stdtime.Time

func (t Time) MarshalText() ([]byte, error) {
	return []byte(stdtime.Time(t).Format(stdtime.RFC1123)), nil
}

func (t *Time) UnmarshalText(text []byte) error {
	parsed, err := stdtime.Parse(stdtime.RFC1123, string(text))
	if err != nil {
		return err
	}

	*t = Time(parsed)

	return nil
}
