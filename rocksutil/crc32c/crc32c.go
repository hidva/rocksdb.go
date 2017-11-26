package crc32c

import (
	"hash/crc32"
)

var g_table = crc32.MakeTable(crc32.Castagnoli)

func Extend(initcrc uint32, content []byte) uint32 {
	return crc32.Update(initcrc, g_table, content)
}

func Value(content []byte) uint32 {
	return Extend(0, content)
}

const kMaskDelta = 0xa282ead8

func Mask(crcval uint32) uint32 {
	return ((crcval >> 15) | (crcval << 17)) + kMaskDelta
}

func Unmask(maskedcrc uint32) uint32 {
	rot := maskedcrc - kMaskDelta
	return ((rot >> 17) | (rot << 15))
}
