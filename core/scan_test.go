package core

import (
	"bufio"
	"errors"
	"fmt"
	"strings"
	"testing"
)

var p = fmt.Println

func (t TokenType) String() string {
	switch t {
	case Annotation:
		return "a"
	case LineString:
		return "s"
	case Indent:
		return "in"
	case Unindent:
		return "un"
	case EOF:
		return "eof"
	}
	return "?"
}

func (t Token) String() string {
	if t.Value == "" {
		return fmt.Sprintf("<%s>", t.Type.String())
	}
	return fmt.Sprintf("<%s:%s>", t.Value, t.Type.String())
}

func TestScan(t *testing.T) {
	for i, testcase := range []struct {
		s        string
		expected string
	}{
		{"\r", "<eof>"},
		{"\n", "<eof>"},
		{"\r\n", "<eof>"},
		{"\n\r", "<eof>"},
		{"\n\n", "<eof>"},
		{"x", "<x:s> <eof>"},
		{"x\ny", "<x:s> <y:s> <eof>"},
		{"#x", "<x:a> <eof>"},
		{"#x\n#y", "<x:a> <y:a> <eof>"},

		{"x\n\ty", "<x:s> <in> <y:s> <un> <eof>"},
		{"x\n\ty\n", "<x:s> <in> <y:s> <un> <eof>"},
		{"x\n\ty\nz", "<x:s> <in> <y:s> <un> <z:s> <eof>"},
		//{"x\n\ty\n\t\tz", "<x:s> <in> <y:s> <un> <eof>"},
	} {
		toks, err := scanAll(testcase.s)
		if err != nil {
			t.Fatalf("testcase %d: %v", i, err)
		}
		actual := strings.Join(toks, " ")
		if actual != testcase.expected {
			t.Fatalf("testcase %d: expect\n%s\ngot\n%s\n", i, testcase.expected, actual)
		}
	}
}

func TestInvalidChar(t *testing.T) {
	for i, testcase := range []string{
		"\x00",
		"\x19",
		"\xed\xa0",
	} {
		s := NewScanner(bufio.NewReader(strings.NewReader(testcase)))
		if s.Scan() != false || s.Err() == nil {
			t.Fatalf("testcase %d: expect error for illegal character.", i)
		}
	}
}

func TestReadError(t *testing.T) {
	s := NewScanner(errRuneReader{})
	if s.Scan() != false || s.Err() == nil {
		t.Fatal("expect read error.")
	}
}
func (errRuneReader) ReadRune() (rune, int, error) {
	return 0, 0, errors.New("any error")
}
func (errRuneReader) UnreadRune() error { return nil }

type errRuneReader struct{}

func scanAll(testcase string) (toks []string, err error) {
	s := NewScanner(bufio.NewReader(strings.NewReader(testcase)))
	for s.Scan() {
		toks = append(toks, s.Token().String())
	}
	if s.Err() != nil {
		return nil, s.Err()
	}
	return
}
