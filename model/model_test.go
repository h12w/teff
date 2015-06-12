package model

import (
	"fmt"
	"reflect"
	"testing"
)

func TestModel(t *testing.T) {
	for i, testcase := range []struct {
		v interface{}
		l List
	}{
		{nil, nil},

		{1, List{{Value: 1}}},
		{pi(1), List{{Value: 1}}},
		{"a", List{{Value: "a"}}},
		{ps("a"), List{{Value: "a"}}},

		{[]int{}, List{}},
		{[]string{"a"}, List{{Value: "a"}}},
		{[]int{1, 2}, List{{Value: 1}, {Value: 2}}},
		{
			[][]int{{1, 2}, {3}},
			List{
				{List: List{{Value: 1}, {Value: 2}}},
				{List: List{{Value: 3}}},
			},
		},

		{[]*int{pi(1), pi(2)}, List{{Value: 1}, {Value: 2}}},
	} {
		{
			list, err := New(testcase.v)
			if err != nil {
				t.Fatalf("testcase %d: New: %v", i, err)
			}
			if !reflect.DeepEqual(list, testcase.l) {
				t.Fatalf("testcase %d: New: mismatch, expect \n%#v\ngot\n%#v", i, testcase.l, list)
			}
		}

		{
			v := newValueOf(testcase.v)
			if err := testcase.l.Fill(v); err != nil {
				t.Fatalf("testcase %d: Fill: %v", i, err)
			}
			list, err := New(v)
			if err != nil {
				t.Fatalf("testcase %d: New: %v", i, err)
			}
			if !reflect.DeepEqual(list, testcase.l) {
				t.Fatalf("testcase %d: Fill: mismatch, expect \n%#v\ngot\n%#v", i, testcase.l, list)
			}
		}
	}
}

func newValueOf(v interface{}) interface{} {
	if v == nil {
		return nil
	}
	return reflect.New(reflect.TypeOf(v)).Interface()
}

var p = fmt.Println

func pi(i int) *int {
	return &i
}

func ps(s string) *string {
	return &s
}
