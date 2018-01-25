// Copyright 2015 Kochava. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in LICENSE.txt.

package envi

import (
	"errors"
	"reflect"
	"testing"
)

func TestDefaultSplit(t *testing.T) {
	// Split on fields by default
	env := envone("a b  c   d     e        f             g")
	want := []string{"a", "b", "c", "d", "e", "f", "g"}
	var got []string

	r := Reader{Source: env}
	if err := r.Getenv(&got, "test"); err != nil {
		t.Errorf("Getenv() err = %v; want nil", err)
	}

	if !reflect.DeepEqual(want, got) {
		t.Errorf("Getenv() = %q; want %q", got, want)
	}
}

func TestErrorInMultiget(t *testing.T) {
	want := NoValueError("test")
	fn := MultiEnvFunc(func(string) ([]string, error) { return nil, want })
	r := Reader{Source: fn}

	var i []string
	if err := r.Getenv(&i, "test"); !IsNoValue(err) {
		t.Errorf("Getenv() err = %v; want %v", err, want)
	}

	if want := []string(nil); !reflect.DeepEqual(i, want) {
		t.Errorf("Getenv() = %#v; want %#v", i, want)
	}
}

func TestPanicInMultiget(t *testing.T) {
	want := errors.New("panicked")
	fn := MultiEnvFunc(func(string) ([]string, error) { panic(want) })
	r := Reader{Source: fn}

	i := -1
	if err := r.Getenv(&i, "test"); err != want {
		t.Errorf("Getenv() err = %v; want %v", err, want)
	}

	if want := -1; i != want {
		t.Errorf("Getenv() = %d; want %d", i, want)
	}
}

func TestValueInMultienvFunc(t *testing.T) {
	fn := MultiEnvFunc(func(key string) ([]string, error) { return nil, nil })
	if _, err := fn.Getenv("test"); !IsNoValue(err) {
		t.Errorf("Getenv() err = %v; want IsNoValue(err)", err)
	}

	want := errors.New("no value")
	fn = func(string) ([]string, error) { return nil, want }

	if _, err := fn.Getenv("test"); err != want {
		t.Errorf("Getenv() err = %v; want %v", err, want)
	}

	wantstr := "foobar"
	wantvals := []string{wantstr, "bazquux", "wimbledog"}
	fn = func(string) ([]string, error) { return wantvals, nil }

	val, err := fn.Getenv("test")
	if err != nil {
		t.Errorf("Getenv() err = %v; want nil", err)
	}

	if wantstr != val {
		t.Errorf("Getenv() = %q; want %q", val, wantstr)
	}

	vals, err := fn.GetenvAll("test")
	if err != nil {
		t.Errorf("Getenv() err = %v; want nil", err)
	}

	if !reflect.DeepEqual(wantvals, vals) {
		t.Errorf("Getenv() = %q; want %q", vals, wantvals)
	}
}
