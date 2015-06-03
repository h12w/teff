package core

import (
	"errors"
	"io"
	"unicode"
)

var (
	errInvalidCodePoint = errors.New("invalid code point")
)

type TokenType int

const (
	Invalid TokenType = iota
	Newline
	Annotation
	LineString
	Indent
	Unindent
)

type Token struct {
	Type  TokenType
	Value []byte
}

type Scanner struct {
	r   io.RuneScanner
	ch  rune
	tok Token
	err error
}

func NewScanner(r io.RuneScanner) *Scanner {
	return &Scanner{r: r}
}

func (s *Scanner) Scan() bool {
	if !s.next() {
		return false
	}
	switch s.ch {
	case '\t', ' ':
	case '\r', '\n':
		return s.newline()
	}
	return true
}

func (s *Scanner) newline() bool {
	s.tok = Token{Type: Newline}
	if s.ch == '\r' {
		if s.next() && s.ch != '\n' {
			return s.prev()
		}
	}
	return true
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
