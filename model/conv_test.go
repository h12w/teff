package model

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

/*
TODO:
1. mismatch type for struct field
2. ignore setting unexported field
3. reading unexported field
4. type S []S
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
			List{{RefID: RefID("1"), Value: 3}, {Value: RefID("1")}},
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
				{Value: Identifier("I"), List: List{{Value: 1}}},
				{Value: Identifier("S"), List: List{{Value: "a"}}},
			},
		},
		{
			func() struct {
				S1 *string
				S2 *string
			} {
				s := "a"
				return struct {
					S1 *string
					S2 *string
				}{&s, &s}
			}(),
			List{
				{Value: Identifier("S1"), List: List{{RefID: "1", Value: "a"}}},
				{Value: Identifier("S2"), List: List{{Value: RefID("1")}}},
			},
		},

		{
			func() *struct { // return pointer so that S1 is addressable and can be correctly referenced.
				S1 string
				S2 *string
				S3 **string
				S4 ***string
			} {
				s := struct {
					S1 string
					S2 *string
					S3 **string
					S4 ***string
				}{S1: "a"}
				s.S2 = &s.S1
				s.S3 = &s.S2
				s.S4 = &s.S3
				return &s
			}(),
			List{
				// RefID should imitate exactly like the original pointer so the data topology can be reconstructed
				{Value: Identifier("S1"), List: List{{RefID: "1", Value: "a"}}},
				{Value: Identifier("S2"), List: List{{RefID: "2", Value: RefID("1")}}},
				{Value: Identifier("S3"), List: List{{RefID: "3", Value: RefID("2")}}},
				{Value: Identifier("S4"), List: List{{Value: RefID("3")}}},
			},
		},

		{
			func() *struct {
				S1 string
				S2 *string
				S3 **string
				S4 ***string
			} {
				s := struct {
					S1 string
					S2 *string
					S3 **string
					S4 ***string
				}{S1: "a"}
				b := "b"
				s.S2 = &b
				s.S3 = &s.S2
				s.S4 = &s.S3
				return &s
			}(),
			List{
				{Value: Identifier("S1"), List: List{{Value: "a"}}},
				{Value: Identifier("S2"), List: List{{RefID: "1", Value: "b"}}},
				{Value: Identifier("S3"), List: List{{RefID: "2", Value: RefID("1")}}},
				{Value: Identifier("S4"), List: List{{Value: RefID("2")}}},
			},
		},

		{
			func() *struct {
				S1 string
				S3 **string
				S4 ***string
				S5 ****string
			} {
				s := struct {
					S1 string
					S3 **string
					S4 ***string
					S5 ****string
				}{S1: "a"}
				s2 := &s.S1
				s.S3 = &s2
				s.S4 = &s.S3
				s.S5 = &s.S4
				return &s
			}(),
			List{
				{Value: Identifier("S1"), List: List{{RefID: "1", Value: "a"}}},
				{Value: Identifier("S3"), List: List{{RefID: "2", Value: RefID("1")}}},
				{Value: Identifier("S4"), List: List{{RefID: "3", Value: RefID("2")}}},
				{Value: Identifier("S5"), List: List{{Value: RefID("3")}}},
			},
		},

		//{
		//	func() *struct { // reverse reference
		//		S2 *string
		//		S1 string
		//	} {
		//		s := struct {
		//			S2 *string
		//			S1 string
		//		}{S1: "a"}
		//		s.S2 = &s.S1
		//		return &s
		//	}(),
		//	List{
		//		{Value: Identifier("S2"), List: List{{Value: RefID("1")}}},
		//		{Value: Identifier("S1"), List: List{{RefID: "1", Value: "a"}}},
		//	},
		//},
	} {
		{
			list, err := New(testcase.v)
			if err != nil {
				t.Fatalf("testcase %d: New: %v", i, err)
			}
			if !reflect.DeepEqual(list, testcase.l) {
				t.Fatalf("testcase %d: New: mismatch, expect \n%v\ngot\n%v", i, testcase.l.String(), list.String())
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
				t.Fatalf("testcase %d: Fill: mismatch, expect \n%v\ngot\n%v", i, testcase.l.String(), list.String())
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
	ref := string(n.RefID)
	if ref != "" {
		ref = "^" + ref
	}
	switch len(n.List) {
	case 0:
		return fmt.Sprintf("%v%s", n.Value, ref)
	case 1:
		return fmt.Sprintf("%v%s: %s", n.Value, ref, n.List[0].String())
	default:
		return fmt.Sprintf("%v%s: %s", n.Value, ref, n.List.String())
	}
}

func (l *List) String() string {
	ss := make([]string, len(*l))
	for i := range ss {
		ss[i] = (*l)[i].String()
	}
	return "{" + strings.Join(ss, ", ") + "}"
}
