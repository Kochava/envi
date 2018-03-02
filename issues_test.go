package envi

import (
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
