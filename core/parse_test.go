package core

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	for i, testcase := range []struct {
		v List
		s string
	}{
		{List{}, ""},
		{List{
			{"a", nil, nil},
		}, `
a
		`},
		{List{
			{"a", nil, nil},
			{"b", nil, nil},
		}, `
a
b
		`},
		{List{
			{"a", List{
				{"b", nil, nil},
				{"c", List{
					{"e", nil, nil},
					{"f", nil, nil},
				}, nil},
				{"d", nil, nil},
			}, nil},
		}, "a\n\tb\n\tc\n\t\te\n\t\tf\n\td"},
	} {
		testcase.s = strings.TrimSpace(testcase.s)
		list, err := Parse(strings.NewReader(testcase.s))
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(list, testcase.v) {
			t.Fatalf("testcase %d: expect \n%#v\nbut got \n%#v", i, testcase.v, list)
		}
	}
}

func TestParseError(t *testing.T) {
	for i, testcase := range []string{
		"\ta",
		"\x00",
	} {
		_, err := Parse(strings.NewReader(testcase))
		if err == nil {
			t.Fatalf("testcase %d, expect error but got nil", i)
		}
	}
}

type stringer interface {
	String() string
}
