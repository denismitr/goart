package goart

import (
	"bytes"
	"testing"
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
