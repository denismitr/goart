package goart

type Tree[T any] struct {
	root *Node[T]
	size uint64
}

func NewTree[T any]() *Tree[T] {
	return &Tree[T]{root: nil, size: 0}
}

func (t *Tree[T]) Size() uint64 {
	return t.size
}

func (t *Tree[T]) Search(key []byte) T {
	key = terminate(key)
	return t.search(t.root, key, 0)
}

func (t *Tree[T]) search(current *Node[T], key []byte, offset int) T {
	for current != nil {
		if current.IsLeaf() {
			if current.leaf.isMatch(key) {
				return current.leaf.value
			}

			return zeroValue[T]()
		}

		in := current.inner
		if current.prefixMatchIndex(key, offset) != in.prefixLen {
			return zeroValue[T]()
		} else {
			offset += in.prefixLen
		}

		v := in.findChild(key[offset])
		if v == nil {
			return zeroValue[T]()
		}
		current = *(v)
		offset++
	}

	return zeroValue[T]()
}

type level[T any] struct {
	node  *Node[T]
	index int
}
