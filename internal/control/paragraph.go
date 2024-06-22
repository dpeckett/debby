// SPDX-License-Identifier: MPL-2.0
/*
 * Copyright (C) 2024 Damian Peckett <damian@pecke.tt>.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * Portions of this file are based on code originally from: github.com/paultag/go-debian
 *
 * Copyright (c) Paul R. Tagliamonte <paultag@debian.org>, 2015
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package control

import (
	"fmt"
	"io"
	"sort"
	"strings"

	mapset "github.com/deckarep/golang-set/v2"
)

// A Paragraph is a block of RFC2822-like key value pairs.
type Paragraph map[string]string

func (p Paragraph) Set(key, value string) {
	p[key] = value
}

func (p Paragraph) WriteTo(w io.Writer, order []string) (total int64, err error) {
	knownFields := mapset.NewSet(order...)

	var missingFieldNames []string
	for key := range p {
		if !knownFields.Contains(key) {
			missingFieldNames = append(missingFieldNames, key)
		}
	}

	sort.Strings(missingFieldNames)

	orderedFieldNames := make([]string, 0, len(order)+len(missingFieldNames))
	for _, fieldName := range order {
		if _, ok := p[fieldName]; ok {
			orderedFieldNames = append(orderedFieldNames, fieldName)
		}
	}
	orderedFieldNames = append(orderedFieldNames, missingFieldNames...)

	for _, fieldName := range orderedFieldNames {
		value := p[fieldName]
		value = strings.Replace(value, "\n", "\n ", -1)
		value = strings.Replace(value, "\n \n", "\n .\n", -1)

		n, err := w.Write([]byte(fmt.Sprintf("%s: %s\n", fieldName, value)))
		total += int64(n)
		if err != nil {
			return total, err
		}
	}

	return
}
