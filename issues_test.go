package envi

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestPointerFields(t *testing.T) { // issue #2
	runner := func(env Env, dest interface{}, want interface{}) func(*testing.T) {
		return func(t *testing.T) {

			r := Reader{Source: env, Sep: "_"}
			err := r.Getenv(dest, "Data")
			if err != nil && !IsNoValue(err) {
				t.Fatalf("unexpected error: %v", err)
			}

			if !reflect.DeepEqual(dest, want) {
				t.Fatalf("got\t%#+ v\nwant\t%#+ v", dest, want)
			}
		}
	}

	type Data struct {
		Size int
	}

	type Value struct {
		Value Data
	}

	type Pointer struct {
		Value *Data
	}

	t.Run("HasEnv", func(t *testing.T) {
		env := Values{"Data_Value_Size": {"16"}}
		t.Run("StructValue", runner(env, new(Value), &Value{Value: Data{Size: 16}}))
		t.Run("StructPointer", runner(env, new(Pointer), &Pointer{Value: &Data{Size: 16}}))
		t.Run("StructValueAssigned", runner(env, &Value{Value: Data{Size: 32}}, &Value{Value: Data{Size: 16}}))
		t.Run("StructPointerAssigned", runner(env, &Pointer{Value: &Data{Size: 32}}, &Pointer{Value: &Data{Size: 16}}))
	})

	t.Run("WithoutEnv", func(t *testing.T) {
		env := Values{}
		t.Run("StructValue", runner(env, new(Value), &Value{Value: Data{}}))
		t.Run("StructPointer", runner(env, new(Pointer), &Pointer{Value: nil}))
		t.Run("StructValueAssigned", runner(env, &Value{Value: Data{Size: 32}}, &Value{Value: Data{Size: 32}}))
		t.Run("StructPointerAssigned", runner(env, &Pointer{Value: &Data{Size: 32}}, &Pointer{Value: &Data{Size: 32}}))
	})
}

func TestCyclicalTypes(t *testing.T) { // issue #4
	type Cycle struct {
		Value int
		Next  *Cycle
	}

	runner := func(max int, want *Cycle, env Env) func(*testing.T) {
		return func(t *testing.T) {
			dest := new(Cycle)
			r := Reader{Source: env, Sep: "_", MaxDepth: max}
			err := r.Getenv(dest, "Cycle")
			if err != nil && !IsNoValue(err) {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(dest, want) {
				// Encode as JSON just to make it easier to visualize
				jsondest, _ := json.Marshal(dest)
				jsonwant, _ := json.Marshal(want)
				t.Fatalf("got\t%s\nwant\t%s",
					jsondest,
					jsonwant,
				)
			}
		}
	}

	t.Run("DefaultOutsideDepth", runner(
		0, // max depth
		&Cycle{Value: 1},
		Values{ // 1  _ 2  _ 3  _ 4  _ 5  _ 6  _ 7  _ 8  _ 9  _ 10  _ 11
			"Cycle_Value": {"1"},
			"Cycle_Next_Next_Next_Next_Next_Next_Next_Next_Next_Value": {"10"},
		},
	))

	t.Run("DefaultInsideDepth", runner(
		0, // max depth
		&Cycle{
			Value: 1,
			Next:  &Cycle{Next: &Cycle{Next: &Cycle{Next: &Cycle{Next: &Cycle{Next: &Cycle{Next: &Cycle{Next: &Cycle{Value: 9}}}}}}}},
		},
		Values{ // 1  _ 2  _ 3  _ 4  _ 5  _ 6  _ 7  _ 8  _ 9  _ 10  _ 11
			"Cycle_Value": {"1"},
			"Cycle_Next_Next_Next_Next_Next_Next_Next_Next_Value": {"9"},
		},
	))

	t.Run("IgnoreOutOfDepth", runner(
		2, // max depth
		&Cycle{Value: 1},
		Values{ // 1  _ 2  _ 3  _ 4  _ 5  _ 6
			"Cycle_Value":                {"1"},
			"Cycle_Next_Next_Next_Value": {"3"},
		},
	))

	t.Run("ResetInsideDepth", runner(
		2, // max depth
		&Cycle{
			Value: 1,
			Next: &Cycle{
				Next: &Cycle{
					Value: 3,
					Next: &Cycle{
						Next: &Cycle{
							Value: 5,
						},
					},
				},
			},
		},
		Values{ // 1  _ 2  _ 3  _ 4  _ 5  _ 6  _ 7  _ 8
			"Cycle_Value":                                    {"1"},
			"Cycle_Next_Next_Value":                          {"3"},
			"Cycle_Next_Next_Next_Next_Value":                {"5"},
			"Cycle_Next_Next_Next_Next_Next_Next_Next_Value": {"8"},
		},
	))

	t.Run("IgnoreWithoutField", runner(
		2, // max depth
		&Cycle{},
		Values{ // 1  _ 2  _ 3  _ 4  _ 5  _ 6
			"Cycle_Next_Next_Value": {"3"},
		},
	))
}

func TestCyclicalTypesAdjacent(t *testing.T) { // issue #8
	type P struct {
		V    int
		L, R *P
	}

	runner := func(max int, want *P, env Env) func(*testing.T) {
		return func(t *testing.T) {
			dest := new(P)
			r := Reader{Source: env, Sep: "_", MaxDepth: max}
			err := r.Getenv(dest, "Cycle")
			if err != nil && !IsNoValue(err) {
				t.Fatalf("unexpected error: %v", err)
			}
			if !reflect.DeepEqual(dest, want) {
				// Encode as JSON just to make it easier to visualize
				jsondest, _ := json.Marshal(dest)
				jsonwant, _ := json.Marshal(want)
				t.Fatalf("got\t%s\nwant\t%s",
					jsondest,
					jsonwant,
				)
			}
		}
	}

	t.Run("ResetCounterOnStructMatch", runner(
		2, // max depth
		&P{
			R: &P{V: 2},
			L: &P{L: &P{V: 4}},
		},
		Values{ //     1 2 3 4 5 6 7 7
			"Cycle_R_V":   {"2"},
			"Cycle_L_L_V": {"4"},
		},
	))

	t.Run("NoMatchOutsideDepth", runner(
		2, // max depth
		&P{},
		Values{ //     1 2 3 4 5 6 7 7
			"Cycle_L_L_V": {"4"},
		},
	))
}
