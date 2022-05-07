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

type leafNode[T any] struct {
	key   []byte
	value T
}

type Node[T any] struct {
	// innerNode map partial keys to child pointers
	inner *innerNode[T]

	// leaf is used to store a possible leaf
	leaf *leafNode[T]
}

type innerNode[T any] struct {
	meta
	nodeType nodeKind
	keys     []byte
	children []*Node[T]
}

func newLeafNode[T any](key []byte, value T) *Node[T] {
	newKey := make([]byte, len(key))
	copy(newKey, key)
	leaf := &leafNode[T]{key: newKey, value: value}
	return &Node[T]{leaf: leaf}
}

func newNode4[T any]() *Node[T] {
	in := &innerNode[T]{
		nodeType: node4,
		keys:     make([]byte, node4Max),
		children: make([]*Node[T], node4Max),
		meta: meta{
			prefix: make([]byte, maxPrefixLen),
		},
	}

	return &Node[T]{inner: in}
}

func newNode16[T any]() *Node[T] {
	in := &innerNode[T]{
		nodeType: node16,
		keys:     make([]byte, node16Max),
		children: make([]*Node[T], node16Max),
		meta: meta{
			prefix: make([]byte, maxPrefixLen),
		},
	}

	return &Node[T]{inner: in}
}

func newNode48[T any]() *Node[T] {
	in := &innerNode[T]{
		nodeType: node48,
		keys:     make([]byte, node48Max),
		children: make([]*Node[T], node48Max),
		meta: meta{
			prefix: make([]byte, maxPrefixLen),
		},
	}

	return &Node[T]{inner: in}
}

func newNode256[T any]() *Node[T] {
	in := &innerNode[T]{
		nodeType: node256,
		children: make([]*Node[T], node256Max),
		meta: meta{
			prefix: make([]byte, maxPrefixLen),
		},
	}

	return &Node[T]{inner: in}
}

// leaf node methods

func (ln *leafNode[T]) isMatch(key []byte) bool {
	return bytes.Equal(ln.key, key)
}

func (ln *leafNode[T]) isPrefixMatch(key []byte) bool {
	if len(key) > len(ln.key) {
		return false
	}

	return bytes.Equal(ln.key[:len(key)], key)
}

func (ln *leafNode[T]) prefixMatchIndex(other *leafNode[T], offset int) int {
	limit := min(len(ln.key), len(other.key)) - offset

	for i := 0; i < limit; i++ {
		if ln.key[offset+i] != other.key[offset+i] {
			return i
		}
	}

	return 0
}

// inner node methods

