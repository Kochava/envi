// Copyright 2015 Kochava. All rights reserved.
// Use of this source code is governed by a BSD-style license that can be found in LICENSE.txt.

package envi

import (
	"encoding"
	"fmt"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// defaultMaxSliceLen is the maximum slice length used when a Reader a maximum of 0 (zero) or less.
const defaultMaxSliceLen = 1000

// DefaultReader is the default envi *Reader used by the Getenv and Load functions. It may be
// changed, but it is not safe to modify it while in use.
var DefaultReader = &Reader{Sep: "_"}

// Reader defines the instance type of an envi unmarshaler.
type Reader struct {
	// Source provides environment variable values. If Source is also a Multienv, its GetenvAll implementation is
	// used in place of Split to get multiple values from the environment.
	Source Env
	// Split is used to split a value from Source into multiple values. If Source is a Multienv, Split is unused.
	Split Splitter
	// Sep is the separator used to join the prefix key and field names when unmarshaling
	// structs (and only structs).
	Sep string
	// MaxSliceLen is the maximum number of environment variables that will be checked for slice
	// values. If MaxSliceLen is less than 1, this defaults to 1000.
	MaxSliceLen int
}

// Getenv attempts to load the value held by the environment variable key into dst. If an error
// occurs, it will return that error. If the environment variable identified by key is not set,
// a no-value error is returned.
//
// If the environment variable is empty for those types without interfaces available, the behavior
// depends on that type. Strings are assigned an empty string, and byte slices are truncated to an
// empty slice.
//
// If the environment value cannot be read, it returns the error provided by the Env source. If the
// error is ErrNoValue, it is returned as a *KeyError with the key attached.
//
// Decoding rules for structs and slices are described under (*Reader).Load.
func (r *Reader) Getenv(dst interface{}, key string) (err error) {
	defer swallowLoadPanic("Getenv", key, &err)
	val, err := r.getenv(key)
	if err != nil && !(IsNoValue(err) && tryNoVal(dst)) {
		if err == ErrNoValue {
			err = NoValueError(key)
		}
		return err
	}
	return r.load(dst, val, key)
}

// getenv retrieves a value from the Reader's environment variable source if configured. If no
// source is configured, however, it looks up the value using os.LookupEnv (where, if the variable
// is not found, it returns a NoValueError).
func (r *Reader) getenv(key string) (v string, err error) {
	if r != nil && r.Source != nil {
		return r.Source.Getenv(key)
	} else if v, ok := os.LookupEnv(key); ok {
		return v, nil
	}
	return "", ErrNoValue
}

// Load attempts to parse the given value (identified as key, which is occasionally relevant when
// assigning names to file descriptors) and store the result in dst. This can be used to circumvent
// envi's standard data source and use your own (e.g., INI files) without overriding the reader's
// source, as it's unused inside of Load.
//
// If dst is a slice and an error occurs, dst is still assigned the value of partially unmarshaling
// the slice (by default using strings.Fields to separate slice values in the environment variable).
// This is only relevant if, for example, you unmarshal a slice of files, connections, listeners, or
// something else that must be closed, as Getenv will not close anything that is the result of
// incorrectly unmarshaling a slice.
//
// Struct field decoding can be configured by using the 'envi' field tag. The first value of a field
// tagis always the name of the environment variable suffix, after a separator. Subsequent fields
// are flags.
//
// For example:
//
//   type Example struct {
//       // Unmarshal from ${Prefix}${Sep}SUFFIX
//       NamedField int `envi:"SUFFIX"`
//       // Set a custom separator of "___" -- unmarshals from ${Prefix}___SepField
//       SepField int `envi:",sep=___"`
//       // Use the field name -- unmarshals from ${Prefix}${Sep}UseFieldName
//       UseFieldName int `envi:""`
//       // Skip this field entirely
//       SkipField int `envi:"-"`
//       // Ignore errors on this field
//       QuietField int `envi:",quiet"
//   }
//
// Flags may be specified in any order, and the last flag seen of its type is the one that is used.
// The quiet flag is used to ignore unmarshaling failures.
// The sep flag sets a custom separator -- this does not allow commas.
// Fields with the suffix "-" are not unmarshaled into, and fields with an empty suffix use their
// field name as the suffix (without any change in case).
//
// Slices of slices are supported but will only ever contain slices of single values.
func (r *Reader) Load(dst interface{}, val, key string) (err error) {
	defer swallowLoadPanic("Load", key, &err)
	return r.load(dst, val, key)
}

func (r *Reader) load(dst interface{}, val, key string) (err error) {
	var ok bool
	if ok, err = loadTypeSwitch(dst, val, key); !ok {
		err = r.loadReflect(dst, val, key)
	}
	if err == ErrNoValue {
		err = NoValueError(key)
	}
	return err
}

func loadTypeSwitch(dst interface{}, val, key string) (ok bool, err error) {
	switch out := dst.(type) {
	case Unmarshaler:
		err = out.UnmarshalEnv(key, val)
	case encoding.TextUnmarshaler:
		err = loadTextUnmarshaler(out, val)
	case *string:
		*out = val
	case *[]byte:
		*out = append((*out)[:0], val...)
	case *bool:
		err = loadBool(out, val)
	case *url.URL:
		err = loadURL(out, val)
	case **url.URL:
		err = loadURLIndirect(out, val)
	case *time.Duration:
		err = loadDuration(out, val)
	default:
		return false, nil
	}
	return true, err
}

func loadTextUnmarshaler(out encoding.TextUnmarshaler, val string) error {
	// Return a no-value error if unmarshaling an empty string into a TextUnmarshaler.
	// We can assume that an envi.Unmarshaler knows how to handle empty values and will
	// understand it, but TextUnmarshalers shouldn't be treated that way.
	if val == "" {
		return ErrNoValue
	}
	return out.UnmarshalText([]byte(val))
}

func loadBool(out *bool, val string) error {
	if val == "" {
		return ErrNoValue
	}
	b, err := parseBool(val)
	if err == nil {
		*out = b
	}
	return err
}

func loadURL(out *url.URL, val string) error {
	if val == "" {
		return ErrNoValue
	}
	u, err := url.Parse(val)
	if err == nil {
		*out = *u
	}
	return err
}

func loadURLIndirect(out **url.URL, val string) error {
	if val == "" {
		return ErrNoValue
	}
	u, err := url.Parse(val)
	if err == nil {
		*out = u
	}
	return err
}

func loadDuration(out *time.Duration, val string) error {
	if val == "" {
		return ErrNoValue
	}
	d, err := time.ParseDuration(val)
	if err == nil {
		*out = d
	}
	return err
}

func (r *Reader) loadReflect(dst interface{}, val, key string) error {
	switch out := indirect(reflect.ValueOf(dst)); out.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return loadInt(out, val)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return loadUint(out, val)
	case reflect.Float32, reflect.Float64:
		return loadFloat(out, val)
	case reflect.Slice:
		return r.loadSlice(out, val, key)
	case reflect.Struct:
		return r.loadStruct(out, key)
	}
	return &TypeError{reflect.TypeOf(dst)}
}

