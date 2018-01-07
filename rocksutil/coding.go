package rocksutil

import (
	"encoding/binary"
	"math"
)

const (
	UintLen32 = 4
)

// MaxUint, MinUint 的类型总是 uint 类型. MinInt, MaxInt 总是 int 类型.
// 仅当 unsafe.Sizeof(int) == Sizeof(uint) 时下面才是正确的. 感谢 go spec 指定: int, same size as uint
const (
	MinUint = uint(0)
	MaxUint = ^MinUint

	MaxInt = int(MaxUint >> 1)
	MinInt = ^MaxInt
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
