package teff

import (
	"fmt"
	"reflect"
	"testing"
)

func TestMarshal(t *testing.T) {
	for i, testcase := range []struct {
		value interface{}
		text  string
	}{
		{nil, "nil"},

		{1, "1"},
		{-1, "-1"},

		{"a", `a`},
		{ns("a"), `a`},

		{[]int{}, ""},
		{[]int{1, 2, 3}, "1\n2\n3"},
		{[]string{"a", "b", "c"}, "a\nb\nc"},
		{[]*string{ns("a"), ns("b"), ns("c")}, "a\nb\nc"},
		{func() []*string {
			a := ns("a")
			return []*string{a, a}
		}(), "# ^1\na\n^1"},
	} {
		{
			buf, err := Marshal(testcase.value)
			if err != nil {
				t.Fatal(err)
			}
			result := string(buf)
			if result != testcase.text {
				t.Fatalf("expect \n%s\n    but got \n%s", testcase.text, result)
			}
		}

		{
			newValue := newValueOf(testcase.value)
			if err := Unmarshal([]byte(testcase.text), newValue); err != nil {
				t.Fatal(err)
			}
			buf, err := Marshal(newValue)
			if err != nil {
				t.Fatal(err)
			}
			result := string(buf)
			if result != testcase.text {
				t.Fatalf("testcase %d: expect \n%s\n    but got \n%s", i, testcase.text, result)
			}
		}
	}
}

func TestAlloc(t *testing.T) {
	{
		var p *int
		if err := Unmarshal([]byte("1"), &p); err != nil {
			t.Fatal(err)
		}
		if p == nil || *p != 1 {
			t.Fatalf("expect 1 but got %v", p)
		}
	}
	{
		var p **int
		if err := Unmarshal([]byte("2"), &p); err != nil {
			t.Fatal(err)
		}
		if p == nil || *p == nil || **p != 2 {
			t.Fatalf("expect 2 but got %v", p)
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

func ns(s string) *string {
	return &s
}
