// Copyright 2015 Kochava. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in LICENSE.txt.

package envi

import (
	"fmt"
)

func ExampleGetenv() {
	// environment is used in this example to provide a consistent source of environment values.
	// By default, envi will use the process's environment.
	environment := Values{
		"ENVI_INTEGER":        {"0xff"},
		"ENVI_STR":            {"This is a normal string, nothing special"},
		"ENVI_INTEGERS":       {"1", "2", "3", "5", "8", "13", "21", "34"},
		"ENVI_STRS_1":         {"foo"},
		"ENVI_STRS_2":         {"bar"},
		"ENVI_STRS_3":         {"baz"},
		"ENVI_STRS_5":         {"wub"}, // No 4th index, so this is ignored
		"ENVI_STRUCT_VALID":   {"1234567"},
		"ENVI_STRUCT_INVALID": {"caterpillar"},
		"ENVI_STRUCT_IGNORED": {"1234567"},
	}

	// You can either allocate your own Reader or use the default one. The default
	// separates fields in structs by underscores and gets its values from the process
	// environment. This example uses a hardcoded Env and an _ separator for fields and
	// indices.
	reader := Reader{
		Source: environment,
		Sep:    "_",
	}

	var (
		integer  int
		str      string
		integers []int
		strs     []string
		novalue  int = -1

		structured struct {
			Valid   int `envi:"VALID"`
			Invalid int `envi:"INVALID,quiet"` // Ignore invalid values (caterpillar)
			Ignored int `envi:"-"`
		}
	)

	reader.Getenv(&integer, "ENVI_INTEGER")
	fmt.Println(integer)
	reader.Getenv(&str, "ENVI_STR")
	fmt.Println(str)
	reader.Getenv(&integers, "ENVI_INTEGERS")
	fmt.Println(integers)
	reader.Getenv(&strs, "ENVI_STRS")
	fmt.Println(strs)

	if err := reader.Getenv(&structured, "ENVI_STRUCT"); err != nil {
		panic(fmt.Errorf("Error unmarshaling structured value: %v", err))
	}
	fmt.Println(structured)

	// ENVI_NOVALUE has no value, so it won't change and Getenv will return an no-value error.
	if err := reader.Getenv(&novalue, "ENVI_NOVALUE"); !IsNoValue(err) {
		panic("NOVALUE was set -- expected no value!")
	}

	// Output:
	// 255
	// This is a normal string, nothing special
	// [1 2 3 5 8 13 21 34]
	// [foo bar baz]
	// {1234567 0 0}
}
