package core

import (
	"reflect"
	"strings"
	"testing"
)

func TestParse(t *testing.T) {
	for i, testcase := range typeTestCases {
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
