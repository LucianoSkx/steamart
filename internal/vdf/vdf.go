package vdf

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// Node é um objeto KeyValues (dicionário ordenado).
type Node struct {
	Children []*Child
}

type Child struct {
	Key   string
	Value any // *Node | string | int32
}

func (n *Node) Get(key string) any {
	for _, c := range n.Children {
		if c.Key == key {
			return c.Value
		}
	}
	return nil
}

func (n *Node) Str(key string) string {
	if v := n.Get(key); v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func (n *Node) Int(key string) int32 {
	if v := n.Get(key); v != nil {
		if i, ok := v.(int32); ok {
			return i
		}
	}
	return 0
}

func (n *Node) Dict(key string) *Node {
	if v := n.Get(key); v != nil {
		if d, ok := v.(*Node); ok {
			return d
		}
	}
	return nil
}

func (n *Node) All(key string) []*Child {
	var out []*Child
	for _, c := range n.Children {
		if c.Key == key {
			out = append(out, c)
		}
	}
	return out
}

type reader struct {
	b []byte
	i int
}

func (r *reader) byte() (byte, bool) {
	if r.i >= len(r.b) {
		return 0, false
	}
	b := r.b[r.i]
	r.i++
	return b, true
}

func (r *reader) peek() (byte, bool) {
	if r.i >= len(r.b) {
		return 0, false
	}
	return r.b[r.i], true
}

func (r *reader) str() string {
	start := r.i
	for r.i < len(r.b) {
		if r.b[r.i] == 0 {
			s := string(r.b[start:r.i])
			r.i++
			return s
		}
		r.i++
	}
	return string(r.b[start:])
}

func (r *reader) int32() int32 {
	if r.i+4 > len(r.b) {
		return 0
	}
	v := int32(binary.LittleEndian.Uint32(r.b[r.i : r.i+4]))
	r.i += 4
	return v
}

// Set define ou atualiza o valor de uma chave neste nó.
func (n *Node) Set(key string, value any) {
	for _, c := range n.Children {
		if c.Key == key {
			c.Value = value
			return
		}
	}
	n.Children = append(n.Children, &Child{Key: key, Value: value})
}

// Marshal serializa um nó KeyValues (binário Valve) de volta a bytes.
func Marshal(n *Node) []byte {
	var buf []byte
	return appendNode(buf, n)
}

func appendNode(buf []byte, n *Node) []byte {
	for _, c := range n.Children {
		buf = appendChild(buf, c)
	}
	buf = append(buf, 0x08)
	return buf
}

func appendChild(buf []byte, c *Child) []byte {
	switch v := c.Value.(type) {
	case *Node:
		buf = append(buf, 0x00)
		buf = appendStr(buf, c.Key)
		buf = appendNode(buf, v)
	case string:
		buf = append(buf, 0x01)
		buf = appendStr(buf, c.Key)
		buf = appendStr(buf, v)
	case int32:
		buf = append(buf, 0x02)
		buf = appendStr(buf, c.Key)
		buf = appendInt32(buf, v)
	}
	return buf
}

func appendStr(buf []byte, s string) []byte {
	buf = append(buf, []byte(s)...)
	buf = append(buf, 0)
	return buf
}

func appendInt32(buf []byte, v int32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, uint32(v))
	return append(buf, b...)
}

// Parse lê um arquivo KeyValues binário (formato usado pela Steam em shortcuts.vdf).
func Parse(data []byte) (*Node, error) {
	r := &reader{b: data}
	root := &Node{}
	if err := r.readNode(root); err != nil {
		return nil, err
	}
	return root, nil
}

func (r *reader) readNode(n *Node) error {
	for {
		t, ok := r.byte()
		if !ok {
			return errors.New("fim inesperado do arquivo VDF")
		}
		if t == 0x08 {
			return nil
		}
		key := r.str()
		var val any
		switch t {
		case 0x00:
			nb, _ := r.peek()
			if nb == 0x08 || nb == 0x00 || nb == 0x01 || nb == 0x02 {
				child := &Node{}
				if err := r.readNode(child); err != nil {
					return err
				}
				val = child
			} else {
				val = r.str()
			}
		case 0x01:
			val = r.str()
		case 0x02:
			val = r.int32()
		case 0x03:
			// float32 (ignorado como valor numérico)
			val = r.int32()
		default:
			return fmt.Errorf("tipo VDF desconhecido 0x%02x em %d", t, r.i)
		}
		n.Children = append(n.Children, &Child{Key: key, Value: val})
	}
}
