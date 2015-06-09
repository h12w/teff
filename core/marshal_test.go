package core

import (
	"strings"
	"testing"
)

//		testcase.s = strings.Trim(testcase.s, "\n")

func TestMarshal(t *testing.T) {
	for i, testcase := range typeTestCases {
		testcase.s = strings.Trim(testcase.s, "\n")
		if str := testcase.v.String(); str != testcase.s {
			p(testcase.s, []byte(testcase.s))
			t.Fatalf("testcase %d: expect \n%s\nbut got \n%s.", i, testcase.s, str)
		}
	}
}
