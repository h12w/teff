package model

import (
	"reflect"
	"testing"
)

func TestModel(t *testing.T) {
	for i, testcase := range []struct {
		v interface{}
		l List
	}{
		{nil, nil},

		{1, List{Node{Value: 1}}},
	} {
		list, err := New(testcase.v)
		if err != nil {
			t.Fatalf("testcase %d: New: %v", i, err)
		}
		if !reflect.DeepEqual(list, testcase.l) {
			t.Fatalf("testcase %d: New: mismatch", i)
		}

		newValue := newValueOf(testcase.v)
		err = testcase.l.Fill(newValue)
		if err != nil {
			t.Fatalf("testcase %d: Fill: %v", i, err)
		}
		if !reflect.DeepEqual(newValue, testcase.v) {
			t.Fatalf("testcase %d: Fill: mismatch", i)
		}
	}
}

func newValueOf(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	return reflect.New(reflect.TypeOf(v)).Interface()
}
