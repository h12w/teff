package core

/*
import (
	"reflect"
	"strings"
	"testing"
)

func TestCore(t *testing.T) {
	for _, testcase := range []struct {
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
		if !reflect.DeepEqual(*list, testcase.v) {
			t.Fatalf("expect %#v, but got %#v", testcase.v, *list)
		}
	}
}
*/

type stringer interface {
	String() string
}
