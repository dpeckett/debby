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

package arch

import (
	"errors"
	"strings"
)

type Arch struct {
	ABI string
	OS  string
	CPU string
}

func (arch *Arch) IsWildcard() bool {
	if arch.CPU == "all" {
		return false
	}

	if arch.ABI == "any" || arch.OS == "any" || arch.CPU == "any" {
		return true
	}
	return false
}

func (arch *Arch) Is(other *Arch) bool {
	if arch.IsWildcard() && other.IsWildcard() {
		/* We can't compare wildcards to other wildcards. That's just
		* insanity. We always need a concrete arch. Not even going to try. */
		return false
	} else if arch.IsWildcard() {
		/* OK, so we're a wildcard. Let's defer to the other
		* struct to deal with this */
		return other.Is(arch)
	}

	if (arch.CPU == other.CPU || (arch.CPU != "all" && other.CPU == "any")) &&
		(arch.OS == other.OS || other.OS == "any") &&
		(arch.ABI == other.ABI || other.ABI == "any") {

		return true
	}

	return false
}

func (arch Arch) String() string {
	/* ABI-OS-CPU -- gnu-linux-amd64 */
	els := []string{}
	if arch.ABI != "any" && arch.ABI != "all" && arch.ABI != "gnu" && arch.ABI != "" {
		els = append(els, arch.ABI)
	}

	if arch.OS != "any" && arch.OS != "all" && arch.OS != "linux" {
		els = append(els, arch.OS)
	}

	els = append(els, arch.CPU)
	return strings.Join(els, "-")
}

func (arch Arch) MarshalText() ([]byte, error) {
	return []byte(arch.String()), nil
}

func (arch *Arch) UnmarshalText(text []byte) error {
	return parseArchInto(arch, string(text))
}

// Parse an architecture string into an Arch struct.
func Parse(arch string) (Arch, error) {
	result := Arch{
		ABI: "any",
		OS:  "any",
		CPU: "any",
	}
	return result, parseArchInto(&result, arch)
}

// MustParse is like Parse, but panics on error.
func MustParse(arch string) Arch {
	result, err := Parse(arch)
	if err != nil {
		panic(err)
	}
	return result
}

func parseArchInto(ret *Arch, arch string) error {
	/* May be in the following form:
	* `any` (implicitly any-any-any)
	* kfreebsd-any (implicitly any-kfreebsd-any)
	* kfreebsd-amd64 (implicitly any-kfreebsd-any)
	* bsd-openbsd-i386 */
	flavors := strings.SplitN(arch, "-", 3)
	switch len(flavors) {
	case 1:
		flavor := flavors[0]
		/* OK, we've got a single guy like `any` or `amd64` */
		switch flavor {
		case "all", "any":
			ret.ABI = flavor
			ret.OS = flavor
			ret.CPU = flavor
		default:
			/* right, so we've got something like `amd64`, which is implicitly
			* gnu-linux-amd64. Confusing, I know. */
			ret.ABI = "gnu"
			ret.OS = "linux"
			ret.CPU = flavor
		}
	case 2:
		/* Right, this is something like kfreebsd-amd64, which is implicitly
		* gnu-kfreebsd-amd64 */
		ret.OS = flavors[0]
		ret.CPU = flavors[1]
	case 3:
		/* This is something like bsd-openbsd-amd64 */
		ret.ABI = flavors[0]
		ret.OS = flavors[1]
		ret.CPU = flavors[2]
	default:
		return errors.New("invalid arch string")
	}

	return nil
}
