// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package list

import (
	"encoding"
	"fmt"
	"strings"
)

// CommaDelimited is a list of T entries separated by commas.
type CommaDelimited[T any] []T

func (l CommaDelimited[T]) MarshalText() ([]byte, error) {
	var sb strings.Builder
	for i, entry := range l {
		if i > 0 {
			sb.WriteString(", ")
		}

		switch v := any(entry).(type) {
		case string:
			sb.WriteString(v)
		case encoding.TextMarshaler:
			text, err := v.MarshalText()
			if err != nil {
				return nil, fmt.Errorf("failed to marshal entry: %w", err)
			}
			sb.Write(text)
		default:
			// Maybe the type has a pointer receiver for MarshalText?
			if ptr, ok := any(&entry).(encoding.TextMarshaler); ok {
				text, err := ptr.MarshalText()
				if err != nil {
					return nil, fmt.Errorf("failed to marshal entry: %w", err)
				}
				sb.Write(text)
			} else {
				sb.WriteString(fmt.Sprintf("%v", entry))
			}
		}
	}

	return []byte(sb.String()), nil
}

func (l *CommaDelimited[T]) UnmarshalText(text []byte) error {
	items := strings.Split(string(text), ",")
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}

		var entry T

		switch v := any(&entry).(type) {
		case *string:
			*v = item
		case encoding.TextUnmarshaler:
			if err := v.UnmarshalText([]byte(item)); err != nil {
				return fmt.Errorf("failed to unmarshal entry: %w", err)
			}
		default:
			_, err := fmt.Sscanf(item, "%v", &entry)
			if err != nil {
				return fmt.Errorf("unable to unmarshal entry: %w", err)
			}
		}

		*l = append(*l, entry)
	}

	return nil
}
