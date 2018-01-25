// Copyright 2015 Kochava. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in LICENSE.txt.

package envi

// TODO: Clean up this mess of tests.

import (
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

type boolUnmarshaler bool

func (b *boolUnmarshaler) UnmarshalEnv(key, value string) error {
	if value == "TRUE" {
		*b = true
	} else if value == "FALSE" {
		*b = false
	} else if value == "" {
		return ErrNoValue
	} else if value == "NONE" {
		panic("value may not be none: " + value)
	} else {
		// Get caught by Getenv recover
		panic(fmt.Errorf("invalid bool-like: %q", value))
	}
	return nil
}

type valueUnmarshaler int

var errValueUnmarshaler = errors.New("value unmarshaler returned error")

func (valueUnmarshaler) UnmarshalEnv(key, value string) error {
	return errValueUnmarshaler
}

type textErrorOnEmpty struct{}

func (textErrorOnEmpty) UnmarshalText(text []byte) error {
	if len(text) == 0 {
		return errors.New("empty")
	}
	return nil
}

type textBoolUnmarshaler bool

func (t *textBoolUnmarshaler) UnmarshalText(text []byte) error {
	return (*boolUnmarshaler)(t).UnmarshalEnv("", string(text))
}

type unmarshalerDummy int

func (unmarshalerDummy) UnmarshalEnv(key, value string) error { return nil }

type unmarshalerDummyPtr int

func (*unmarshalerDummyPtr) UnmarshalEnv(key, value string) error { return nil }

type SingleSource map[string]string

func (s SingleSource) Getenv(key string) (string, error) {
	return s[key], nil
}

func TestTextUnmarshalEmpty(t *testing.T) {
	var (
		r = Reader{
			Source: SingleSource{
				"test": "a",
			},
			Sep: "_",
		}
		err error

		dst  textErrorOnEmpty
		sdst []textErrorOnEmpty
	)
	if err := r.Getenv(dst, "test"); err != nil {
		t.Errorf("Getenv(textErrorOnEmpty) = %v; want nil", err)
	}
	if err = r.Getenv(&sdst, "test"); err != nil {
		t.Errorf("Getenv([]textErrorOnEmpty) = %v; want nil", err)
	}
}

func TestPermittedNoValue(t *testing.T) {
	var x interface{}
	cases := []struct {
		in   interface{}
		want bool
	}{
		// Bad: values, non-pointers
		{"string", false},
		{1234, false},
		{uintptr(1234), false},
		{float64(1234), false},
		{byte(123), false},
		{struct{}{}, false},
		{struct{ x int }{}, false},
		{nil, false},
		{&x, false},       // Pointer to wrong type
		{new(int), false}, // Not a pointer to an integer once deref'd
		{[]byte{'a', 'b', 'c'}, false},
		{[]int{1, 2, 3}, false},
		{new(interface{}), false},
		{new(Unmarshaler), false},
		{unmarshalerDummyPtr(0), false}, // Pointer receiver
		{new(*int), false},              // Pointer to pointer, but not a **struct or **slice

		// OK: pointer to a pointer, a struct, or a slice
		{new(struct{}), true},            // Pointer to struct
		{new(struct{ x int }), true},     // Pointer to struct
		{new([]int), true},               // Pointer to slice
		{new([]byte), true},              // Pointer to slice
		{new(unmarshalerDummyPtr), true}, // Pointer receiver
		{unmarshalerDummy(0), true},      // Value receiver (useless, but checks iface)
	}

	for i, c := range cases {
		i, c := i, c
		t.Run(fmt.Sprintf("%d %T", i, c.in), func(t *testing.T) {
			b := tryNoVal(c.in)
			if b != c.want {
				t.Fatalf("tryNoVal(%#v) = %t; want %t", c.in, b, c.want)
			}
		})
	}
}

func TestParseBool(t *testing.T) {
	cases := []struct {
		in        string
		want, bad bool
	}{
		// Valid true
		{"1", true, false},
		{"t", true, false},
		{"T", true, false},
		{"true", true, false},
		{"True", true, false},
		{"TRUE", true, false},
		{"yes", true, false},
		{"Yes", true, false},
		{"YES", true, false},
		{"on", true, false},
		{"On", true, false},
		{"ON", true, false},

		// Valid false
		{"0", false, false},
		{"f", false, false},
		{"F", false, false},
		{"false", false, false},
		{"False", false, false},
		{"FALSE", false, false},
		{"no", false, false},
		{"No", false, false},
		{"NO", false, false},
		{"off", false, false},
		{"Off", false, false},
		{"OFF", false, false},

		// Invalid
		{"", false, true},
		{"n", false, true},
		{"nO", false, true},
		{"fAlSe", false, true},
		{"boop", false, true},
		{"oFF", false, true},
		{"tRUE", false, true},
		{"yES", false, true},
		{"yEs", false, true},
		{"yeS", false, true},
		{"YeS", false, true},
		{"YEs", false, true},
		{"bool", false, true},
		{"2", false, true},
		{"3", false, true},
		{"-1", false, true},
		{"0.5", false, true},
		{"truefalse", false, true},
		{"falsetrue", false, true},
		{"01", false, true},
		{"10", false, true},
	}

	for i, c := range cases {
		i, c := i, c
		t.Run(fmt.Sprintf("%d %q", i, c.in), func(t *testing.T) {
			b, err := parseBool(c.in)

			if (err != nil) != c.bad {
				want := "nil"
				if c.bad {
					want = "error"
				}
				t.Errorf("parseBool(%q) = _, %v; want %s", c.in, err, want)
			}

			if b != c.want {
				t.Fatalf("parseBool(%q) = %v, _; want %v", c.in, b, c.want)
			}
		})
	}
}

type testData struct {
	Def   int             // Key is "Def"
	Val   int             `envi:"blob"`             // Key is "blob"
	Sep   int             `envi:"sepd,sep=_,sep=:"` // Seperator is :
	Q     []int           `envi:"q,quiet"`          // Errors squelched
	Spec  boolUnmarshaler `envi:"bool"`             // Unmarshaler
	Empty int             `envi:"-"`                // Skip

	unexported boolUnmarshaler // Skip
}

type nestedData struct {
	Values []testData `envi:"v"`
	Bob    testData   `envi:"bob"`
}

type nestedPointerData struct {
	Values []*testData `envi:"v"`
}

type getenvTest struct {
	keyer func() string
	dst   func() interface{}
	want  interface{}
	bad   bool
	empty bool
	in    Env
}

func (c *getenvTest) key() string {
	if c.keyer != nil {
		return c.keyer()
	}
	return "test"
}

type nestedValueMarshal struct {
	Value valueUnmarshaler `envi:"v"`
}

func derefTestValue(v interface{}) interface{} {
	if rv := reflect.ValueOf(v); rv.Kind() == reflect.Ptr && rv.Type().Elem().Kind() != reflect.Struct {
		return rv.Elem().Interface()
	}
	return v
}

func (c *getenvTest) Run(t *testing.T) {
	r := Reader{
		Source: c.in,
		Split:  FormatSplitter{StringSplitter(","), strings.TrimSpace},
		Sep:    "_",
	}

	defer func(r *Reader) { DefaultReader = r }(DefaultReader)
	DefaultReader = &r

	var (
		dst    = c.dst()
		dstype = reflect.TypeOf(dst)

		err      = Getenv(dst, c.key())
		gotbad   = err != nil
		gotempty = IsNoValue(err)
	)

	t.Logf("env = %q", c.in)

	// Check if we got an error as expected
	if gotbad != (c.bad || c.empty) || gotempty != c.empty {
		want := "nil"
		if c.empty {
			want = ErrNoValue.Error()
		} else if c.bad {
			want = "error"
		}
		t.Errorf("env.Getenv(%v) = %v;\n    want %s", dstype, err, want)
	} else if gotbad || gotempty {
		t.Logf("* env.Getenv(%v) = %v", dstype, err)
	}

	// Check if the unmarshaled value is correct
	if got := derefTestValue(dst); (!c.bad || c.want != nil) && !reflect.DeepEqual(got, c.want) {
		t.Errorf("env.Getenv(%v) =>\n    dst = (%T) %#v;\n    want (%T) %#v", dstype, got, got, c.want, c.want)
	} else if c.bad || gotbad || c.empty || gotempty {
		t.Logf("* env.Getenv(%v) =>\n    dst = (%T) %#v", dstype, got, got)
	}
}

func TestLoad(t *testing.T) {
	r := Reader{
		Split: FormatSplitter{StringSplitter(","), strings.TrimSpace},
		Sep:   "_",
	}

	defer func(r *Reader) { DefaultReader = r }(DefaultReader)
	DefaultReader = &r

	const (
		want = 12345601
		in   = "0xbc6101"
	)
	var num int
	if err := Load(&num, in, "number"); err != nil {
		t.Errorf("Load(%p, %q, key) = %v; want nil", &num, in, err)
	}

	if num != want {
		t.Errorf("num = %d; want %d", num, want)
	}
}

func testGetenv(t *testing.T, cases []getenvTest) {
	for i, c := range cases {
		i, c := i, c
		name := fmt.Sprintf("%d/%T", i, c.dst())
		t.Run(name, c.Run)
	}
}

func TestUnmarshaler(t *testing.T) {
	var (
		mkbool       = func() interface{} { return new(boolUnmarshaler) }
		mktextbool   = func() interface{} { return new(textBoolUnmarshaler) }
		mkvmarshaler = func() interface{} { return new(nestedValueMarshal) }
	)
	testGetenv(t, []getenvTest{
		// env
		{dst: mkbool, want: boolUnmarshaler(true), in: envone("TRUE")},
		{dst: mkbool, want: boolUnmarshaler(false), in: envone("FALSE")},
		{dst: mkbool, empty: true, want: boolUnmarshaler(false), in: envone("")},
		{dst: mkbool, bad: true, want: boolUnmarshaler(false), in: envone("invalid-value")},
		{dst: mkbool, bad: true, want: boolUnmarshaler(false), in: envone("NONE")},
		// text
		{dst: mktextbool, want: textBoolUnmarshaler(true), in: envone("TRUE")},
		{dst: mktextbool, want: textBoolUnmarshaler(false), in: envone("FALSE")},
		{dst: mktextbool, empty: true, want: textBoolUnmarshaler(false), in: envone("")},
		{dst: mktextbool, bad: true, want: textBoolUnmarshaler(false), in: envone("invalid-value")},
		{dst: mktextbool, bad: true, want: textBoolUnmarshaler(false), in: envone("NONE")},
		{dst: mkvmarshaler, bad: true, want: &nestedValueMarshal{Value: 0}, in: Values{"test_v": []string{"5"}}},
	})
}

func TestGetenvNestedStructs(t *testing.T) {
	testGetenv(t, []getenvTest{
		// Load multiple, nested
		{dst: func() interface{} { return new(nestedData) },
			want: &nestedData{
				Values: []testData{
					{Def: 1, Q: []int{2, 34}, Val: 3, Spec: true, Sep: 4},
					{Def: 5, Q: []int{6, 78}, Val: 7, Spec: false, Sep: 8},
					{Def: 9, Q: []int{10, 11, 12}, Val: 11, Spec: false, Sep: 12},
				},
			},
			in: Values{
				"test_v_1_Def":  {"1"},
				"test_v_1_q":    {"2", "34", "nope"},
				"test_v_1_blob": {"3", "3"},
				"test_v_1_bool": {"TRUE"},
				"test_v_1:sepd": {"4", "4"},

				"test_v_2_Def":  {"5"},
				"test_v_2_q":    {"6", "78"},
				"test_v_2_blob": {"7", "3"},
				"test_v_2_bool": {"FALSE"},
				"test_v_2:sepd": {"8", "4"},

				"test_v_3_Def":  {"9"},
				"test_v_3_q":    {"10", "11", "12"},
				"test_v_3_blob": {"11", "3"},
				"test_v_3:sepd": {"12", "4"},
			},
		},

		// Load multiple, keep defaults
		{dst: func() interface{} { return &nestedData{Bob: testData{Def: 1234}} },
			want: &nestedData{
				// Value member
				Bob: testData{
					Def:  1234,
					Spec: true,
				},
			},
			in: Values{
				"test_bob_bool": {"TRUE"},
			},
		},

		// Load multiple with no root key, keep defaults
		{dst: func() interface{} { return &nestedData{Bob: testData{Def: 1234}} },
			keyer: func() string { return "" },
			want: &nestedData{
				// Value member
				Bob: testData{
					Def:  1234,
					Spec: true,
				},
			},
			in: Values{
				"bob_bool": {"TRUE"},
			},
		},

		// Load none, nested
		{dst: func() interface{} { return new(nestedData) },
			empty: true,
			want:  &nestedData{},
			in:    Values{},
		},

		// Load multiple, nested with pointers
		{dst: func() interface{} { return new(nestedPointerData) },
			want: &nestedPointerData{
				Values: []*testData{
					&testData{Def: 1, Q: []int{2, 34}, Val: 3, Spec: true, Sep: 4},
					&testData{Def: 5, Q: []int{6, 78}, Val: 7, Spec: false, Sep: 8},
					&testData{Def: 9, Q: []int{10, 11, 12}, Val: 11, Spec: false, Sep: 12},
				},
			},
			in: Values{
				"test_v_1_Def":  {"1"},
				"test_v_1_q":    {"2", "34", "nope"},
				"test_v_1_blob": {"3", "3"},
				"test_v_1_bool": {"TRUE"},
				"test_v_1:sepd": {"4", "4"},

				"test_v_2_Def":  {"5"},
				"test_v_2_q":    {"6", "78"},
				"test_v_2_blob": {"7", "3"},
				"test_v_2_bool": {"FALSE"},
				"test_v_2:sepd": {"8", "4"},

				"test_v_3_Def":  {"9"},
				"test_v_3_q":    {"10", "11", "12"},
				"test_v_3_blob": {"11", "3"},
				"test_v_3:sepd": {"12", "4"},
			},
		},

		// Load none, nested
		{dst: func() interface{} { return new(nestedPointerData) },
			empty: true,
			want:  &nestedPointerData{},
			in:    Values{},
		},
	})
}

func TestGetenvStructs(t *testing.T) {
	testGetenv(t, []getenvTest{
		// Standard unmarshaling of all fields
		{dst: func() interface{} { return &testData{Q: []int{-1}, Empty: -2} },
			want: &testData{Def: 1, Val: 2, Sep: 3, Q: []int{12, 34}, Empty: -2},
			in: Values{
				"test_Def":   {"1", "2"},
				"test_q":     {"12", "34", "nope"},
				"test_blob":  {"2", "3"},
				"test:sepd":  {"3", "4"},
				"test_Empty": {"1234"},
			},
		},

		// Leave Def alone
		{dst: func() interface{} { return &testData{Def: 1234, Q: []int{-1}, Empty: -2} },
			want: &testData{Def: 1234, Val: 2, Sep: 3, Q: []int{12, 34}, Empty: -2},
			in: Values{
				"test_q":     {"12", "34", "nope"},
				"test_blob":  {"2", "3"},
				"test:sepd":  {"3", "4"},
				"test_Empty": {"1234"},
			},
		},

		// Leave Def alone
		{dst: func() interface{} { return &testData{Q: []int{-1}, Empty: -2} },
			// I don't want to try to guarantee that certain fields will be loaded in an erroneous case
			bad: true,
			in: Values{
				"test_Def":   {"nope"},
				"test_q":     {"12", "34", "nope"},
				"test_blob":  {"2", "3"},
				"test:sepd":  {"3", "4"},
				"test_Empty": {"1234"},
			},
		},

		// Load multiple
		{dst: func() interface{} { return new([]testData) },
			want: []testData{
				{Def: 1, Q: []int{2, 34}, Val: 3, Sep: 4},
				{Def: 5, Q: []int{6, 78}, Val: 7, Sep: 8},
				{Def: 9, Q: []int{10, 11, 12}, Val: 11, Sep: 12},
			},
			in: Values{
				"test_1_Def":  {"1"},
				"test_1_q":    {"2", "34", "nope"},
				"test_1_blob": {"3", "3"},
				"test_1:sepd": {"4", "4"},

				"test_2_Def":  {"5"},
				"test_2_q":    {"6", "78"},
				"test_2_blob": {"7", "3"},
				"test_2:sepd": {"8", "4"},

				"test_3_Def":  {"9"},
				"test_3_q":    {"10", "11", "12"},
				"test_3_blob": {"11", "3"},
				"test_3:sepd": {"12", "4"},
			},
		},

		// Load none
		{dst: func() interface{} { return new([]testData) },
			want:  []testData(nil),
			empty: true,
			in:    Values{},
		},
	})
}

func TestGetenvURL(t *testing.T) {
	var (
		newurl    = func() interface{} { return new(url.URL) }
		newurlptr = func() interface{} { return new(*url.URL) }

		valid = func(s string) getenvTest {
			u, err := url.Parse(s)

			// If the string is empty or there's an error, we still have a URL instance
			// (since we're doing Getenv(*url.URL), not Getenv(**url.URL), to load the
			// URL from the environment), it's just empty.
			if u == nil {
				u = new(url.URL)
			}

			return getenvTest{
				dst:   newurl,
				in:    envone(s),
				want:  u,
				bad:   err != nil,
				empty: s == "",
			}
		}

		validptr = func(s string) getenvTest {
			u, err := url.Parse(s)
			if s == "" {
				// If the string is empty for **url.URL, then nothing is assigned
				u = nil
			}

			return getenvTest{
				dst:   newurlptr,
				in:    envone(s),
				want:  u,
				bad:   err != nil,
				empty: s == "",
			}
		}
	)

	testGetenv(t, []getenvTest{
		valid("http://foo:bar@wub.com/"),
		validptr("http://foo:bar@wub.com/"),
		valid(""),
		validptr(""),
		valid(":"),
		validptr(":"),
	})
}

func TestGetenvBools(t *testing.T) {
	newbool := func() interface{} { return new(bool) }
	testGetenv(t, []getenvTest{
		// Borrowed from parseBool test
		// Valid true
		{dst: newbool, in: envone("1"), want: true},
		{dst: newbool, in: envone("t"), want: true},
		{dst: newbool, in: envone("T"), want: true},
		{dst: newbool, in: envone("true"), want: true},
		{dst: newbool, in: envone("True"), want: true},
		{dst: newbool, in: envone("TRUE"), want: true},
		{dst: newbool, in: envone("yes"), want: true},
		{dst: newbool, in: envone("Yes"), want: true},
		{dst: newbool, in: envone("YES"), want: true},
		{dst: newbool, in: envone("on"), want: true},
		{dst: newbool, in: envone("On"), want: true},
		{dst: newbool, in: envone("ON"), want: true},

		// Valid false
		{dst: newbool, in: envone("0"), want: false},
		{dst: newbool, in: envone("f"), want: false},
		{dst: newbool, in: envone("F"), want: false},
		{dst: newbool, in: envone("false"), want: false},
		{dst: newbool, in: envone("False"), want: false},
		{dst: newbool, in: envone("FALSE"), want: false},
		{dst: newbool, in: envone("no"), want: false},
		{dst: newbool, in: envone("No"), want: false},
		{dst: newbool, in: envone("NO"), want: false},
		{dst: newbool, in: envone("off"), want: false},
		{dst: newbool, in: envone("Off"), want: false},
		{dst: newbool, in: envone("OFF"), want: false},

		// Invalid
		{dst: newbool, in: envone(""), empty: true, want: false},
		{dst: newbool, in: envone("n"), bad: true, want: false},
		{dst: newbool, in: envone("nO"), bad: true, want: false},
		{dst: newbool, in: envone("fAlSe"), bad: true, want: false},
		{dst: newbool, in: envone("boop"), bad: true, want: false},
		{dst: newbool, in: envone("oFF"), bad: true, want: false},
		{dst: newbool, in: envone("tRUE"), bad: true, want: false},
		{dst: newbool, in: envone("yES"), bad: true, want: false},
		{dst: newbool, in: envone("yEs"), bad: true, want: false},
		{dst: newbool, in: envone("yeS"), bad: true, want: false},
		{dst: newbool, in: envone("YeS"), bad: true, want: false},
		{dst: newbool, in: envone("YEs"), bad: true, want: false},
		{dst: newbool, in: envone("bool"), bad: true, want: false},
		{dst: newbool, in: envone("2"), bad: true, want: false},
		{dst: newbool, in: envone("3"), bad: true, want: false},
		{dst: newbool, in: envone("-1"), bad: true, want: false},
		{dst: newbool, in: envone("0.5"), bad: true, want: false},
		{dst: newbool, in: envone("truefalse"), bad: true, want: false},
		{dst: newbool, in: envone("falsetrue"), bad: true, want: false},
		{dst: newbool, in: envone("01"), bad: true, want: false},
		{dst: newbool, in: envone("10"), bad: true, want: false},
	})
}

func TestGetenvIntegers(t *testing.T) {
	testGetenv(t, []getenvTest{
		// positive int
		{dst: func() interface{} { return new(int) },
			want: 1234,
			in:   envone("1234"),
		},
		// negative
		{dst: func() interface{} { return new(int) },
			want: -1234,
			in:   envone("-1234"),
		},
		// hex int
		{dst: func() interface{} { return new(int) },
			want: 0x4d2,
			in:   envone("0x4d2"),
		},
		// hex int
		{dst: func() interface{} { return new(int) },
			want: 0x4d2,
			in:   envone("0x4D2"),
		},
		// octal int
		{dst: func() interface{} { return new(int) },
			want: 0632,
			in:   envone("0632"),
		},
		// not an int
		{dst: func() interface{} { return new(int) },
			bad:  true,
			want: 0,
			in:   envone("0FFF"),
		},

		// Types
		{dst: func() interface{} { return new(int64) },
			want: int64(12345), in: envone("12345")},
		{dst: func() interface{} { return new(int32) },
			want: int32(12345), in: envone("12345")},
		{dst: func() interface{} { return new(int16) },
			want: int16(12345), in: envone("12345")},
		{dst: func() interface{} { return new(int8) },
			want: int8(123), in: envone("123")},
		{dst: func() interface{} { return new(uint64) },
			want: uint64(12345), in: envone("12345")},
		{dst: func() interface{} { return new(uint32) },
			want: uint32(12345), in: envone("12345")},
		{dst: func() interface{} { return new(uint16) },
			want: uint16(12345), in: envone("12345")},
		{dst: func() interface{} { return new(uint8) },
			want: uint8(123), in: envone("123")},

		{dst: func() interface{} { return new(int64) },
			want: int64(-12345), in: envone("-12345")},
		{dst: func() interface{} { return new(int32) },
			want: int32(-12345), in: envone("-12345")},
		{dst: func() interface{} { return new(int16) },
			want: int16(-12345), in: envone("-12345")},
		{dst: func() interface{} { return new(int8) },
			want: int8(-123), in: envone("-123")},

		// Empty
		{dst: func() interface{} { return new(int64) },
			empty: true, want: int64(0), in: envone("")},
		{dst: func() interface{} { return new(int32) },
			empty: true, want: int32(0), in: envone("")},
		{dst: func() interface{} { return new(int16) },
			empty: true, want: int16(0), in: envone("")},
		{dst: func() interface{} { return new(int8) },
			empty: true, want: int8(0), in: envone("")},
		{dst: func() interface{} { return new(uint64) },
			empty: true, want: uint64(0), in: envone("")},
		{dst: func() interface{} { return new(uint32) },
			empty: true, want: uint32(0), in: envone("")},
		{dst: func() interface{} { return new(uint16) },
			empty: true, want: uint16(0), in: envone("")},
		{dst: func() interface{} { return new(uint8) },
			empty: true, want: uint8(0), in: envone("")},

		// Negative uint -> invalid
		{dst: func() interface{} { return new(uint) },
			bad: true, want: uint(0), in: envone("-12345")},
		{dst: func() interface{} { return new(uint64) },
			bad: true, want: uint64(0), in: envone("-12345")},
		{dst: func() interface{} { return new(uint32) },
			bad: true, want: uint32(0), in: envone("-12345")},
		{dst: func() interface{} { return new(uint16) },
			bad: true, want: uint16(0), in: envone("-12345")},
		{dst: func() interface{} { return new(uint8) },
			bad: true, want: uint8(0), in: envone("-123")},

		// Bit size exceeded
		{dst: func() interface{} { return new(uint64) },
			bad: true, want: uint64(0), in: envone("0x1FFFFFFFFFFFFFFFF")},
		{dst: func() interface{} { return new(uint32) },
			bad: true, want: uint32(0), in: envone("0x1FFFFFFFF")},
		{dst: func() interface{} { return new(uint16) },
			bad: true, want: uint16(0), in: envone("0x1FFFF")},
		{dst: func() interface{} { return new(uint8) },
			bad: true, want: uint8(0), in: envone("0x1FF")},

		// Not supported:
		{dst: func() interface{} { return new(uintptr) },
			bad:  true,
			want: uintptr(0),
			in:   envone("12345"),
		},
	})
}

func TestGetenvFloat(t *testing.T) {
	mkf32 := func() interface{} { return new(float32) }
	mkf64 := func() interface{} { return new(float64) }

	testGetenv(t, []getenvTest{
		// positive 32
		{dst: mkf32, want: float32(1234.5678), in: envone("1234.5678")},
		// negative 32
		{dst: mkf32, want: float32(-1234.5678), in: envone("-1234.5678")},

		// positive 64
		{dst: mkf64, want: float64(1234.5678), in: envone("1234.5678")},
		// negative 64
		{dst: mkf64, want: float64(-1234.5678), in: envone("-1234.5678")},

		// invalid 64
		{dst: mkf64, bad: true, want: float64(0), in: envone("x")},
		// invalid 32
		{dst: mkf32, bad: true, want: float32(0), in: envone("x")},

		// empty 64
		{dst: mkf64, empty: true, want: float64(0), in: envone("")},
		// empty 32
		{dst: mkf32, empty: true, want: float32(0), in: envone("")},
	})
}

func TestGetenvStringlike(t *testing.T) {
	mkstr := func() interface{} { return new(string) }
	mkbytes := func() interface{} { return new([]byte) }

	testGetenv(t, []getenvTest{
		// empty string (ErrNoValue)
		{dst: func() interface{} { return new(int) },
			empty: true,
			want:  0,
			in:    envone(""),
		},
		// string
		{dst: mkstr,
			want: "a string",
			in:   envone("a string"),
		},
		// first string
		{dst: mkstr,
			want: "1",
			in:   envmany{"1", "2", "3"},
		},
		// empty string
		{dst: mkstr,
			want: "",
			in:   envone(""),
		},

		// byte slice
		{dst: mkbytes,
			want: []byte("foobar"),
			in:   envone("foobar"),
		},
		// empty byte slice -- preserves nil
		{dst: mkbytes,
			want: []byte(nil),
			in:   envone(""),
		},
		// empty byte slice -- preserves nil
		{dst: func() interface{} { b := make([]byte, 128); return &b },
			want: []byte{},
			in:   envone(""),
		},
	})
}

func TestGetenvScalarSlices(t *testing.T) {
	testGetenv(t, []getenvTest{
		// string slice
		{dst: func() interface{} { return new([]string) },
			want: []string{"a", "b", "c"},
			in:   envone("a,b,c"),
		},
		{dst: func() interface{} { return new([]string) },
			want: []string{"a", "b", "c"},
			in:   envmany{"a", "b", "c"},
		},

		// byte-slice slice
		{dst: func() interface{} { return new([][]byte) },
			want: [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			in:   envone("a,b,c"),
		},
		{dst: func() interface{} { return new([][]byte) },
			want: [][]byte{[]byte("a"), []byte("b"), []byte("c")},
			in:   envmany{"a", "b", "c"},
		},

		// int slice
		{dst: func() interface{} { return new([]int) },
			want: []int{50},
			in:   envone("50"),
		},
		// multi-int slice (env many)
		{dst: func() interface{} { return new([]int) },
			want: []int{50, 51, 52, 64},
			in:   envmany{"50", "51", "52", "64"},
		},
		// multi-int slice (split and trim)
		{dst: func() interface{} { return new([]int) },
			want: []int{50, 51, 52, 64},
			in:   envone("50, 51, 52, 64"),
		},
		// multi-int slice (split and trim -- wrong separator)
		{dst: func() interface{} { return new([]int) },
			bad:  true,
			want: []int(nil),
			in:   envone("50|51|52|64"),
		},
		// multi-int slice (split and trim -- wrong separator, one element)
		{dst: func() interface{} { return new([]int) },
			bad:  true,
			want: []int{50},
			in:   envone("50  ,  51|52|64"),
		},
		// empty int slice
		{dst: func() interface{} { return new([]int) },
			want:  []int(nil),
			empty: true,
			in:    envone(""),
		},
	})
}

func TestGetenvDuration(t *testing.T) {
	newdur := func() interface{} { return new(time.Duration) }
	testGetenv(t, []getenvTest{
		{dst: newdur, empty: true, want: time.Duration(0), in: envone("")},
		{dst: newdur, want: time.Duration(time.Second), in: envone("1s")},
		{dst: newdur,
			want: time.Duration(500*time.Millisecond + 31*time.Second + 2*time.Minute + 33*time.Hour),
			in:   envone("1.5s2m33h30s"),
		},
		{dst: newdur,
			bad:  true,
			want: time.Duration(0),
			in:   envone("not a duration"),
		},
	})
}

type NamedInterface interface{}

func TestGetenvInterface(t *testing.T) {
	type Value struct {
		X NamedInterface `envi:"x,quiet"`
	}
	newiface := func() interface{} { return &Value{X: "do not modify"} }
	testGetenv(t, []getenvTest{
		{
			dst:   newiface,
			want:  &Value{X: "do not modify"},
			empty: true,
			in:    envone("foobar"),
		},
	})
}

type envone string
type envmany []string

var _ Env = envone("foo")
var _ Multienv = envmany(nil)

func (s envone) Getenv(key string) (v string, err error) {
	if key != "test" {
		return "", ErrNoValue
	}
	return string(s), nil
}

func (s envmany) Getenv(key string) (v string, err error) {
	if key != "test" {
		return "", ErrNoValue
	} else if len(s) == 0 {
		return "", nil
	}
	return s[0], nil
}

func (s envmany) GetenvAll(key string) (v []string, err error) {
	if key != "test" {
		return nil, ErrNoValue
	}
	return append(make([]string, 0, len(s)), []string(s)...), nil
}

func TestSplitters(t *testing.T) {
	t.Run("SplitterFunc", func(t *testing.T) {
		var fn SplitterFunc = func(_, val string) []string {
			return strings.Fields(val)
		}
		want := []string{"1", "2", "3,4"}
		got := fn.SplitString("", "1 2 3,4")
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("got %v; want %v", got, want)
		}
	})

	t.Run("SplitterFuncNoKey", func(t *testing.T) {
		var fn SplitterFuncNoKey = strings.Fields
		want := []string{"1", "2", "3,4"}
		got := fn.SplitString("", "1 2 3,4")
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("got %v; want %v", got, want)
		}
	})

	t.Run("FormatSplitterDefaults", func(t *testing.T) {
		var s FormatSplitter
		want := []string{"1", "2", "3,4"}
		got := s.SplitString("", "   1 2 3,4   ")
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("got %v; want %v", got, want)
		}
	})

	t.Run("FormatSplitter", func(t *testing.T) {
		var s = FormatSplitter{
			Split:  StringSplitter("|"),
			Format: strings.TrimSpace,
		}
		want := []string{"1", "2", "3,4"}
		got := s.SplitString("", "   1  | 2  | 3,4   ")
		if !reflect.DeepEqual(want, got) {
			t.Fatalf("got %v; want %v", got, want)
		}
	})
}
