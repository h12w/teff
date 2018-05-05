// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tff

import (
	"bytes"
	"reflect"
	"strconv"
	"time"

	"h12.io/gspec"
)

func init() {
	Register(INT(0))
}

type cyclicStruct struct {
	P *cyclicStruct
}

type structKey struct {
	IKey int
	SKey string
}

func newValue(v interface{}) reflect.Value {
	t := reflect.TypeOf(v)
	if t == nil {
		return reflect.ValueOf(nil)
	}
	return reflect.New(t)
}

type encodingTestCase struct {
	value interface{}
	text  string
}

type encodingTestGroup struct {
	typ   string
	cases []encodingTestCase
}

type encodingTestGroups []encodingTestGroup

func (tgs encodingTestGroups) Test(desc string, s gspec.S, visit func(tc encodingTestCase)) {
	testgroup, testcase := s.Alias(desc), s.Alias("testcase")
	for _, tg := range tgs {
		testgroup(tg.typ, func() {
			for _, tc := range tg.cases {
				typ := reflect.TypeOf(tc.value)
				typStr := ""
				if typ != nil {
					typStr = "(" + typ.String() + ") "
				}
				testcase(typStr+strconv.Quote(tc.text), func() {
					visit(tc)
				})
			}
		})
	}

}

var _encodingTestGroups = encodingTestGroups{
	{"untyped literals",
		[]encodingTestCase{
			{nil, "nil"},

			{true, "true"},
			{false, "false"},

			{0, "0"},
			{42, "42"},
			{-42, "-42"},

			{0.0, "0"},
			{3.1415, "3.1415"},
			{-3.1415, "-3.1415"},

			{1.2 + 3.4i, "1.2+3.4i"},
			{5.6i, "0+5.6i"},
			{-7.8i, "0-7.8i"},
		},
	},

	{"integer literals",
		[]encodingTestCase{
			{int8(1), "1"},
			{int16(2), "2"},
			{int32(3), "3"},
			{int64(4), "4"},
			{uint8(5), "5"},
			{uint16(6), "6"},
			{uint32(7), "7"},
			{uint64(8), "8"},
			{int(9), "9"},
			{uint(10), "10"},
			{uintptr(11), "11"},
			{byte(12), "12"},
			{rune(13), "13"},
		},
	},

	{"float literals",
		[]encodingTestCase{
			{float32(1.234), "1.234"},
			{float64(5.678), "5.678"},
		},
	},

	{"complex literals",
		[]encodingTestCase{
			{complex64(-2.3 + 4.5i), "-2.3+4.5i"},
			{complex128(4.5 - 6.7i), "4.5-6.7i"},
			{complex64(4.5i), "0+4.5i"},
			{complex128(4.5), "4.5+0i"},
		},
	},

	{"string literals",
		[]encodingTestCase{
			{"a", `"a"`},
			{"\n", `"\n"`},
		},
	},

	{"slice",
		[]encodingTestCase{
			{[]int(nil), "nil"},
			{[]int{1}, "1"},
			{[]int{1, 2}, "1\n2"},
		},
	},

	{"array",
		[]encodingTestCase{
			{[...]int{1}, "1"},
			{[...]int{1, 2}, "1\n2"},
			//{[...][2]int{{1, 2}, {3, 4}}, "_\n\t1\n\t2_\n\t3\n\t4"},
		},
	},

	{"struct",
		[]encodingTestCase{
			{struct{}{}, ""},
			{struct{ IVal int }{IVal: 1}, "IVal\n\t1"},
			{struct {
				IntVal    int
				StringVal string
			}{IntVal: 1, StringVal: "a"}, "IntVal\n\t1\nStringVal\n\t\"a\""},
		},
	},

	{"map",
		[]encodingTestCase{
			{map[string]bool(nil), "nil"},
			{make(map[string]bool), ""},
			{map[string]bool{"a": true, "b": false}, "\"a\"\n\ttrue\n\"b\"\n\tfalse"},
			{map[int]bool{1: true, 2: false}, "1\n\ttrue\n2\n\tfalse"},
			{map[structKey]bool{structKey{1, "a"}: true, structKey{2, "b"}: false},
				"{1, \"a\"}\n\ttrue\n{2, \"b\"}\n\tfalse"},
		},
	},

	{"pointer",
		[]encodingTestCase{
			{func() *int {
				i := 1
				return &i
			}(), "1"},
			{func() **int {
				i := 2
				pi := &i
				return &pi
			}(), "2"},
			{&struct{ I int }{42}, "I\n\t42"},
		},
	},

	{"TextMarshaler/TextUnmarshaler",
		[]encodingTestCase{
			{func() *time.Time {
				t := time.Date(2014, 5, 27, 20, 40, 11, 99, time.UTC)
				return &t
			}(), `2014-05-27T20:40:11.000000099Z`},
		},
	},

	{"cyclic references",
		[]encodingTestCase{
			{func() *cyclicStruct {
				a := &cyclicStruct{}
				a.P = a
				return a
			}(), "^1\n\tP\n\t\t^1"},
			//{func() *cyclicStruct {
			//	a := &cyclicStruct{}
			//	b := &cyclicStruct{}
			//	a.P = b
			//	b.P = a
			//	return a
			//}(), "^1\n\tP\n\t\tP\n\t\t\t^1"},
			{func() *struct{ I, J, K *int } {
				i := 42
				return &struct{ I, J, K *int }{&i, &i, &i}
			}(), "I\n\t^1\n\t\t42\nJ\n\t^1\nK\n\t^1"},
		},
	},

	{"interface",
		[]encodingTestCase{
			{INT(1), "!INT\n\t1"},
			{struct{ I interface{} }{1}, "I\n\t!int\n\t\t1"},
		},
	},
}

