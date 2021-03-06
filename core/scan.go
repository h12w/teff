package core

import (
	"errors"
	"io"
	"strings"
	"unicode"
)

var (
	errInvalidCodePoint = errors.New("invalid code point")
	errMismatchIndent   = errors.New("mismatch indent")
)

type TokenType int

const (
	Invalid TokenType = iota
	Annotation
	Reference
	LineValue
	Indent
	Unindent
	EOF
	_SOF
)

type Token struct {
	Type    TokenType
	Content string
}

type Scanner struct {
	reader
	indenter
	tokenQueue
	err error
}

func NewScanner(r io.RuneScanner) *Scanner {
	return &Scanner{
		reader: reader{r: r},
		indenter: indenter{
			indents: []string{""},
		},
		tokenQueue: tokenQueue{
			toks: []Token{Token{Type: _SOF}},
		},
	}
}

func (s *Scanner) Scan() bool {
	s.popTok()
	if s.tokCount() > 0 {
		return true
	}
	if s.err != nil {
		return false
	}
	s.scanLine()
	if s.err == io.EOF {
		for i := 0; i < s.eofUnindentLevel(); i++ {
			s.pushTok(Token{Type: Unindent})
		}
		s.pushTok(Token{Type: EOF})
	}
	return s.tokCount() > 0
}

func (s *Scanner) scanLine() {
	var indent string
	indent, s.err = s.readValidIndent()
	if s.err != nil {
		return
	}
	indentType, n, err := s.indentLevel(indent)
	if err != nil {
		s.err = err
		return
	}
	for i := 0; i < n; i++ {
		s.pushTok(Token{Type: indentType})
	}
	var line string
	line, s.err = s.readLine()
	switch line[0] {
	case '#':
		s.pushTok(Token{Type: Annotation, Content: line[1:]})
	case '^':
		s.pushTok(Token{Type: Reference, Content: line[1:]})
	default:
		s.pushTok(Token{Type: LineValue, Content: line})
	}
}

func (s *Scanner) Err() error {
	if s.err == io.EOF {
		return nil
	}
	return s.err
}

type reader struct {
	r   io.RuneScanner
	ch  rune
	err error
}

func (s *reader) readLine() (string, error) {
	rs := []rune{}
	for s.next() {
		switch s.ch {
		case '\r', '\n':
			s.prev()
			return string(rs), nil
		}
		rs = append(rs, s.ch)
	}
	return string(rs), s.err
}

// readValidIndent reads an indent that not ends with line breaks, and skips
// a line when necessary.
func (s *reader) readValidIndent() (string, error) {
	for {
		indent := s.indentSpaces()
		if s.err != nil {
			return indent, s.err
		}
		if !s.skipLineBreaks() {
			return indent, s.err
		}
	}
}
func (s *reader) skipLineBreaks() (isLineBreak bool) {
	for s.next() {
		switch s.ch {
		case '\r', '\n':
			isLineBreak = true
		default:
			s.prev()
			return
		}
	}
	return
}
func (s *reader) indentSpaces() string {
	rs := []rune{}
	for s.next() {
		switch s.ch {
		case ' ', '\t':
		default:
			s.prev()
			return string(rs)
		}
		rs = append(rs, s.ch)
	}
	return string(rs)
}

func (s *reader) next() bool {
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

func (s *reader) prev() bool {
	s.err = s.r.UnreadRune()
	return s.err == nil
}

type indenter struct {
	indents []string
}

func (s *indenter) indentLevel(indent string) (TokenType, int, error) {
	top := s.indents[len(s.indents)-1]
	if indent == top {
		return 0, 0, nil
	} else if strings.HasPrefix(indent, top) {
		s.indents = append(s.indents, indent)
		return Indent, 1, nil
	}
	for i := 1; i < len(s.indents); i++ {
		if indent == s.indents[len(s.indents)-i-1] {
			s.indents = s.indents[:len(s.indents)-i]
			return Unindent, i, nil
		}
	}
	return 0, 0, errMismatchIndent
}

func (s *indenter) eofUnindentLevel() int {
	if len(s.indents) > 1 {
		n := len(s.indents) - 1
		s.indents = s.indents[:1]
		return n
	}
	return 0
}

type tokenQueue struct {
	toks []Token
}

func (s *tokenQueue) Token() Token {
	return s.toks[0]
}

func (s *tokenQueue) pushTok(tok Token) {
	s.toks = append(s.toks, tok)
}

func (s *tokenQueue) popTok() {
	s.toks = s.toks[1:]
}

func (s *tokenQueue) tokCount() int {
	return len(s.toks)
}
