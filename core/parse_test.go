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
		}, `
a
    b
    c
        e
        f
    d
    `},

		{List{
			{"a", nil, []string{"a1"}},
		}, `
#a1
a
		`},

		{List{
			{"a", nil, []string{"a1", "a2"}},
		}, `
#a1
#a2
a
		`},

		{List{
			{"a", nil, []string{"a1", "a2"}},
			{"b", nil, []string{"b1", "b2"}},
		}, `
#a1
#a2
a
#b1
#b2
b
		`},
	} {
		if i != 6 {
			continue
		}
		testcase.s = strings.Trim(testcase.s, "\n")
		list, err := Parse(strings.NewReader(testcase.s))
		if err != nil {
			t.Fatalf("testcase %d, %v", i, err)
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
		`
    #a
a
`,
		`
a
    #b
b
`,
		`
a
#b
    b
`,
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
