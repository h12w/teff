// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tff

import (
	"io"

	"h12.io/gombi/experiment/gre/scan"
)

const (
	tokenEOF = iota
	tokenComment
	tokenLeftBrace
	tokenRightBrace
	tokenComma
	tokenString
	tokenSpace
)

type tokenType int

func (t tokenType) String() string {
	switch t {
	case tokenEOF:
		return "tokenEOF"
	case tokenComment:
		return "tokenComment"
	case tokenString:
		return "tokenString"
	case tokenLeftBrace:
		return "tokenLeftBrace"
	case tokenRightBrace:
		return "tokenRightBrace"
	case tokenComma:
		return "tokenComma"
	}
	return "token unkown"
}

type scanner struct {
	scan.Scanner
}

func (s *scanner) Scan() bool {
	for s.Scanner.Scan() {
		if s.Token().ID != tokenSpace {
			return true
		}
	}
	return false
}

func newScanner(r io.Reader) *scanner {
	var (
		char  = scan.Char
		pat   = scan.Pat
		merge = scan.Merge
		or    = scan.Or
		con   = scan.Con

		nonctrl   = char(`[:cntrl:]`).Negate()
		indent    = char(`\t `)
		lineBreak = char(`\n\r`)
		space     = merge(indent, lineBreak)
		any       = merge(nonctrl, space)
		inline    = any.Exclude(lineBreak)
		delim     = char(`,{}`)
		empty     = pat(``)
		//invalid   = any.Negate()

		newline        = or(lineBreak, pat(`\r\n`))
		inlineComment  = con(pat(`//`), inline.ZeroOrMore(), or(newline, empty))
		quoted         = or(inline.Exclude(char(`"`)), pat(`\\"`))
		quotedString   = con(pat(`"`), quoted.ZeroOrMore(), pat(`"`))
		unquoted       = any.Exclude(delim, space)
		unquotedString = unquoted.OneOrMore()
		generalString  = or(quotedString, unquotedString)

		matcher = scan.NewMatcher(
			inlineComment,
			char(`{`),
			char(`}`),
			char(`,`),
			generalString,
			space.OneOrMore(),
		)
	)
	s := scan.Scanner{Matcher: matcher}
	err := s.SetReader(r)
	if err != nil {
		panic(err)
	}
	return &scanner{s}
}