type INT int

type structType struct {
	IVal int
}

var marshalIndentTestGroups = encodingTestGroups{
	{"slice",
		[]encodingTestCase{
			{[]int(nil), "nil"},
			{[]int{1, 2}, " 1\n 2"},
		},
	},

	{"array",
		[]encodingTestCase{
			{[...]int{1, 2}, " 1\n 2"},
		},
	},

	{"struct",
		[]encodingTestCase{
			{struct{}{}, ""},
			{struct {
				IVal int
				SVal string
			}{IVal: 1, SVal: "a"}, " IVal\n \t1\n SVal\n \t\"a\""},
		},
	},

	{"map",
		[]encodingTestCase{
			{map[string]bool(nil), "nil"},
			{make(map[string]bool), ""},
			{map[int]bool{1: true, 2: false}, " 1\n \ttrue\n 2\n \tfalse"},
			{
				map[structKey]bool{
					structKey{1, "a"}: true,
					structKey{2, "b"}: false,
				},
				" {1, \"a\"}\n \ttrue\n {2, \"b\"}\n \tfalse",
			},
		},
	},
}

var _ = gspec.Add(func(s gspec.S) {
	describe := s.Alias("describe")
	expect := gspec.Expect(s.Fail)

	describe("Encoder", func() {
		var buf bytes.Buffer
		enc := NewEncoder(&buf)
		_encodingTestGroups.Test("encoding", s, func(tc encodingTestCase) {
			err := enc.Encode(tc.value)
			expect(err).Equal(nil)
			expect(buf.String()).Equal(tc.text)
		})
		marshalIndentTestGroups.Test("encoding and indenting", s, func(tc encodingTestCase) {
			r, err := MarshalIndent(tc.value, " ", "\t")
			expect(err).Equal(nil)
			expect(string(r)).Equal(tc.text)
		})
	})

	//describe("Decoder", func() {
	//	_encodingTestGroups.Test("decoding", s, func(tc encodingTestCase) {
	//		dec := NewDecoder(strings.NewReader(tc.text))
	//		nv := newValue(tc.value)
	//		if nv.IsValid() {
	//			err := dec.Decode(nv.Interface())
	//			expect(err).Equal(nil)
	//			expect(nv.Elem().Interface()).Equal(tc.value)
	//		}
	//	})
	//})
})
