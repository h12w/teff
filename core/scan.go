package core

import (
	"errors"
	"io"
	"unicode"
)

var (
	errInvalidCodePoint = errors.New("invalid code point")
)

type Scanner struct {
	r   io.RuneReader
	err error
}

func NewScanner(r io.RuneReader) *Scanner {
	return &Scanner{r: r}
}

func (s *Scanner) Scan() bool {
	r, _, err := s.r.ReadRune()
	if err != nil {
		s.err = err
		return false
	}
	switch r {
	case '\t', ' ':
	case '\r', '\n':
	case unicode.ReplacementChar:
		s.err = errInvalidCodePoint
		return false
	default:
		if '\x00' <= r && r <= '\x19' {
			s.err = errInvalidCodePoint
			return false
		}
	}
	return true
}

func (s *Scanner) Err() error {
	return s.err
}
