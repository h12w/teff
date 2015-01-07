package tff

import "testing"

func TestNode(t *testing.T) {
	for i, testcase := range []struct {
		a stringer
		b string
	}{
		{List{}, ""},
		{&Node{}, ""},
		{List{
			{"a", nil},
			{"b", nil},
		}, "a\nb"},
		{&Node{"a", List{
			{"b", nil},
			{"c", List{
				{"e", nil},
				{"f", nil},
			}},
			{"d", nil},
		}},
			"a\n\tb\n\tc\n\t\te\n\t\tf\n\td",
		},
	} {
		if str := testcase.a.String(); str != testcase.b {
			t.Errorf("testcase %d: expect %s but got %s.", i, testcase.b, str)
		}
	}
}

type stringer interface {
	String() string
}