func loadInt(out reflect.Value, val string) error {
	if val == "" {
		return ErrNoValue
	}
	i, err := strconv.ParseInt(val, 0, out.Type().Bits())
	if err != nil {
		return mksyntaxerr(val, err)
	}
	out.SetInt(i)
	return nil
}

func loadUint(out reflect.Value, val string) error {
	if val == "" {
		return ErrNoValue
	}
	u, err := strconv.ParseUint(val, 0, out.Type().Bits())
	if err != nil {
		return mksyntaxerr(val, err)
	}
	out.SetUint(u)
	return err
}

func loadFloat(out reflect.Value, val string) error {
	if val == "" {
		return ErrNoValue
	}
	f, err := strconv.ParseFloat(val, out.Type().Bits())
	if err != nil {
		return mksyntaxerr(val, err)
	}
	out.SetFloat(f)
	return nil
}

func allocindirect(v reflect.Value) reflect.Value {
	switch v.Kind() {
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
	default:
		if v.CanAddr() {
			v = v.Addr()
		}
	}
	return v
}

var (
	unmarshalerType     = reflect.TypeOf((*Unmarshaler)(nil)).Elem()
	textUnmarshalerType = reflect.TypeOf((*encoding.TextUnmarshaler)(nil)).Elem()
)

func isMarshalerType(t reflect.Type) bool {
	return t.Implements(unmarshalerType) || t.Implements(textUnmarshalerType)
}

func isSplitType(t reflect.Type) bool {
	if isMarshalerType(t) {
		return true
	} else if t.Kind() == reflect.Ptr {
		t = t.Elem()
		if isMarshalerType(t) {
			return true
		}
	}

	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.String, reflect.Bool:
		return true
	}
	return false
}

func (r *Reader) splitstring(key, value string) []string {
	if me, ok := r.Source.(Multienv); ok {
		if all, err := me.GetenvAll(key); err != nil {
			// Will be caught by Load.
			panic(err)
		} else {
			return all
		}
	} else if r != nil && r.Split != nil {
		return r.Split.SplitString(key, value)
	}
	return strings.Fields(value)
}

func (r *Reader) loadSliceSeq(out reflect.Value, elemtype reflect.Type, val, key string) (ok bool, err error) {
	if kind := elemtype.Kind(); val != "" && kind != reflect.Ptr && kind != reflect.Struct {
		return false, nil
	}

	var (
		slice  = reflect.MakeSlice(out.Type(), 0, 1)
		maxLen = r.MaxSliceLen
	)

	if maxLen <= 0 {
		maxLen = defaultMaxSliceLen
	}

	for i := 1; i <= maxLen && err == nil; i++ {
		elemKey := key + r.Sep + strconv.Itoa(i)
		tmp := reflect.New(elemtype)
		dst := allocindirect(indirect(tmp))
		if err = r.Getenv(dst.Interface(), elemKey); IsNoValue(err) {
			break
		}
		slice = reflect.Append(slice, tmp.Elem())
	}

	if slice.Len() > 0 {
		out.Set(slice)
		if IsNoValue(err) {
			err = nil
		}
	}

	return true, err
}

