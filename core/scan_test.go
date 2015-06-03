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
		{"\r", ""},
		{"\n", ""},
		{"\r\n", ""},
		{"\n\r", ""},
		{"\n\n", ""},
		{"x", "<x:s>"},
		{"x\ny", "<x:s> <y:s>"},
		{"#x", "<x:a>"},
		{"#x\n#y", "<x:a> <y:a>"},

		{"x\n\ty", "<x:s> <in> <y:s> <un>"},
	} {
		if i != 9 {
			continue
		}
		toks, err := scanAll(testcase.s)
		if err != nil {
			t.Fatalf("testcase %d: %v", i, err)
		}
		actual := strings.Join(toks, " ")
		if actual != testcase.expected {
			t.Fatalf("testcase %d: expect %s, got %s", i, testcase.expected, actual)
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
