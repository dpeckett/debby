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
	"errors"
	"io"
	"reflect"
	"time"

	"github.com/go-viper/mapstructure/v2"
)

// Marshal is a one-off interface to serialize a single object to a writer.
//
// Most notably, this will *not* separate Paragraphs with a newline as is
// expected upon repeated calls, please use the Encoder streaming interface
// for that.
//
// Given a struct (or list of structs), write to the io.Writer stream
// in the RFC822-alike Debian control-file format
//
// This code will attempt to unpack it into the struct based on the
// literal name of the key, This can be overridden by the struct tag
// `control:""`.
func Marshal(writer io.Writer, data any) error {
	encoder, err := NewEncoder(writer)
	if err != nil {
		return err
	}
	return encoder.Encode(data)
}

// Encoder is a struct that allows for the streaming Encoding of data
// back out to an `io.Writer`. Most notably, this will separate
// subsequent `Encode` calls of a Struct with a newline.
//
// It's also worth noting that this *will* also write out elements that
// were Unmarshaled into a Struct without a member of that name if (and only
// if) the target Struct contains a `control.Paragraph` anonymous member.
//
// This is handy if the Unmarshaler was given any `X-*` keys that were not
// present on your Struct.
//
// Given a struct (or list of structs), write to the io.Writer stream
// in the RFC822-alike Debian control-file format
//
// This code will attempt to unpack it into the struct based on the
// literal name of the key, This can be overridden by the struct tag
// `control:""`.
//
// If you're dehydrating a list of strings, you have the option of defining
// a string to join the tokens with (`delim:", "`).
//
// In order to Marshal a custom Struct, you are required to implement the
// Marshallable interface. It's highly encouraged to put this interface on
// the struct without a pointer receiver, so that pass-by-value works
// when you call Marshal.
type Encoder struct {
	writer         io.Writer
	alreadyWritten bool
}

// Create a new Encoder, which is configured to write to the given `io.Writer`.
func NewEncoder(writer io.Writer) (*Encoder, error) {
	return &Encoder{
		writer:         writer,
		alreadyWritten: false,
	}, nil
}

// Take a Struct, Encode it into a Paragraph, and write that out to the
// io.Writer set up when the Encoder was configured.
func (e *Encoder) Encode(incoming interface{}) error {
	data := reflect.ValueOf(incoming)
	return e.encode(data)
}

func (e *Encoder) encode(data reflect.Value) error {
	if data.Type().Kind() == reflect.Ptr {
		return e.encode(data.Elem())
	}

	switch data.Type().Kind() {
	case reflect.Slice:
		return e.encodeSlice(data)
	case reflect.Struct:
		return e.encodeStruct(data)
	}
	return errors.New("unknown type")
}

func (e *Encoder) encodeSlice(data reflect.Value) error {
	for i := 0; i < data.Len(); i++ {
		if err := e.encodeStruct(data.Index(i)); err != nil {
			return err
		}
	}
	return nil
}

func (e *Encoder) encodeStruct(data reflect.Value) error {
	if e.alreadyWritten {
		_, err := e.writer.Write([]byte("\n"))
		if err != nil {
			return err
		}
	}
	paragraph, err := convertToParagraph(data)
	if err != nil {
		return err
	}
	e.alreadyWritten = true
	_, err = paragraph.WriteTo(e.writer)
	return err
}

func convertToParagraph(data reflect.Value) (Paragraph, error) {
	if data.Type().Kind() != reflect.Struct {
		return nil, errors.New("can only Decode a Struct")
	}

	values := make(map[string]string)
	decoder, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		Result:           &values,
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			timeToStringHookFunc("Mon, 02 Jan 2006 15:04:05 MST"),
			mapstructure.TextMarshallerHookFunc(),
			boolToYesNoHookFunc(),
			sliceToStringHookFunc(),
		),
		TagName: "control",
	})
	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(data.Interface()); err != nil {
		return nil, err
	}

	return Paragraph(values), nil
}

func timeToStringHookFunc(layout string) mapstructure.DecodeHookFunc {
	return func(from, to reflect.Type, data any) (any, error) {
		if from != reflect.TypeOf(time.Time{}) {
			return data, nil
		}
		if to.Kind() != reflect.String {
			return data, nil
		}

		return data.(time.Time).Format(layout), nil
	}
}
