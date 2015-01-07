package tff

import (
	"bytes"
	"io"
)

type (
	Node struct {
		Value string
		List  List
	}
	List []Node
)

func (n *Node) Marshal(w io.Writer, prefix, indent string) error {
	if n.Value == "" && len(n.List) == 0 {
		return nil
	}
	if _, err := w.Write([]byte(prefix)); err != nil {
		return err
	}
	if n.Value == "" {
		if _, err := w.Write([]byte{'_'}); err != nil {
			return err
		}
	} else {
		if _, err := w.Write([]byte(n.Value)); err != nil {
			return err
		}
	}
	if len(n.List) == 0 {
		return nil
	}
	if _, err := w.Write([]byte{'\n'}); err != nil {
		return err
	}
	return n.List.Marshal(w, prefix+indent, indent)
}

func (list List) Marshal(w io.Writer, prefix, indent string) error {
	for i := range list {
		if i > 0 {
			if _, err := w.Write([]byte{'\n'}); err != nil {
				return err
			}
		}
		if err := list[i].Marshal(w, prefix, indent); err != nil {
			return err
		}
	}
	return nil
}

func (list List) String() string {
	var w bytes.Buffer
	list.Marshal(&w, "", "\t")
	return w.String()
}

func (n *Node) String() string {
	var w bytes.Buffer
	n.Marshal(&w, "", "\t")
	return w.String()
}
