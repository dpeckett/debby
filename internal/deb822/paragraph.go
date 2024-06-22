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

package deb822

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

// A Paragraph is a block of RFC2822-like key value pairs. This struct contains
// two methods to fetch values, a Map called Values, and a Slice called
// Order, which maintains the ordering as defined in the RFC2822-like block
type Paragraph struct {
	Values map[string]string
	Order  []string
}

func (p *Paragraph) Set(key, value string) {
	if p.Values == nil {
		p.Values = make(map[string]string)
	}

	if _, found := p.Values[key]; found {
		/* We've got the key */
		p.Values[key] = value
		return
	}

	/* Otherwise, go ahead and set it in the order and dict,
	* and call it a day */
	p.Order = append(p.Order, key)
	p.Values[key] = value
}

func (p *Paragraph) WriteTo(w io.Writer) (total int64, err error) {
	for _, key := range p.Order {
		value := p.Values[key]

		value = strings.Replace(value, "\n", "\n ", -1)
		value = strings.Replace(value, "\n \n", "\n .\n", -1)

		n, err := w.Write([]byte(fmt.Sprintf("%s: %s\n", key, value)))
		total += int64(n)
		if err != nil {
			return total, err
		}
	}
	return
}

// MarshalJSON ensures the keys are marshaled in the order specified by Order
func (p Paragraph) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteByte('{')
	encoder := json.NewEncoder(&buf)
	first := true
	for _, key := range p.Order {
		if p.Values[key] == "" {
			continue
		}

		if !first {
			buf.WriteByte(',')
		}
		first = false

		// add key
		if err := encoder.Encode(key); err != nil {
			return nil, err
		}
		buf.WriteByte(':')

		// add value
		if err := encoder.Encode(p.Values[key]); err != nil {
			return nil, err
		}
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}

// UnmarshalJSON ensures the keys are unmarshaled and ordered as they appear in the JSON object
func (p *Paragraph) UnmarshalJSON(data []byte) error {
	if p.Values == nil {
		p.Values = make(map[string]string)
	}

	decoder := json.NewDecoder(bytes.NewReader(data))

	// Read the opening brace
	if _, err := decoder.Token(); err != nil {
		return err
	}

	// Iterate through the JSON object
	for decoder.More() {
		// Read the key
		token, err := decoder.Token()
		if err != nil {
			return err
		}
		key := token.(string)

		// Read the value
		var value string
		if err := decoder.Decode(&value); err != nil {
			return err
		}

		if value != "" {
			p.Set(key, value)
		}
	}

	// Read the closing brace
	if _, err := decoder.Token(); err != nil {
		return err
	}

	return nil
}
