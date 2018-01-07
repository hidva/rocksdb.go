package rockstable

import (
	"encoding/binary"
	"fmt"

	"github.com/pp-qq/rocksdb.go/rocksutil"
)

type block struct {
	// data, 不包括 restarts 数组部分, 只包含 kv 数据, data 为 nil 表明为空.
	// restarts, 可能为 nil, 表明为空. 可能存在 i, restarts[i] >= len(data) 这种不合法情况!
	// data, restarts 同时为空, 或者同时不为空.
	data     []byte
	restarts []int
}

// 另 (val, err) 标识返回值; 若 err == nil, 则表明此次转换是安全无溢出的, val 存放着转换结果; 若
// err != nil, 则表明无法转换, 此时 val 不确定.
func ui322i(val uint32) (int, error) {
	// 可以证明, 这里总是正确安全无溢出的. 首先 uint(MaxInt) 总是安全无溢出的.
	// 当 sizeof(uint) >= sizeof(uint32) 时, uint(val) 是安全无溢出的.
	// 当 sizeof(uint) < sizeof(uint32) 时; 此时有两种情况: val > MaxUint, 此时 uint(val) == MaxUint,
	// 又因为 MaxUint > uint(MaxInt), 所以这时也是安全的; val <= MaxUint, 此时 uint(val) 是安全无溢出的.
	if uint(val) > uint(rocksutil.MaxInt) {
		return 0x66ccff, fmt.Errorf("BadArg")
	}
	return int(val), nil
}

func NewBlock(data []byte) (*block, error) {
	if len(data) < rocksutil.UintLen32 {
		return nil, fmt.Errorf("BadArg")
	}
	split := len(data) - rocksutil.UintLen32
	numrestarts, err := ui322i(binary.LittleEndian.Uint32(data[split:]))
	if err != nil {
		return nil, err
	}
	data = data[:split]

	restarts := make([]int, 0, numrestarts)
	sizerestarts := numrestarts * rocksutil.UintLen32
	if len(data) < sizerestarts {
		return nil, fmt.Errorf("BadArg")
	}
	split = len(data) - sizerestarts
	restartsdata := data[split:]
	data = data[:split]
	var restart int
	for len(restartsdata) >= rocksutil.UintLen32 {
		restart, err = ui322i(binary.LittleEndian.Uint32(restartsdata))
		if err != nil {
			return nil, err
		}
		restarts = append(restarts, restart)
		restartsdata = restartsdata[rocksutil.UintLen32:]
	}

	if !((len(data) == 0 && len(restarts) == 0) || (len(data) > 0 && len(restarts) > 0)) {
		return nil, fmt.Errorf("BadArg")
	}
	return &block{data: data, restarts: restarts}, nil
}

func (this *block) NewIterator(cmp rocksutil.Comparator) rocksutil.Iterator {
	if len(this.data) <= 0 {
		return rocksutil.NewEmptyIterator()
	} else {
		return newBlockIter(this, cmp)
	}
}

/*
解析 offset 指定的 entry 得到 k, v 以及 nextoffset. 若在解析时发现 key shared_bytes 不为 0, 则 anchor
负责提供这部分内容.

返回的 k 可能是 this.data 或者 anchor 的 slice. 返回的 v 可能是 this.data 的 slice.
*/
func (this *block) parse(offset int, anchor []byte) (k, v []byte, nextoffset int, err error) {
	tmp, readed := rocksutil.U32varint(this.data[offset:])
	shared, err := ui322i(tmp)
	if readed <= 0 || err != nil || shared > len(anchor) {
		err = fmt.Errorf("invalid shared_bytes")
		return
	}
	offset += readed
	tmp, readed = rocksutil.U32varint(this.data[offset:])
	unshared, err := ui322i(tmp)
	if readed <= 0 || err != nil {
		err = fmt.Errorf("invalid unshared_bytes")
		return
	}
	offset += readed
	tmp, readed = rocksutil.U32varint(this.data[offset:])
	valsize, err := ui322i(tmp)
	if readed <= 0 || err != nil {
		err = fmt.Errorf("invalid value_length")
		return
	}
	offset += readed
	// 这里可能产生溢出啊! 啊, 安全编码好麻烦啊!
	// 所幸 golang 在 [Index expressions](http://godoc.hhhh233.xyz/ref/spec#Index_expressions) 中作
	// 了下标范围检测. 所以这里如果溢出了会导致下面的 panic, 不会导致非法内存访问. 一切还在可控之中!
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
