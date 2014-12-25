// Copyright 2014, Hǎiliàng Wáng. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package flow

import (
	"bytes"
	"io"
	"reflect"
)

type Composer interface {
	io.Writer
	Init(prefix, indent string)
	ComposeAny(v reflect.Value) error
	ComposeList(length int, composeElem func(i int) error) error
	Inline() bool
}

type composer struct {
	bytes.Buffer
	inline    bool
	prefix    string
	indent    string
	depth     int
	lineClear bool
}

func (t *composer) Init(prefix, indent string) {
	t.inline = (indent == "")
	t.prefix, t.indent = prefix, indent
	t.depth = 0
	t.lineClear = true
}

func (t *composer) Inline() bool {
	return t.inline
}

// length is the length of the list
func (t *composer) ComposeList(length int, composeElem func(i int) error) error {
	if t.inline {
		t.WriteString("{")
		t.lineClear = true
	}
	t.depth++
	for i := 0; i < length; i++ {
		t.listSep()
		t.writePrefix()
		t.lineClear = false
		if err := composeElem(i); err != nil {
			return err
		}
	}
	t.depth--
	if t.inline {
		t.WriteString("}")
	}
	return nil
}

func (t *composer) start() {
}

func (t *composer) stop() {
	t.inline = false
}

func (t *composer) writePrefix() {
	t.WriteString(t.prefix)
	if !t.inline {
		for i := 1; i < t.depth; i++ {
			t.WriteString(t.indent)
		}
	}
}

func (t *composer) listSep() {
	if !t.lineClear {
		if t.inline {
			t.WriteString(", ")
		} else {
			t.WriteString("\n")
			t.lineClear = true
		}
	}
}

func (t *composer) encodeNil() {
	t.WriteString("nil")
}
