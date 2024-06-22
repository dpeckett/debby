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

// NewLineDelimited is a list of T entries separated by newlines.
type NewLineDelimited[T any] []T

func (l NewLineDelimited[T]) MarshalText() ([]byte, error) {
	var sb strings.Builder
	for i, entry := range l {
		if i > 0 {
			sb.WriteString("\n")
		}

		switch v := any(entry).(type) {
		case string:
			sb.WriteString(v)
		case encoding.TextMarshaler:
			text, err := v.MarshalText()
			if err != nil {
				return nil, fmt.Errorf("failed to marshal entry: %w", err)
			}
			sb.WriteRune(' ')
			sb.Write(text)
		default:
			// Maybe the type has a pointer receiver for MarshalText?
			if ptr, ok := any(&entry).(encoding.TextMarshaler); ok {
				text, err := ptr.MarshalText()
				if err != nil {
					return nil, fmt.Errorf("failed to marshal entry: %w", err)
				}
				sb.WriteRune(' ')
				sb.Write(text)
			} else {
				sb.WriteString(fmt.Sprintf(" %v", entry))
			}
		}
	}

	return []byte(sb.String()), nil
}

func (l *NewLineDelimited[T]) UnmarshalText(text []byte) error {
	lines := strings.Split(string(text), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var entry T

		switch v := any(&entry).(type) {
		case *string:
			*v = line
		case encoding.TextUnmarshaler:
			if err := v.UnmarshalText([]byte(line)); err != nil {
				return fmt.Errorf("failed to unmarshal entry: %w", err)
			}
		default:
			_, err := fmt.Sscanf(line, "%v", &entry)
			if err != nil {
				return fmt.Errorf("unable to unmarshal entry: %w", err)
			}
		}

		*l = append(*l, entry)
	}

	return nil
}
