package core

import (
	"errors"
	"io"
	"strings"
	"unicode"
)

var (
	errInvalidCodePoint = errors.New("invalid code point")
)

type TokenType int

const (
	Invalid TokenType = iota
	Annotation
	LineString
	Indent
	Unindent
)

type Token struct {
	Type  TokenType
	Value string
}

type Scanner struct {
	r       io.RuneScanner
	ch      rune
	indents []string
	indent  string
	tok     Token
	toks    []Token
	err     error
}

func NewScanner(r io.RuneScanner) *Scanner {
	return &Scanner{r: r, indents: []string{""}, indent: ""}
}

func (s *Scanner) Scan() bool {
	if s.err != nil {
		return false
	}
	if len(s.toks) > 0 {
		s.tok, s.toks = s.toks[0], s.toks[1:]
		return true
	}
	s.indent = string(s.indentSpaces())
	if s.err != nil {
		return false
	}
	if s.indent == s.lastIndent() {
		return s.afterIndent()
	} else if strings.HasPrefix(s.indent, s.lastIndent()) {
		s.indents = append(s.indents, s.indent)
		s.tok = Token{Type: Indent}
		return true
	} else if len(s.indent) >= len(s.lastIndent()) {
		s.err = errors.New("mismatch indent")
		return false
	}
	for i := len(s.indents) - 1; i >= 0; i-- {
		if s.indent == s.indents[i] {
			s.tok = Token{Type: Unindent}
			return true
		}
		s.toks = append(s.toks, Token{Type: Unindent})
	}
	s.err = errors.New("mismatch indent")
	return false
}
func (s *Scanner) lastIndent() string {
	return s.indents[len(s.indents)-1]
}

func (s *Scanner) afterIndent() bool {
	if !s.next() {
		return false
	}
	switch s.ch {
	case '#':
		return s.annotation()
	case '\r', '\n':
		return s.skipNewline()
	default:
		return s.lineString()
	}
	return false
}

func (s *Scanner) skipNewline() bool {
	for s.next() {
		if s.ch != '\r' && s.ch != '\n' {
			s.prev()
			break
		}
	}
	return s.Scan()
}

func (s *Scanner) annotation() bool {
	s.tok = Token{Type: Annotation, Value: s.inline()[1:]}
	return true
}

func (s *Scanner) inline() string {
	rs := []rune{s.ch}
	for s.next() {
		if s.ch == '\r' || s.ch == '\n' {
			s.prev()
			break
		}
		rs = append(rs, s.ch)
	}
	return string(rs)
}

func (s *Scanner) lineString() bool {
	s.tok = Token{Type: LineString, Value: s.inline()}
	return true
}

func (s *Scanner) indentSpaces() string {
	rs := []rune{}
	for s.next() {
		if s.ch != ' ' && s.ch != '\t' {
			s.prev()
			break
		}
		rs = append(rs, s.ch)
	}
	return string(rs)
}

func (s *Scanner) Token() Token {
	return s.tok
}

func (s *Scanner) next() bool {
	s.ch, _, s.err = s.r.ReadRune()
	if s.err != nil {
		return false
	}
	switch s.ch {
	case '\t', ' ', '\r', '\n':
	case unicode.ReplacementChar:
		s.err = errInvalidCodePoint
		return false
	default:
		if '\x00' <= s.ch && s.ch <= '\x19' {
			s.err = errInvalidCodePoint
			return false
		}
	}
	return true
}

func (s *Scanner) prev() bool {
	s.err = s.r.UnreadRune()
	return s.err == nil
}

func (s *Scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}
