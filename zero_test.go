// Copyright 2015 Kochava. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in LICENSE.txt.

package envi

import (
	"reflect"
	"testing"
)

type parentConfig struct {
	Child childConfig `envi:"child"`
}

type childConfig struct {
	Left  string `envi:"left"`
	Right string `envi:"right"`
}

type singleValues map[string]string

func (v singleValues) Getenv(key string) (value string, err error) {
	val, ok := v[key]
	if !ok || len(val) == 0 {
		return "", NoValueError(key)
	}
	return val, nil
}

func TestParentChildDecode(t *testing.T) {
	var (
		parent = parentConfig{Child: childConfig{Left: "left", Right: "right"}}
		want   = parentConfig{Child: childConfig{Left: "still_left", Right: "right"}}
		vals   = singleValues{"parent_child_left": "still_left"}
		r      = Reader{Source: vals, Sep: "_"}
	)

	const prefix = "parent"

	if err := r.Getenv(&parent, prefix); err != nil {
		t.Errorf("Getenv(&parent, %q) err = %v; want nil", prefix, err)
	}

	if !reflect.DeepEqual(parent, want) {
		t.Errorf("Getenv(&parent, %q) = %#+ v\nwant %#+ v", prefix, parent, want)
	}
}
