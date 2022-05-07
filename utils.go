package goart

import (
	"bytes"

	"golang.org/x/exp/constraints"
)

func min[T constraints.Ordered](a, b T) T {
	if a < b {
		return a
	} else {
		return b
	}
}

func replaceInnerNode[T any](old, new *innerNode[T]) {
	*old = *new
}

func replaceNode[T any](old, new *Node[T]) {
	*old = *new
}

func replaceNodeRef[T any](old **Node[T], new *Node[T]) {
	*old = new
}

func copyBytes(dst, src []byte, n int) {
	for i := 0; i < n && i < len(src) && i < len(dst); i++ {
		dst[i] = src[i]
	}
}

func terminate(key []byte) []byte {
	index := bytes.Index(key, []byte{0})
	if index < 0 {
		key = append(key, byte(0))
	}
	return key
}

func zeroValue[T any]() T {
	var result T
	return result
}
