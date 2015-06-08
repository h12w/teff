package core

import (
	"bufio"
	"errors"
	"io"
)

var errSyntax = errors.New("syntax error")

func Parse(reader io.Reader) (List, error) {
	s := newParseStack()
	scanner := NewScanner(bufio.NewReader(reader))
	for scanner.Scan() {
		tok := scanner.Token()
		switch tok.Type {
		case Value:
			s.top().add(Node{Value: tok.Content})
		case Indent:
			last := s.top().last()
			if last == nil {
				return nil, errSyntax
			}
			s.push(&last.List)
		case Unindent:
			s.pop()
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return *s.top(), nil
}

type parseStack struct {
	s []*List
}

func newParseStack() parseStack {
	return parseStack{[]*List{&List{}}}
}

func (s *parseStack) top() *List {
	return s.s[len(s.s)-1]
}

func (s *parseStack) push(l *List) {
	s.s = append(s.s, l)
}

func (s *parseStack) pop() {
	s.s = s.s[:len(s.s)-1]
}

func (l *List) last() *Node {
	if len(*l) == 0 {
		return nil
	}
	return &(*l)[len(*l)-1]
}

func (l *List) add(node Node) {
	*l = append(*l, node)
}
