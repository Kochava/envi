// Copyright 2015 Kochava. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in LICENSE.txt.

package envi

import (
	"os"
	"testing"
)

func testOSEnv(t *testing.T, getenv func(interface{}, string) error) {
	const name = "ENVI_TEST_VAR_VALUE"
	const value = "123456789"
	const want = 123456789

	if err := os.Setenv(name, value); err != nil {
		t.Skipf("os.Setenv() = %v; skipping", err)
	}

	num := -1

	err := getenv(&num, name)
	if err != nil {
		t.Errorf("Getenv() = %v; want nil", err)
	}

	if num != want {
		t.Errorf("Getenv() num = %d; want %d", num, want)
	}

	if err := os.Unsetenv(name); err != nil {
		t.Fatalf("os.Unsetenv() = %v; skipping", err)
	}

	num = -1
	err = getenv(&num, name)
	if !IsNoValue(err) {
		t.Errorf("Getenv() = %v; want IsNoValue(err)", err)
	}

	if want := -1; num != want {
		t.Errorf("Getenv() num = %d; want %d", num, want)
	}
}

func TestGlobalGetenv(t *testing.T) {
	testOSEnv(t, Getenv)
}

func TestDefaultEnv(t *testing.T) {
	testOSEnv(t, DefaultReader.Getenv)
}

func TestOSEnv(t *testing.T) {
	r := Reader{Source: OSEnv}
	testOSEnv(t, r.Getenv)
}

func TestEnvFunc(t *testing.T) {
	r := Reader{Source: EnvFunc(OSEnv.Getenv)}
	testOSEnv(t, r.Getenv)
}
