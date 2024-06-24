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

package dependency

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"

	"github.com/dpeckett/debby/internal/types/arch"
)

// Parse a string into a Dependency object. The input should look something
// like "foo, bar | baz".
func Parse(in string) (Dependency, error) {
	var result Dependency
	return result, parseDependency(in, &result)
}

// MustParse is a helper function to wrap Parse and panic on error.
func MustParse(in string) Dependency {
	result, err := Parse(in)
	if err != nil {
		panic(err)
	}
	return result
}

func peekRune(reader *bufio.Reader) (rune, error) {
	r, _, err := reader.ReadRune()
	if err != nil {
		return -1, err
	}
	if err := reader.UnreadRune(); err != nil {
		return r, err
	}
	return r, nil
}

func eatWhitespace(reader *bufio.Reader) {
	for {
		peek, err := peekRune(reader)
		if err != nil {
			return
		}
		switch peek {
		case '\r', '\n', ' ', '\t':
			_, _, _ = reader.ReadRune()
			continue
		}
		break
	}
}

func parseDependency(in string, ret *Dependency) error {
	reader := bufio.NewReader(bytes.NewReader([]byte(in)))

	eatWhitespace(reader) /* Clean out leading whitespace */

	for {
		peek, err := peekRune(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		switch peek {
		case ',': /* Next relation set */
			_, _, _ = reader.ReadRune()
			eatWhitespace(reader)
			continue
		}

		if err := parseRelation(reader, ret); err != nil {
			return err
		}
	}
}

func parseRelation(reader *bufio.Reader, dependency *Dependency) error {
	eatWhitespace(reader) /* Clean out leading whitespace */

	ret := &Relation{Possibilities: []Possibility{}}

	for {
		peek, err := peekRune(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				goto PARSE_RELATION_DONE
			}
			return err
		}

		switch peek {
		case ',': /* Done with this relation! yay */
			goto PARSE_RELATION_DONE
		case '|': /* Next Possibility */
			_, _, _ = reader.ReadRune()
			eatWhitespace(reader)
			continue
		}
		if err := parsePossibility(reader, ret); err != nil {
			return err
		}
	}

PARSE_RELATION_DONE:
	dependency.Relations = append(dependency.Relations, *ret)
	return nil
}

func parsePossibility(reader *bufio.Reader, relation *Relation) error {
	eatWhitespace(reader) /* Clean out leading whitespace */

	peek, err := peekRune(reader)
	if err != nil {
		return err
	}

	if peek == '$' {
		/* OK, nice. So, we've got a substvar. Let's eat it. */
		return parseSubstvar(reader, relation)
	}

	/* Otherwise, let's punt and build it up ourselves. */

	ret := &Possibility{
		Name:          "",
		Version:       nil,
		Architectures: &ArchSet{Architectures: []arch.Arch{}},
		StageSets:     []StageSet{},
		Substvar:      false,
	}

	for {
		peek, err := peekRune(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				goto PARSE_POSSIBILITY_DONE
			}
			return err
		}

		switch peek {
		case ':':
			err := parseMultiarch(reader, ret)
			if err != nil {
				return err
			}
			continue
		case ' ', '(':
			err := parsePossibilityControllers(reader, ret)
			if err != nil {
				return err
			}
			continue
		case ',', '|': /* I'm out! */
			goto PARSE_POSSIBILITY_DONE
		}
		/* Not a control, let's append */
		next, _, _ := reader.ReadRune()
		ret.Name += string(next)
	}

PARSE_POSSIBILITY_DONE:
	if ret.Name == "" {
		return nil // e.g. trailing comma in Build-Depends
	}
	relation.Possibilities = append(relation.Possibilities, *ret)
	return nil
}

func parseSubstvar(reader *bufio.Reader, relation *Relation) error {
	eatWhitespace(reader)
	_, _, _ = reader.ReadRune() /* Assert ch == '$' */
	_, _, _ = reader.ReadRune() /* Assert ch == '{' */

	ret := &Possibility{
		Name:     "",
		Version:  nil,
		Substvar: true,
	}

	for {
		peek, err := peekRune(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("reached EOF before Substvar finished: %w", err)
			}
			return err
		}

		if peek == '}' {
			_, _, _ = reader.ReadRune()
			relation.Possibilities = append(relation.Possibilities, *ret)
			return nil
		}
		next, _, _ := reader.ReadRune()
		ret.Name += string(next)
	}
}

func parseMultiarch(reader *bufio.Reader, possi *Possibility) error {
	_, _, _ = reader.ReadRune() /* mandated to be a : */
	name := ""
	for {
		peek, err := peekRune(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				goto PARSE_MULTIARCH_DONE
			}
			return err
		}

		switch peek {
		case ',', '|', ' ', '(', '[', '<':
			goto PARSE_MULTIARCH_DONE
		default:
			next, _, _ := reader.ReadRune()
			name += string(next)
		}
	}

PARSE_MULTIARCH_DONE:
	archObj, err := arch.Parse(name)
	if err != nil {
		return err
	}
	possi.Arch = &archObj
	return nil
}

