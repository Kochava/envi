// Copyright 2015 Kochava. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in LICENSE.txt.

package envi

// Values is an Env- and Multienv-conformant map of keys to strings. Getenv will only return the first value held by the
// key's slice. If the slice is empty but the key is set, it returns the empty string. GetenvAll will return nil and
// ErrNoValue only if the key is unset.
type Values map[string][]string

var _ = Multienv(Values(nil))

// Add appends the value to the slice of values held by key.
func (v Values) Add(key, value string) { v[key] = append(v[key], value) }

// Del deletes the key from the receiver.
func (v Values) Del(key string) { delete(v, key) }

// Set assigns the given value for the key, erasing any other values held by it.
func (v Values) Set(key, value string) { v[key] = []string{value} }

// Getenv returns the first value held by the key. If the key is defined, but the slice is empty, it returns the empty
// string. If the key is not set, it returns the empty string and ErrNoValue.
func (v Values) Getenv(key string) (value string, err error) {
	if vals, ok := v[key]; !ok {
		return "", NoValueError(key)
	} else if len(vals) > 0 {
		return vals[0], nil
	}
	return "", nil
}

// GetenvAll returns all values held by the key. If the key is undefined, it returns nil and ErrNoValue.
func (v Values) GetenvAll(key string) ([]string, error) {
	if vals, ok := v[key]; ok {
		return vals, nil
	}
	return nil, NoValueError(key)
}
