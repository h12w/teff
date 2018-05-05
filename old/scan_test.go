// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tff

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"h12.io/gombi/experiment/gre/scan"
	"h12.io/gspec"
)

const EOF = 0

type testCase struct {
	text   string
	tokens []*scan.Token
}

var testCases = []testCase{
	{
		"a",
		[]*scan.Token{
			{tokenString, []byte("a"), 0},
		},
	},
	{
		"a b",
		[]*scan.Token{
			{tokenString, []byte("a"), 0},
			{tokenString, []byte("b"), 2},
		},
	},
	{
		"{}",
		[]*scan.Token{
			{tokenLeftBrace, []byte("{"), 0},
			{tokenRightBrace, []byte("}"), 1},
		},
	},
	{
		"{a}",
		[]*scan.Token{
			{tokenLeftBrace, []byte("{"), 0},
			{tokenString, []byte("a"), 1},
			{tokenRightBrace, []byte("}"), 2},
		},
	},
	{
		"{a, b}",
		[]*scan.Token{
			{tokenLeftBrace, []byte("{"), 0},
			{tokenString, []byte("a"), 1},
			{tokenComma, []byte(","), 2},
			{tokenString, []byte("b"), 4},
			{tokenRightBrace, []byte("}"), 5},
		},
	},
	{
		"{a,}",
		[]*scan.Token{
			{tokenLeftBrace, []byte("{"), 0},
			{tokenString, []byte("a"), 1},
			{tokenComma, []byte(","), 2},
			{tokenRightBrace, []byte("}"), 3},
		},
	},
	{
		"a{b}",
		[]*scan.Token{
			{tokenString, []byte("a"), 0},
			{tokenLeftBrace, []byte("{"), 1},
			{tokenString, []byte("b"), 2},
			{tokenRightBrace, []byte("}"), 3},
		},
	},
	{
		`"a"`,
		[]*scan.Token{
			{tokenString, []byte(`"a"`), 0},
		},
	},
	/*
		{
			"a:",
			[]*scan.Token{
				{tokenString, 0, []byte("a")},
				{tokenString, 1, []byte(":")},
			},
		},
		{
			`"a":`,
			[]*scan.Token{
				{tokenString, 0, []byte(`"a"`)},
				{tokenString, 3, []byte(`:`)},
			},
		},
		{
			`{a}:`,
			[]*scan.Token{
				{tokenLeftBrace, 0, []byte("{")},
				{tokenString, 1, []byte("a")},
				{tokenRightBrace, 2, []byte("}")},
				{tokenString, 3, []byte(":")},
			},
		},
	*/
	{
		"/usr/bin",
		[]*scan.Token{
			{tokenString, []byte("/usr/bin"), 0},
		},
	},
	{
		"a //0123",
		[]*scan.Token{
			{tokenString, []byte("a"), 0},
			{tokenComment, []byte("//0123"), 2},
		},
	},
}

var _ = gspec.Add(func(s gspec.S) {
	describe, testcase := s.Alias("describe"), s.Alias("testcase")
	expect := gspec.Expect(s.Fail)

	describe("tff.scanner", func() {
		for _, tc := range testCases {
			testcase(fmt.Sprint(tc), func() {
				s := newScanner(strings.NewReader(tc.text))
				tokens := s.scanAll()
				expect(len(tokens)).NotEqual(0)
				tokens, eof := tokens[:len(tokens)-1], tokens[len(tokens)-1]
				expect(tokens).Equal(tc.tokens)
				expect(eof.ID).Equal(EOF)
			})
		}
	})
})

func (s *scanner) scanAll() (tokens []*scan.Token) {
	for s.Scan() {
		tokens = append(tokens, s.Token())
	}
	return
}

func TestAll(t *testing.T) {
	//	for i := 0; i < 500; i++ {
	gspec.Test(t)
	//	}
}

func p(v ...interface{}) {
	fmt.Println(v...)
}

func init() {
	gspec.SetSprint(ogdlPrint)
}

func ogdlPrint(v interface{}) string {
	buf, _ := MarshalIndent(v, "    ", "    ")
	typ := ""
	if v != nil {
		typ = reflect.TypeOf(v).String() + "\n"
	}
	return "\n" +
		typ +
		string(buf) +
		"\n"
}

func dumpPrint(v interface{}) string {
	spew.Config.Indent = "    "
	return "\n" + spew.Sdump(v)
}

func jsonPrint(v interface{}) string {
	buf, _ := json.MarshalIndent(v, "    ", "  ")
	return "\n    " + string(buf) + "\n"
}
