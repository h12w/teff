package core

import (
	"bufio"
	"bytes"
	"io"
)

func (list List) String() string {
	var w bytes.Buffer
	list.Marshal(&w, "", "\t")
	return w.String()
}

func (list List) Marshal(w io.Writer, prefix, indent string) error {
	ew := newErrWriter(w)
	list.marshal(&ew, "", "\t")
	ew.flush()
	return ew.err
}

func (list List) marshal(w *errWriter, prefix, indent string) {
	for i := range list {
		if i > 0 {
			w.writeByte('\n')
		}
		list[i].marshal(w, prefix, indent)
	}
}

func (n *Node) marshal(w *errWriter, prefix, indent string) {
	for _, a := range n.Annotations {
		w.writeString(prefix)
		w.writeByte('#')
		w.writeString(a)
		w.writeByte('\n')
	}
	if n.Value != "" {
		w.writeString(prefix)
		if n.IsReference {
			w.writeByte('^')
		}
		w.writeString(n.Value)
	}
	if len(n.List) > 0 {
		w.writeByte('\n')
		n.List.marshal(w, prefix+indent, indent)
	}
}

type errWriter struct {
	w   *bufio.Writer
	err error
}

func newErrWriter(w io.Writer) errWriter {
	return errWriter{w: bufio.NewWriter(w)}
}

func (w *errWriter) writeString(s string) {
	if w.err != nil {
		return
	}
	_, w.err = w.w.WriteString(s)
}

func (w *errWriter) writeByte(b byte) {
	if w.err != nil {
		return
	}
	w.err = w.w.WriteByte(b)
}

func (w *errWriter) flush() {
	w.w.Flush()
}
