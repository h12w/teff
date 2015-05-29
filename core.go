package teff

import (
	"bufio"
	"errors"
	"io"
	"unicode"
)

var (
	errIllegalChar = errors.New("illegal Unicode character")
	errInvalidChar = errors.New("invalid character in TEFF")
)

type (
	Node struct {
		Value       string
		List        List
		Annotations []string
	}
	List []Node
)

func (n Node) String() string {
	return ""
}

func (n List) String() string {
	return ""
}

func ParseCore(reader io.Reader) (*List, error) {
	return newParser(reader).Parse()
}

type parser struct {
	rd  *bufio.Reader
	ch  rune
	err error
}

func newParser(r io.Reader) *parser {
	return &parser{rd: bufio.NewReader(r)}
}

func (pa *parser) Parse() (*List, error) {
	for pa.next() {
		switch pa.ch {
		case ' ', '\t':
		case '\r', '\n':
		case '#':
		default:
		}
	}
	if pa.err != nil && pa.err != io.EOF {
		return nil, pa.err
	}
	return &List{}, nil
}

func (pa *parser) next() bool {
	if pa.err != nil {
		return false
	}
	r, _, err := pa.rd.ReadRune()
	if err == io.EOF {
		pa.err = io.EOF
	} else if err != nil {
		pa.err = pa.formatErr(err)
	} else if r == unicode.ReplacementChar {
		pa.err = pa.formatErr(errIllegalChar)
	} else if r < 0x20 && r != '\t' && r != '\r' && r != '\n' {
		pa.err = pa.formatErr(errInvalidChar)
	} else {
		pa.ch = r
		return true
	}
	return false
}

// TODO: at line position
func (pa *parser) formatErr(e error) error {
	return e
}
