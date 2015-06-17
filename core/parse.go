package core

import (
	"bufio"
	"errors"
	"io"
)

var (
	errWrongIndent           = errors.New("syntax error, wrong indent")
	errAnnotationWithoutNode = errors.New("syntax error, annotation without a node")
)

func Parse(reader io.Reader) (List, error) {
	s := newParseStack()
	scanner := NewScanner(bufio.NewReader(reader))
	var a []string
	for scanner.Scan() {
		tok := scanner.Token()
		switch tok.Type {
		case LineValue:
			s.top().add(Node{Value: tok.Content, Annotations: a})
			a = nil
		case Reference:
			s.top().add(Node{Value: tok.Content, IsReference: true, Annotations: a})
			a = nil
		case Annotation:
			a = append(a, tok.Content)
		case Indent:
			if len(a) > 0 {
				return nil, errAnnotationWithoutNode
			}
			last := s.top().last()
			if last == nil {
				return nil, errWrongIndent
			}
			s.push(&last.List)
		case Unindent:
			if len(a) > 0 {
				return nil, errAnnotationWithoutNode
			}
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
