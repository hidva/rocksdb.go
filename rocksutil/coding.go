package rocksutil

import (
	"encoding/binary"
	"math"
)

// 语义等同于 binary.Uvarint(), 除了返回的是 uint32 类型.
func U32varint(buf []byte) (uint32, int) {
	val, bytes := binary.Uvarint(buf)
	if bytes <= 0 {
		return 0, bytes
	}
	if val > math.MaxUint32 {
		return 0, -bytes
	}
	return uint32(val), bytes
}
