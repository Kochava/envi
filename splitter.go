// Copyright 2015 Kochava. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in LICENSE.txt.

package envi

import "strings"

// Data formatting / splitting stuff

// Splitter is an interface used by a Reader when the environment source cannot provide multiple
// strings per key. In those cases, a key's value is split into multiple values using a Splitter.
type Splitter interface {
	SplitString(key, value string) []string
}

// SplitterFunc is a callback Splitter.
type SplitterFunc func(k, v string) []string

var _ Splitter = SplitterFunc(nil)

// SplitString invokes the underlying callback with the given key and value. This implements
// Splitter.
func (fn SplitterFunc) SplitString(key, val string) []string {
	return fn(key, val)
}

// SplitterFuncNoKey is a callback Splitter that does not receive a key argument.
type SplitterFuncNoKey func(v string) []string

var _ Splitter = SplitterFuncNoKey(nil)

// SplitString invokes the underlying callback with the given value. This implements Splitter.
func (fn SplitterFuncNoKey) SplitString(_, val string) []string {
	return fn(val)
}

// StringSplitter is a splitter that splits strings on itself (via strings.Split).
//
// For example, StringSplitter(",") will split input values on commas.
type StringSplitter string

var _ Splitter = StringSplitter("")

// SplitString splits the input string, val, using the string form of the receiver, delim. This
// implements Splitter.
func (delim StringSplitter) SplitString(_, val string) []string {
	return strings.Split(val, string(delim))
}

// FormatSplitter is a Splitter that, after splitting a value, uses a Format function to reformat
// each value from the split. By default, the splitter used is strings.Fields and the format
// function is strings.TrimSpace -- defaults take effect when Split or Format is nil, respectively.
type FormatSplitter struct {
	Split  Splitter            // optional; defaults to strings.Fields
	Format func(string) string // optional; defaults to strings.TrimSpace
}

var _ Splitter = FormatSplitter{}

// SplitString splits the given string value using its Split field, or strings.Fields if nil, and
// then reformats the resulting strings using its Format function (strings.TrimSpace if nil). This
// implements Splitter.
func (t FormatSplitter) SplitString(key, val string) (dst []string) {
	fn := t.Format
	if fn == nil {
		fn = strings.TrimSpace
	}

	if t.Split == nil {
		dst = strings.Fields(val)
	} else {
		dst = t.Split.SplitString(key, val)
	}

	for i, s := range dst {
		dst[i] = fn(s)
	}
	return dst
}
