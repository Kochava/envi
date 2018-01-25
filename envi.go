// Copyright 2015 Kochava. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in LICENSE.txt.

// Package envi contains functions for unmarshaling data / configuration info from environment
// variables. These are primarily strings and other scalars, but can also include structs and
// slices.
//
// Envi unmarshals structs by walking each field and concatenating them by a separator, configured
// as part of the Reader. It unmarshals slices by attempting to either split their explicit
// environment variable's value using a delimiter (defaults to whitespace) or by walking var_1
// through var_N, determined by the Reader's MaxSliceLen.
package envi

// Unmarshaler defines an interface that allows a type to declare it supports unmarshaling from
// environment variable text.
type Unmarshaler interface {
	UnmarshalEnv(key, value string) error
}

// Getenv attempts to load the value held by the environment variable key into dst using the
// DefaultReader. If an error occurs, that error is returned. See Reader.Getenv for more
// information.
func Getenv(dst interface{}, key string) error {
	return DefaultReader.Getenv(dst, key)
}

// Load attempts to load the given value identified by key into dst using the DefaultReader. See
// Reader.Load and Reader.Getenv for more information.
func Load(dst interface{}, val, key string) error {
	return DefaultReader.Load(dst, val, key)
}
