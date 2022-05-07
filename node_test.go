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
