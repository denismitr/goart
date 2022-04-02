package goart

type Tree struct {
	root *Node
	size uint64
}

func NewTree() *Tree {
	return &Tree{root: nil, size: 0}
}

func (t *Tree) Size() uint64 {
	return t.size
}

func (t *Tree) Search(key []byte) interface{} {
	key = terminate(key)
	return t.search(t.root, key, 0)
}

func (t *Tree) search(current *Node, key []byte, offset int) interface{} {
	for current != nil {
		if current.IsLeaf() {
			if current.leaf.isMatch(key) {
				return current.leaf.value
			}

			return nil
		}

		in := current.inner
		if current.prefixMatchIndex(key, offset) != in.prefixLen {
			return nil
		} else {
			offset += in.prefixLen
		}

		v := in.findChild(key[offset])
		if v == nil {
			return nil
		}
		current = *(v)
		offset++
	}

	return nil
}

type level struct {
	node  *Node
	index int
}
