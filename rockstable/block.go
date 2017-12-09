package rockstable

import (
	"encoding/binary"
	"fmt"
	"math"
	"unsafe"

	"github.com/pp-qq/rocksdb.go/rocksutil"
)

type block struct {
	// data, 不包括 restarts 数组部分, 只包含 kv 数据, data 为 nil 表明为空.
	// restarts, 可能为 nil, 表明为空. 可能存在 i, restarts[i] >= len(data) 这种不合法情况!
	data     []byte
	restarts []int
}

func newBlock(data []byte) (*block, error) {
	if len(data) < unsafe.Sizeof(uint32(0)) {
		return nil, fmt.Errorf("BadArg")
	}
	split := len(data) - unsafe.Sizeof(uint32(0))
	numrestarts := binary.LittleEndian.Uint32(data[split:])
	data = data[:split]

	restarts := make([]int, 0, numrestarts)
	sizerestarts := numrestarts * unsafe.Sizeof(uint32(0))
	if len(data) < sizerestarts {
		return nil, fmt.Errorf("BadArg")
	}
	split = len(data) - sizerestarts
	restartsdata := data[split:]
	data = data[:split]
	for len(restartsdata) >= unsafe.Sizeof(uint32(0)) {
		restarts = append(restarts, binary.LittleEndian.Uint32(restartsdata))
		restartsdata = restartsdata[unsafe.Sizeof(uint32(0)):]
	}

	if !((len(data) == 0 && len(restarts) == 0) || (len(data) > 0 && len(restarts) > 0)) {
		return nil, fmt.Errorf("BadArg")
	}
	return &block{data: data, restarts: restarts}, nil
}

/*
解析 offset 指定的 entry 得到 k, v 以及 nextoffset. 若在解析时发现 key shared_bytes 不为 0, 则 anchor
负责提供这部分内容.

返回的 k 可能是 this.data 或者 anchor 的 slice. 返回的 v 可能是 this.data 的 slice.
*/
func (this *block) parse(offset int, anchor []byte) (k, v []byte, nextoffset int, err error) {
	shared, readed := rocksutil.U32varint(this.data[offset:])
	if readed <= 0 || shared > len(anchor) {
		err = fmt.Errorf("invalid shared_bytes")
		return
	}
	offset += readed
	unshared, readed := rocksutil.U32varint(this.data[offset:])
	if readed <= 0 {
		err = fmt.Errorf("invalid unshared_bytes")
		return
	}
	offset += readed
	valsize, readed := rocksutil.U32varint(this.data[offset:])
	if readed <= 0 {
		err = fmt.Errorf("invalid value_length")
		return
	}
	offset += readed
	if offset+valsize+unshared > len(this.data) {
		err = fmt.Errorf("invalid unshared_bytes or value_length")
		return
	}

	k_end := offset + unshared
	if unshared <= 0 {
		k = anchor[:shared]
	} else if shared <= 0 {
		k = this.data[offset:k_end]
	} else {
		k = make([]byte, 0, shared+unshared)
		k = append(k, anchor[:shared]...)
		k = append(k, this.data[offset:k_end]...)
	}
	offset = k_end

	v_end := offset + valsize
	v = this.data[offset:v_end]
	nextoffset = v_end
	return
}

/*
与 parse() 完全一致, 除了不显式构造出 k, v.
*/
func (this *block) parseSkip(offset int, anchor []byte) (nextoffset int, err error) {
	// 做一些优化这里.
	_, _, nextoffset, err = this.parse(offset, anchor)
	return
}