func parsePossibilityControllers(reader *bufio.Reader, possi *Possibility) error {
	for {
		eatWhitespace(reader) /* Clean out leading whitespace */
		peek, err := peekRune(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return nil
			}
			return err
		}

		switch peek {
		case ',', '|':
			return nil
		case '(':
			if possi.Version != nil {
				return errors.New("only one Version relation per Possibility")
			}
			err := parsePossibilityVersion(reader, possi)
			if err != nil {
				return err
			}
			continue
		case '[':
			if len(possi.Architectures.Architectures) != 0 {
				return errors.New("only one Arch relation per Possibility")
			}
			err := parsePossibilityArchs(reader, possi)
			if err != nil {
				return err
			}
			continue
		case '<':
			err := parsePossibilityStageSet(reader, possi)
			if err != nil {
				return err
			}
			continue
		}
		return fmt.Errorf("trailing garbage in Possibility: %s", string(peek))
	}
}

func parsePossibilityVersion(reader *bufio.Reader, possi *Possibility) error {
	eatWhitespace(reader)
	_, _, _ = reader.ReadRune() /* mandated to be ( */
	version := VersionRelation{}

	err := parsePossibilityOperator(reader, &version)
	if err != nil {
		return err
	}

	err = parsePossibilityNumber(reader, &version)
	if err != nil {
		return err
	}

	_, _, _ = reader.ReadRune() /* mandated to be ) */
	possi.Version = &version
	return nil
}

func parsePossibilityOperator(reader *bufio.Reader, version *VersionRelation) error {
	eatWhitespace(reader)
	leader, _, err := reader.ReadRune()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("reached EOF before Operator finished: %w", err)
		}
		return err
	}

	if leader == '=' {
		version.Operator = "="
		return nil
	}

	secondary, _, err := reader.ReadRune()
	if err != nil {
		if errors.Is(err, io.EOF) {
			return fmt.Errorf("reached EOF before Operator finished: %w", err)
		}
		return err
	}

	operator := string([]rune{leader, secondary})

	switch operator {
	case ">=", "<=", "<<", ">>":
		version.Operator = operator
		return nil
	}

	return fmt.Errorf("unknown Operator in Possibility Version modifier: %s", operator)
}

func parsePossibilityNumber(reader *bufio.Reader, version *VersionRelation) error {
	eatWhitespace(reader)
	versionStr := ""
	for {
		peek, err := peekRune(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("reached EOF before Number finished: %w", err)
			}
			return err
		}

		if peek == ')' {
			return version.Version.UnmarshalText([]byte(versionStr))
		}

		next, _, err := reader.ReadRune()
		if err != nil {
			return fmt.Errorf("error reading next rune: %w", err)
		}
		versionStr += string(next)
	}
}

func parsePossibilityArchs(reader *bufio.Reader, possi *Possibility) error {
	eatWhitespace(reader)
	_, _, _ = reader.ReadRune() /* Assert ch == '[' */

	for {
		peek, err := peekRune(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("reached EOF before Arch list finished: %w", err)
			}
			return err
		}

		if peek == ']' {
			_, _, _ = reader.ReadRune()
			return nil
		}

		if err := parsePossibilityArch(reader, possi); err != nil {
			return err
		}
	}
}

func parsePossibilityArch(reader *bufio.Reader, possi *Possibility) error {
	eatWhitespace(reader)
	name := ""

	peek, err := peekRune(reader)
	if err != nil {
		return err
	}
	hasNot := peek == '!'
	if hasNot {
		_, _, _ = reader.ReadRune() // '!'
	}
	if len(possi.Architectures.Architectures) == 0 {
		possi.Architectures.Not = hasNot
	} else if possi.Architectures.Not != hasNot {
		return errors.New("cannot mix negated and non-negated architectures")
	}

	for {
		peek, err := peekRune(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("reached EOF before Arch list finished: %w", err)
			}
			return err
		}

		switch peek {
		case '!':
			return errors.New("you can only negate whole blocks")
		case ']', ' ': /* Let our parent deal with both of these */
			archObj, err := arch.Parse(name)
			if err != nil {
				return err
			}
			possi.Architectures.Architectures = append(possi.Architectures.Architectures, archObj)
			return nil
		}
		next, _, _ := reader.ReadRune()
		name += string(next)
	}
}

func parsePossibilityStageSet(reader *bufio.Reader, possi *Possibility) error {
	eatWhitespace(reader)
	_, _, _ = reader.ReadRune() /* Assert ch == '<' */

	stageSet := StageSet{}
	for {
		peek, err := peekRune(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("reached EOF before StageSet finished: %w", err)
			}
			return err
		}

		if peek == '>' {
			_, _, _ = reader.ReadRune()
			possi.StageSets = append(possi.StageSets, stageSet)
			return nil
		}

		if err := parsePossibilityStage(reader, &stageSet); err != nil {
			return err
		}
	}
}

func parsePossibilityStage(reader *bufio.Reader, stageSet *StageSet) error {
	eatWhitespace(reader)

	stage := Stage{}
	for {
		peek, err := peekRune(reader)
		if err != nil {
			if errors.Is(err, io.EOF) {
				return fmt.Errorf("reached EOF before Stage finished: %w", err)
			}
			return err
		}

		switch peek {
		case '!':
			_, _, _ = reader.ReadRune()
			if stage.Not {
				return errors.New("double negation in Stage is not permitted")
			}
			stage.Not = !stage.Not
		case '>', ' ': /* Let our parent deal with both of these */
			stageSet.Stages = append(stageSet.Stages, stage)
			return nil
		}
		next, _, _ := reader.ReadRune()
		stage.Name += string(next)
	}
}
