// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 */

package control

import (
	"reflect"
	"strconv"
	"strings"

	"github.com/go-viper/mapstructure/v2"
)

func yesNoToBoolHookFunc() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Kind, data any) (any, error) {
		if from != reflect.String || to != reflect.Bool {
			return data, nil
		}

		val, err := strconv.ParseBool(data.(string))
		if err == nil {
			return val, nil
		}

		return data.(string) == "yes", nil
	}
}

func boolToYesNoHookFunc() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Kind, data any) (any, error) {
		if from != reflect.Bool || to != reflect.String {
			return data, nil
		}

		if data.(bool) {
			return "yes", nil
		}
		return "no", nil
	}
}

func stringToSliceHookFunc() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Kind, data any) (any, error) {
		if from != reflect.String || to != reflect.Slice {
			return data, nil
		}

		raw := data.(string)
		if raw == "" {
			return []string{}, nil
		}

		fields := strings.Fields(raw)
		for i, field := range fields {
			fields[i] = strings.Trim(strings.TrimSpace(field), ",")
		}

		return fields, nil
	}
}

func sliceToStringHookFunc() mapstructure.DecodeHookFunc {
	return func(from, to reflect.Kind, data any) (any, error) {
		if from != reflect.Slice || to != reflect.String {
			return data, nil
		}

		return strings.Join(data.([]string), ", "), nil
	}
}
