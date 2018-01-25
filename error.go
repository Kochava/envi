// Copyright 2015 Kochava. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in LICENSE.txt.

package envi

import (
	"errors"
	"fmt"
	"reflect"
	"strconv"
)

// IsNoValue is a convenience function returning whether the given error is a no-value error (i.e.,
// is either a ErrNoValue or a KeyError containing ErrNoValue).
func IsNoValue(err error) bool {
	if err == nil {
		return false
	}

	if ke, ok := err.(*KeyError); ok {
		err = ke.Err
	}
	return err == ErrNoValue
}

// TypeError represents any error that occurs as a result of loading an environment variable into
// the error's Type.
type TypeError struct {
	Type reflect.Type
}

func (t *TypeError) Error() string {
	return "couldn't convert string to " + t.Type.String()
}

// SyntaxError is any syntax error encountered as a result of parsing values.
type SyntaxError struct {
	Str string
	Err error
}

func mksyntaxerr(str string, err error) error {
	return &SyntaxError{Str: str, Err: err}
}

func (e *SyntaxError) Error() string {
	return "syntax error: cannot parse " + strconv.Quote(e.Str) + errstr(e.Err)
}

func errstr(e error) string {
	if e == nil {
		return ""
	}
	if s, ok := e.(fmt.Stringer); ok {
		return ": " + s.String()
	}
	return ": " + e.Error()
}

// KeyError is any error encountered when unmarshaling a specific key.
type KeyError struct {
	Key string
	Err error
}

func newKeyError(key string, err error) error {
	return &KeyError{Key: key, Err: err}
}

// NoValueError returns a new KeyError containing ErrNoValue and the given key. This is mainly used
// for implementations of Unmarshaler and Env.
func NoValueError(key string) error {
	return newKeyError(key, ErrNoValue)
}

func (k *KeyError) Error() string {
	return k.Key + ": " + k.Err.Error()
}

// ErrInvalidBool is returned if a boolean is not valid.
var ErrInvalidBool = errors.New("bool is not valid")

// ErrNoValue is returned if a key has no value.
var ErrNoValue = errors.New("no value")
