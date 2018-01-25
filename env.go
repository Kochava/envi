// Copyright 2015 Kochava. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in LICENSE.txt.

package envi

import "os"

// Env is an environment variable source. Given an identifying key, it must return a corresponding
// value. If the resulting value is the empty string, it is considered unset for certain types.
type Env interface {
	Getenv(key string) (string, error)
}

// Multienv is an environment variable source capable of returning multiple environment variables. It must conform to
// Env, but in those cases, it is better for Env to return only the first of all results. Results will not be split when
// an Env is also a Multienv.
type Multienv interface {
	Env
	GetenvAll(key string) ([]string, error)
}

type osenv int

func (osenv) Getenv(key string) (value string, err error) {
	value, ok := os.LookupEnv(key)
	if !ok {
		return "", NoValueError(key)
	}
	return value, nil
}

// OSEnv is an Env implementation that can be used to simply return os.Getenv values. This will
// allow empty values provided the environment variable is defined (i.e., LookupEnv returns a value
// and OK=true).
const OSEnv osenv = 0

// EnvFunc is a callback implementing Env.
type EnvFunc func(string) (string, error)

// Getenv implements Env.
func (fn EnvFunc) Getenv(key string) (string, error) { return fn(key) }

// MultiEnvFunc is a function to return multiple environment values. It implements Env by returning
// only the first value in the result slice. If the result slice is empty, it returns
// a no-value error.
type MultiEnvFunc func(string) ([]string, error)

// GetenvAll implements Multienv.
func (fn MultiEnvFunc) GetenvAll(key string) ([]string, error) {
	return fn(key)
}

// Getenv implements Env.
func (fn MultiEnvFunc) Getenv(key string) (string, error) {
	v, err := fn(key)
	if err != nil {
		return "", err
	}
	if len(v) == 0 {
		return "", NoValueError(key)
	}
	return v[0], nil
}
