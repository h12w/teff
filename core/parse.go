package core

import (
	"bufio"
	"errors"
	"io"
)

var errSyntax = errors.New("syntax error")

func Parse(reader io.Reader) (List, error) {
	var list List
	s := []*List{&list}
	top := s[len(s)-1]
	scanner := NewScanner(bufio.NewReader(reader))
	for scanner.Scan() {
		tok := scanner.Token()
		switch tok.Type {
		case Value:
			*top = append(*top, Node{Value: tok.Content})
		case Indent:
			if len(*top) == 0 {
				return nil, errSyntax
			}
			s = append(s, &(*top)[len(*top)-1].List)
			top = s[len(s)-1]
		case Unindent:
			s = s[:len(s)-1]
			top = s[len(s)-1]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return *top, nil
}
