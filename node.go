package goart

import (
	"bytes"
	"math/bits"
)

type nodeKind int8

const (
	corrupted nodeKind = iota
	node4
	node16
	node48
	node256
	leaf
)

const (
	node4Max = 4
	node4Min = 2

	node16Min = node4Max + 1
	node16Max = 16

	node48Min = node16Max + 1
	node48Max = 48

	node256Min = node48Max + 1
	node256Max = 256

	maxPrefixLen = 10
)

type meta struct {
	prefix    []byte
	prefixLen int
	size      int
}

type leafNode struct {
	key   []byte
	value interface{}
}

type Node struct {
	// innerNode map partial keys to child pointers
	inner *innerNode

	// leaf is used to store a possible leaf
	leaf *leafNode
}

type innerNode struct {
	meta
	nodeType nodeKind
	keys     []byte
	children []*Node
}

func newLeafNode(key []byte, value interface{}) *Node {
	newKey := make([]byte, len(key))
	copy(newKey, key)
	leaf := &leafNode{key: newKey, value: value}
	return &Node{leaf: leaf}
}

func newNode4() *Node {
	in := &innerNode{
		nodeType: node4,
		keys:     make([]byte, node4Max),
		children: make([]*Node, node4Max),
		meta: meta{
			prefix: make([]byte, maxPrefixLen),
		},
	}

	return &Node{inner: in}
}

func newNode16() *Node {
	in := &innerNode{
		nodeType: node16,
		keys:     make([]byte, node16Max),
		children: make([]*Node, node16Max),
		meta: meta{
			prefix: make([]byte, maxPrefixLen),
		},
	}

	return &Node{inner: in}
}

func newNode48() *Node {
	in := &innerNode{
		nodeType: node48,
		keys:     make([]byte, node48Max),
		children: make([]*Node, node48Max),
		meta: meta{
			prefix: make([]byte, maxPrefixLen),
		},
	}

	return &Node{inner: in}
}

func newNode256() *Node {
	in := &innerNode{
		nodeType: node256,
		children: make([]*Node, node256Max),
		meta: meta{
			prefix: make([]byte, maxPrefixLen),
		},
	}

	return &Node{inner: in}
}

// leaf node methods

func (ln *leafNode) isMatch(key []byte) bool {
	return bytes.Equal(ln.key, key)
}

func (ln *leafNode) isPrefixMatch(key []byte) bool {
	if len(key) > len(ln.key) {
		return false
	}

	return bytes.Equal(ln.key[:len(key)], key)
}

func (ln *leafNode) prefixMatchIndex(other *leafNode, offset int) int {
	limit := min(len(ln.key), len(other.key)) - offset

	for i := 0; i < limit; i++ {
		if ln.key[offset+i] != other.key[offset+i] {
			return i
		}
	}

	return 0
}

// inner node methods

func (in *innerNode) minSize() int {
	switch in.nodeType {
	case node4:
		return node4Min
	case node16:
		return node16Min
	case node48:
		return node48Min
	case node256:
		return node256Min
	default:
	}
	return 0
}

func (in *innerNode) maxSize() int {
	switch in.nodeType {
	case node4:
		return node4Max
	case node16:
		return node16Max
	case node48:
		return node48Max
	case node256:
		return node256Max
	default:
	}
	return 0
}

func (in *innerNode) index(key byte) int {
	switch in.nodeType {
	case node4:
		for i := 0; i < in.size; i++ {
			if in.keys[i] == key {
				return i
			}
		}
	case node16:
		bitfield := uint(0)
		for i := 0; i < in.size; i++ {
			if in.keys[i] == key {
				bitfield |= 1 << i
			}
		}
		mask := (1 << in.size) - 1
		bitfield &= uint(mask)
		if bitfield != 0 {
			return bits.TrailingZeros(bitfield)
		}
		return -1
	case node48:
		index := int(in.keys[key])
		if index > 0 {
			return int(index) - 1
		}

		return -1
	case node256:
		return int(key)
	}

	return -1
}

func (in *innerNode) findChild(key byte) **Node {
	index := in.index(key)

	switch in.nodeType {
	case node4, node16, node48:
		if index > 0 {
			return &in.children[index]
		}

		return nil
	case node256:
		child := in.children[key]
		if child != nil {
			return &in.children[key]
		}
	}

	return nil
}

// node methods

// IsLeaf - checks if node is a leaf
func (n *Node) IsLeaf() bool {
	return n.leaf != nil
}

func (n *Node) kind() nodeKind {
	if n.inner != nil {
		return n.inner.nodeType
	}

	if n.leaf != nil {
		return leaf
	}

	return corrupted
}

func (n *Node) Key() []byte {
	if n.kind() != leaf {
		return nil
	}

	// return key without null as it is
	// being appended internally
	return n.leaf.key[:len(n.leaf.key)-1]
}

func (n *Node) Value() interface{} {
	if n.kind() != leaf || n.leaf == nil {
		return nil
	}

	return n.leaf.value
}

func (n *Node) prefixMatchIndex(key []byte, offset int) int {
	in := n.inner
	p := in.prefix

	i := 0
	for ; i < in.prefixLen && offset+i < len(key) && key[offset+i] == p[i]; i++ {
		if i == maxPrefixLen-1 {
			min := n.minimum()
			p = min.leaf.key[offset:]
		}
	}

	return i
}

func (n *Node) maximum() *Node {
	in := n.inner

	switch n.kind() {
	case leaf:
		return n
	case node4, node16:
		return in.children[in.size-1].maximum()
	case node48:
		i := len(in.keys) - 1
		for in.keys[i] == 0 {
			i--
		}
		child := in.children[in.keys[i]-1]
		return child.maximum()
	case node256:
		i := len(in.children) - 1
		for i > 0 && in.children[byte(i)] == nil {
			i--
		}
		return in.children[i].maximum()
	}

	return n
}

func (n *Node) minimum() *Node {
	in := n.inner

	switch n.kind() {
	case node4, node16:
		return in.children[0].minimum()
	case node48:
		i := 0
		for in.keys[i] == 0 {
			i++
		}
		child := in.children[in.keys[i]-1]
		return child.minimum()
	case node256:
		i := 0
		for in.children[i] == nil {
			i++
		}
		return in.children[i].minimum()
	case leaf:
		return n
	}

	return n
}