func (in *innerNode[T]) minSize() int {
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

func (in *innerNode[T]) maxSize() int {
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

func (in *innerNode[T]) findIndex(key byte) int {
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

func (in *innerNode[T]) findChild(key byte) **Node[T] {
	index := in.findIndex(key)

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

func (in *innerNode[T]) copyMeta(src *innerNode[T]) {
	in.meta = src.meta
}

func (in *innerNode[T]) isFull() bool { return in.size == in.maxSize() }

func (in *innerNode[T]) addChild(key byte, node *Node[T]) {
	if in.isFull() {
		in.grow()
		in.addChild(key, node)
		return
	}

	switch in.nodeType {
	case node4:
		i := 0
		// find the insertion point
		for ; i < in.size; i++ {
			if key < in.keys[i] {
				break
			}
		}

		// move all segments larger than key to the right
		for j := in.size; j > i; j-- {
			if in.keys[j-1] > key {
				in.keys[j] = in.keys[j-1]
				in.children[j] = in.children[j-1]
			}
		}

		// the actual insert action
		in.keys[i] = key
		in.children[i] = node
		in.size += 1
	case node16:
		i := in.size
		bitfield := uint(0)
		for j := 0; j < in.size; j++ {
			if in.keys[j] >= key {
				bitfield |= 1 << i
			}
		}
		mask := (1 << in.size) - 1
		bitfield &= uint(mask)
		if bitfield != 0 {
			i = bits.TrailingZeros(bitfield)
		}

		for j := in.size; j > i; j-- {
			if in.keys[j-1] > key {
				in.keys[j] = in.keys[j-1]
				in.children[j] = in.children[j-1]
			}
		}

		in.keys[i] = key
		in.children[i] = node
		in.size += 1
	case node48:
		i := 0
		for j := 0; j < len(in.children); j++ {
			if in.children[i] != nil {
				i++
			}
		}
		in.children[i] = node
		in.keys[key] = byte(i + 1)
		in.size += 1
	case node256:
		// this is basically a map of symbols
		// rather than an array
		in.children[key] = node
	}
}

func (in *innerNode[T]) grow() {
	switch in.nodeType {
	case node4:
		in16 := newNode16[T]().inner
		in16.copyMeta(in)
		for i := 0; i < in.size; i++ {
			in16.keys[i] = in.keys[i]
			in16.children[i] = in16.children[i]
		}
		replaceInnerNode(in, in16)
	case node16:
		in48 := newNode48[T]().inner
		in48.copyMeta(in)
		i := 0
		for j := 0; j < in.size; j++ {
			child := in.children[j]
			if child != nil {
				in48.keys[in.keys[j]] = byte(i + 1)
				in48.children[i] = child
				i++
			}
		}
		replaceInnerNode(in, in48)
	case node48:
		in256 := newNode256[T]().inner
		in256.copyMeta(in)

		for i := 0; i < len(in.keys); i++ {
			child := in.findChild(byte(i))
			if child != nil {
				in256.children[byte(i)] = *child
			}
		}
		replaceInnerNode(in, in256)
	case node256:
	}
}

// node methods

// IsLeaf - checks if node is a leaf
func (n *Node[T]) IsLeaf() bool {
	return n.leaf != nil
}

func (n *Node[T]) kind() nodeKind {
	if n.inner != nil {
		return n.inner.nodeType
	}

	if n.leaf != nil {
		return leaf
	}

	return corrupted
}

func (n *Node[T]) Key() []byte {
	if n.kind() != leaf {
		return nil
	}

	// return key without null as it is
	// being appended internally
	return n.leaf.key[:len(n.leaf.key)-1]
}

func (n *Node[T]) Value() interface{} {
	if n.kind() != leaf || n.leaf == nil {
		return nil
	}

	return n.leaf.value
}

func (n *Node[T]) prefixMatchIndex(key []byte, offset int) int {
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

func (n *Node[T]) maximum() *Node[T] {
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

func (n *Node[T]) minimum() *Node[T] {
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

func (n *Node[T]) shrink() {
	in := n.inner

	switch in.nodeType {
	case node4:
		firstChild := in.children[0]
		if !firstChild.IsLeaf() {
			child := firstChild.inner
			currentPrefixLen := in.prefixLen

			if currentPrefixLen < maxPrefixLen {
				in.prefix[currentPrefixLen] = in.keys[0]
				currentPrefixLen++
			}

			if currentPrefixLen < maxPrefixLen {
				childPrefixLen := min(child.prefixLen, maxPrefixLen-currentPrefixLen)
				copyBytes(in.prefix[currentPrefixLen:], child.prefix, childPrefixLen)
				currentPrefixLen = childPrefixLen
			}

			copyBytes(child.prefix, in.prefix, min(currentPrefixLen, maxPrefixLen))
			child.prefixLen += in.prefixLen + 1
		}

		replaceNode(n, firstChild)
	case node16:
		n4 := newNode4[T]()
		n4in := n4.inner
		n4in.copyMeta(n.inner)
		n4in.size = 0

		for i := 0; i < len(n4in.keys); i++ {
			n4in.keys[i] = in.keys[i]
			n4in.children[i] = in.children[i]
			n4in.size++
		}

		replaceNode(n, n4)
	case node48:
		n16 := newNode16[T]()
		n16in := n16.inner
		n16in.copyMeta(n.inner)
		n16in.size = 0

		for i := 0; i < len(in.keys); i++ {
			idx := in.keys[byte(i)]
			if idx > 0 {
				child := in.children[idx-1]
				if child != nil {
					n16in.children[n16in.size] = child
					n16in.keys[n16in.size] = byte(i)
					n16in.size++
				}
			}
		}

		replaceNode(n, n16)
	case node256:
		n48 := newNode48[T]()
		n48in := n48.inner
		n48in.copyMeta(n.inner)
		n48in.size = 0

		for i := 0; i < len(in.children); i++ {
			child := in.children[byte(i)]
			if child != nil {
				n48in.children[n48in.size] = child
				n48in.keys[byte(i)] = byte(n48in.size + 1)
				n48in.size++
			}
		}

		replaceNode(n, n48)
	}
}
