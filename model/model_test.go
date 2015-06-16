package model

import (
	"fmt"
	"reflect"
	"testing"
)

/*
TODO:
1. mismatch type for struct field
2. ignore setting unexported field
3. reading unexported field
*/

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

		{
			[]int{},
			List{},
		},
		{
			[]string{"a"},
			List{{Value: "a"}},
		},
		{
			[]int{1, 2},
			List{{Value: 1}, {Value: 2}},
		},
		{
			[][]int{{1, 2}, {3}},
			List{
				{List: List{{Value: 1}, {Value: 2}}},
				{List: List{{Value: 3}}},
			},
		},

		{
			[]*int{pi(1), pi(2)},
			List{{Value: 1}, {Value: 2}},
		},
		{
			func() []*int {
				i := pi(3)
				return []*int{i, i}
			}(),
			List{{Label: Label("1"), Value: 3}, {Value: Label("1")}},
		},
		{
			struct{}{},
			List{},
		},
		{
			struct {
				I int
				S string
			}{1, "a"},
			List{
				{Value: IdentValue{"I", 1}},
				{Value: IdentValue{"S", "a"}},
			},
		},
		//{
		//	func() struct {
		//		I1 *int
		//		I2 *int
		//	} {
		//		i := 3
		//		return struct {
		//			I1 *int
		//			I2 *int
		//		}{&i, &i}
		//	}(),
		//	List{
		//		{Value: IdentValue{"I1", 3}, Label: "1"},
		//		{Value: Label("1")},
		//	},
		//},
	} {
		{
			list, err := New(testcase.v)
			if err != nil {
				t.Fatalf("testcase %d: New: %v", i, err)
			}
			if !reflect.DeepEqual(list, testcase.l) {
				t.Fatalf("testcase %d: New: mismatch, expect \n%v\ngot\n%v", i, testcase.l, list)
			}
		}
		{
			v := newValueOf(testcase.v)
			if err := Fill(testcase.l, v); err != nil {
				t.Fatalf("testcase %d: Fill: %v", i, err)
			}
			list, err := New(v)
			if err != nil {
				t.Fatalf("testcase %d: New: %v", i, err)
			}
			if !reflect.DeepEqual(list, testcase.l) {
				t.Fatalf("testcase %d: Fill: mismatch, expect \n%v\ngot\n%v", i, testcase.l, list)
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

func (n *Node) String() string {
	return fmt.Sprintf("%#v", *n)
}
