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

/* Package dependency provides an interface to parse and inspect Debian
 * Dependency relationships.
 *
 * Dependency               |               foo, bar (>= 1.0) [amd64] | baz
 *	-> Relations            | -> Relation        bar (>= 1.0) [amd64] | baz
 *			 -> Possibilities   | -> Possibility     bar (>= 1.0) [amd64]
 *					| Name          | -> Name            bar
 *					| Version       | -> Version             (>= 1.0)
 *					| Architectures | -> Arch                          amd64
 *					| Stages        |
 */
package dependency

import (
	"strings"

	"github.com/dpeckett/debby/internal/types/arch"
	"github.com/dpeckett/debby/internal/types/version"
)

// ArchSet models an architecture dependency restriction, commonly used to
// restrict the relation to one some architectures. This is also usually
// used in a string of many possibilities.
type ArchSet struct {
	Not           bool
	Architectures []arch.Arch
}

func (set ArchSet) String() string {
	if len(set.Architectures) == 0 {
		return ""
	}
	not := ""
	if set.Not {
		not = "!"
	}
	arches := []string{}
	for _, arch := range set.Architectures {
		arches = append(arches, not+arch.String())
	}
	return "[" + strings.Join(arches, " ") + "]"
}

// VersionRelation models a version restriction on a possibility, such as
// greater than version 1.0, or less than 2.0. The values that are valid
// in the Operator field are defined by section 7.1 of Debian policy.
//
//	The relations allowed are <<, <=, =, >= and >> for strictly earlier,
//	earlier or equal, exactly equal, later or equal and strictly later,
//	respectively.
type VersionRelation struct {
	Version  version.Version
	Operator string
}

func (ver VersionRelation) String() string {
	return "(" + ver.Operator + " " + ver.Version.String() + ")"
}

// Stage models a build stage that a Possibility may be restricted to. For
// example, a Possibility may only be satisfied during the build stage "build".
type Stage struct {
	Not  bool
	Name string
}

func (stage Stage) String() string {
	if stage.Not {
		return "!" + stage.Name
	}
	return stage.Name
}

// StageSet models a set of build stages that a Possibility may be restricted
// to. For example, a Possibility may only be satisfied during the build
// stages "build" and "host".
type StageSet struct {
	Stages []Stage
}

func (set StageSet) String() string {
	if len(set.Stages) == 0 {
		return ""
	}
	stages := []string{}
	for _, stage := range set.Stages {
		stages = append(stages, stage.String())
	}
	return "<" + strings.Join(stages, " ") + ">"
}

// Possibility models a concrete Possibility that may be satisfied in order
// to satisfy the Dependency Relation. Given the Dependency line:
//
//	Depends: foo, bar | baz
//
// All of foo, bar and baz are Possibilities. Possibilities may come with
// further restrictions, such as restrictions on Version, Architecture, or
// Build Stage.
type Possibility struct {
	Name          string
	Arch          *arch.Arch
	Architectures *ArchSet
	StageSets     []StageSet
	Version       *VersionRelation
	Substvar      bool
}

func (pos Possibility) String() string {
	str := pos.Name
	if pos.Arch != nil {
		str += ":" + pos.Arch.String()
	}
	if pos.Architectures != nil {
		if arch := pos.Architectures.String(); arch != "" {
			str += " " + arch
		}
	}
	if pos.Version != nil {
		str += " " + pos.Version.String()
	}
	for _, stageSet := range pos.StageSets {
		if stages := stageSet.String(); stages != "" {
			str += " " + stages
		}
	}
	return str
}

// A Relation is a set of Possibilities that must be satisfied. Given the
// Dependency line:
//
//	Depends: foo, bar | baz
//
// There are two Relations, one composed of foo, and another composed of
// bar and baz.
type Relation struct {
	Possibilities []Possibility
}

func (rel Relation) String() string {
	possis := []string{}
	for _, possi := range rel.Possibilities {
		possis = append(possis, possi.String())
	}
	return strings.Join(possis, " | ")
}

// A Dependency is the top level type that models a full Dependency relation.
type Dependency struct {
	Relations []Relation
}

func (dep Dependency) String() string {
	relations := []string{}
	for _, rel := range dep.Relations {
		relations = append(relations, rel.String())
	}
	return strings.Join(relations, ", ")
}

func (dep Dependency) MarshalText() ([]byte, error) {
	return []byte(dep.String()), nil
}

func (dep *Dependency) UnmarshalText(text []byte) error {
	return parseDependency(string(text), dep)
}