func (r *Reader) loadSliceSplit(out reflect.Value, elemtype reflect.Type, val, key string) (err error) {
	var (
		vals   = r.splitstring(key, val)
		slice  = reflect.MakeSlice(out.Type(), len(vals), len(vals))
		loaded = 0
	)

	for i, v := range vals {
		elemKey := key + r.Sep + strconv.Itoa(1+i)
		if err = r.Load(slice.Index(i).Addr().Interface(), v, elemKey); err != nil {
			// Do set it, because there might be open resources held by the slice (files) that
			// haven't been closed. It's on the person who called Getenv to clean up anything that
			// needs closing after an error.
			break
		}
		loaded++
		out.Set(slice.Slice(0, loaded))
	}

	if loaded > 0 {
		out.Set(slice.Slice(0, loaded))
		// Discard no-value error if we just hit something empty while walking keys
		if IsNoValue(err) {
			err = nil
		}
	}

	return err
}

func (r *Reader) loadSlice(out reflect.Value, val, key string) error {
	elemtype := out.Type().Elem()
	ok, err := r.loadSliceSeq(out, elemtype, val, key)
	if ok && (err == nil || !IsNoValue(err) || !isSplitType(elemtype)) {
		return err
	} else if val == "" {
		return NoValueError(key)
	}

	return r.loadSliceSplit(out, elemtype, val, key)
}

func (r *Reader) loadStruct(out reflect.Value, key string) (err error) {
	// NOTE: struct loading ignores ErrNoValue
	typ := out.Type()
	empty := true

	var isset bool
	for i, n := 0, typ.NumField(); i < n; i++ {
		isset, err = r.loadStructField(out, typ, i, key)
		if err != nil {
			break
		}
		empty = empty && !isset
	}

	if empty && err == nil {
		return NoValueError(key)
	}

	return err
}

func (r *Reader) loadStructField(out reflect.Value, typ reflect.Type, fieldIdx int, key string) (isset bool, err error) {
	f := typ.Field(fieldIdx)
	if f.PkgPath != "" {
		return
	}

	var (
		// NOTE: this does not respect escapes or anything -- don't know why you'd
		// need them but you shouldn't plan on having them
		etag  = strings.Split(f.Tag.Get("envi"), ",")
		fname = etag[0]
		flags = structFlags{sep: r.Sep}
	)
	flags.parse(etag[1:])
	if fname == "-" {
		return
	}
	fname = flags.fieldName(key, etag[0], f.Name)

	var (
		field  = out.FieldByIndex(f.Index)
		tmp    = reflect.New(field.Type()).Elem()
		target = allocindirect(indirect(tmp))
		dst    = target.Interface()
	)

	// Only copy field's value to temporary storage if it's valid
	if fi := reflect.Indirect(field); fi.IsValid() {
		target.Elem().Set(fi)
	}
	if err = r.Getenv(dst, fname); err != nil && !IsNoValue(err) && !flags.quiet {
		return isset, err
	}
	if flags.quiet {
		if !IsNoValue(err) {
			field.Set(tmp)
			isset = true
		}
	} else if err == nil {
		field.Set(tmp)
		isset = true
	}

	return isset, nil
}

// tryNoVal returns whether a no-value error should be ignored for the given destination interface.
// This currently only covers structs and values that implement Unmarshaler.
//
// encoding.TextUnmarshaler gets a pass right now.
func tryNoVal(dst interface{}) bool {
	if _, ok := dst.(Unmarshaler); ok {
		return true
	}

	v := reflect.ValueOf(dst)
	if v.Kind() != reflect.Ptr {
		return false
	}

	kind := v.Type().Elem().Kind()
	return kind == reflect.Struct || kind == reflect.Slice
}

func swallowLoadPanic(fn string, key string, err *error) {
	if err == nil || *err != nil {
		// There's no storage or there's already an error, in which case a panic is
		// unexpected.
		return
	}
	switch rc := recover(); perr := rc.(type) {
	case nil:
	case error:
		if perr == ErrNoValue {
			perr = &KeyError{Key: key, Err: perr}
		}
		*err = perr
	default:
		*err = fmt.Errorf("panic: %s(%q): %v", fn, key, perr)
	}
}

type structFlags struct {
	quiet bool
	sep   string
}

func (flags *structFlags) fieldName(key, tagName, fieldName string) (name string) {
	name = tagName
	if name == "" {
		name = fieldName
	}
	if key != "" {
		name = key + flags.sep + name
	}
	return name
}

func (flags *structFlags) parse(tags []string) {
	const (
		fSep   = "sep="
		fQuiet = "quiet"
	)

	for _, t := range tags {
		switch {
		case strings.HasPrefix(t, fSep):
			flags.sep = t[len(fSep):]
		case t == fQuiet:
			flags.quiet = true
		}
	}
}
