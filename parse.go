// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"fmt"
	"io"
	"reflect"
)

type Parser interface {
	Value() ([]byte, error)
	ParseList(walkFn func(int) error) error
	ParseAny(v reflect.Value) error
	GoToOnlyChild() error
}

type parser struct {
	*scanner
}

func newParser(r io.Reader) *parser {
	t := &parser{scanner: newScanner(r)}
	return t
}

func (t *parser) ParseList(parseElem func(int) error) error {
	if !t.isList() {
		return t.Error()
	}
	if err := t.GoToOnlyChild(); err != nil {
		return err
	}
	for i := 0; !t.isListEnd(); i++ {
		if err := parseElem(i); err != nil {
			return err
		}
		if err := t.nextSibling(); err != nil {
			return err
		}
	}
	return nil
}

func (t *parser) GoToOnlyChild() error {
	return t.next()
}

func (t *parser) nextSibling() error {
	if t.isSep() {
		return t.next()
	}
	if !t.isListEnd() {
		return t.error()
	}
	return nil
}

func (t *parser) next() error {
	for t.Scan() {
		if t.Token().ID == tokenComment {
			continue
		} else {
			break
		}
	}
	return t.Error()
}

func (t *parser) isType() bool {
	return t.isValue() && len(t.Token().Value) > 0 && t.Token().Value[0] == '!'
}

func (t *parser) isRef() bool {
	return t.isValue() && len(t.Token().Value) > 0 && t.Token().Value[0] == '^'
}

func (t *parser) isList() bool {
	return t.Token().ID == tokenLeftBrace
}

func (t *parser) isNil() bool {
	val, err := t.Value()
	if err != nil {
		return false
	}
	return string(val) == "nil"
}

func (t *parser) isListEnd() bool {
	return t.Token().ID == tokenRightBrace
}

func (t *parser) isSepOrListEnd() bool {
	return t.isSep() || t.isListEnd()
}

func (t *parser) isSep() bool {
	return t.Token().ID == tokenComma
}

func (t *parser) isValue() bool {
	return t.Token().ID == tokenString
}

func (t *parser) Value() ([]byte, error) {
	if !t.isValue() {
		return nil, t.error()
	}
	return t.Token().Value, nil
}

func (t *parser) error() error {
	return fmt.Errorf("unexpected token: %v, %s", t.Token().ID,
		string(t.Token().Value))
}
