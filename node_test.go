package goart

import (
	"bytes"
	"testing"
)

var (
	emptyKey = []byte("")
)

func TestNode4_AddChild_PreserveSorted(t *testing.T) {
	n := newNode4[string]().inner

	for i := 4; i > 0; i-- {
		n.addChild(byte(i), newNode4[string]())
	}

	if n.size != 4 {
		t.Error("expected inner node size to be 4")
	}

	expectedKeys := []byte{1, 2, 3, 4}
	if !bytes.Equal(n.keys, expectedKeys) {
		t.Errorf("unexpected key sequence %+v", n.keys)
	}
}

func TestNode16_AddChild4_PreserveSort(t *testing.T) {
	n := newNode16[int]().inner

	for i := 16; i > 0; i-- {
		n.addChild(byte(i), newNode4[int]())
	}

	if n.size != 16 {
		t.Error("expected inner node size to be 16")
	}

	for i := 0; i < 16; i++ {
		want := byte(i + 1)
		got := n.keys[i]
		if want != got {
			t.Errorf("unexpected key sequence at index %d: want=%+v, got=%+v", i, want, got)
		}
	}
}

func TestGrow(t *testing.T) {
	nodes := []*Node[string]{newNode4[string](), newNode16[string](), newNode48[string]()}
	expectedKinds := []nodeKind{node16, node48, node256}

	for i, n := range nodes {
		n.inner.grow()
		want := expectedKinds[i]
		got := n.kind()
		if want != got {
			t.Errorf("Unexpected node kind at index %d after growing: want=%+v, got=%+v", i, want, got)
		}
	}
}

func TestShrink(t *testing.T) {
	nodes := []*Node[string]{newNode256[string](), newNode48[string](), newNode16[string](), newNode4[string]()}
	expectedKinds := []nodeKind{node48, node16, node4, leaf}

	for i, n := range nodes {
		in := n.inner
		for j := 0; j < in.minSize(); j++ {
			if n.kind() != node4 {
				in.addChild(byte(i), newNode4[string]())
			} else {
				in.addChild(byte(i), newLeafNode[string](emptyKey, ""))
			}
		}

		n.shrink()
		want := expectedKinds[i]
		got := n.kind()
		if want != got {
			t.Errorf("Unexpected node kind at index %d after shrinking: want=%+v, got=%+v", i, want, got)
		}
	}
}
