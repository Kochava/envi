// Copyright 2015 Kochava. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in LICENSE.txt.

package envi

import (
	"reflect"
	"testing"
)

func TestValuesMethods(t *testing.T) {
	want := Values{
		"a": {"zug", "zub"},
		"b": {"quux"},
		"c": {},
		"d": {"7", "6", "5", "4", "3", "2", "1"},
	}

	values := make(Values)

	values.Add("a", "zug")
	values.Add("a", "zub")

	values.Add("Bamf", "fontaine")
	values.Add("Bamf", "shortie")

	values.Set("b", "foo")
	values.Set("b", "bar")
	values.Set("b", "baz")
	values.Set("b", "quux")

	values.Insert("d", "1")
	values.Insert("d", "2")
	values.Insert("d", "3")
	values.Insert("d", "4")
	values.Insert("d", "5")
	values.Insert("d", "6")
	values.Insert("d", "7")

	values["c"] = []string{}

	values.Del("Bamf")

	if !reflect.DeepEqual(want, values) {
		t.Errorf("values = %q;\n    \nwant %q", values, want)
	}

	if vs, err := values.Getenv("a"); err != nil || vs != "zug" {
		t.Errorf("values.Getenv(%q) = %#v, %#v; want %#v, %#v", "a", vs, err, want["a"], nil)
	}

	if vs, err := values.GetenvAll("a"); err != nil || vs == nil {
		t.Errorf("values.GetenvAll(%q) = %#v, %#v; want %#v, %#v", "a", vs, err, want["a"], nil)
	}

	if vs, err := values.Getenv("c"); err != nil || vs != "" {
		t.Errorf("values.Getenv(%q) = %#v, %#v; want %#v, %#v", "a", vs, err, want["c"], nil)
	}

	if vs, err := values.GetenvAll("c"); err != nil || vs == nil {
		t.Errorf("values.GetenvAll(%q) = %#v, %#v; want %#v, %#v", "a", vs, err, want["c"], nil)
	}

	wanterr := NoValueError("none")
	if vs, err := values.Getenv("none"); !IsNoValue(err) || vs != "" {
		t.Errorf("values.Getenv(%q) = %#v, %#v; want %#v, %#v", "a", vs, err, "", wanterr)
	}

	if vs, err := values.GetenvAll("none"); !IsNoValue(err) || vs != nil {
		t.Errorf("values.GetenvAll(%q) = %#v, %#v; want %#v, %#v", "a", vs, err, nil, wanterr)
	}
}
