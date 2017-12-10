package rocksutil

import (
	"bytes"
)

type bytewiseComparator struct {
}

func (this bytewiseComparator) Compare(a, b []byte) int {
	return bytes.Compare(a, b)
}

func (this bytewiseComparator) Name() string {
	return "leveldb.BytewiseComparator"
}

// 有空优化一下这里吧==

func (this bytewiseComparator) FindShortestSeparator(start []byte, limit []byte) []byte {
	return start
}

func (this bytewiseComparator) FindShortSuccessor(start []byte) []byte {
	return start
}

var g_cmp Comparator = bytewiseComparator{}

func NewBytewiseComparator() Comparator {
	return g_cmp
}
