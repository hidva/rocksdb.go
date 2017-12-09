package rockstable

import (
	"github.com/pp-qq/rocksdb.go/rocksutil"
)

type blockIter struct {
	cmp rocksutil.Comparator
	blk *block

	// 若 key 为 nil, 则表明 invalid; 否则表明 valid.
	// curptr 与 key 对应, 表明当前 key 在 block 中的 offset.
	// nextptr 是 next key 在 block 中的 offset.
	curptr  int
	nextptr int
	key     []byte
	val     []byte
	err     error
}

// blk 必须是 newBlock() 返回的非空 blk, 函数内不会检测!!!
func newBlockIter(blk *block, comp rocksutil.Comparator) *blockIter {
	return &blockIter{cmp: comp, blk: blk}
}

func (this *blockIter) Close() error {
	return nil
}

func (this *blockIter) Valid() bool {
	return this.key != nil
}

func (this *blockIter) SeekToFirst() {
	this.nextptr = 0
	this.key = nil
	this.Next()
	return
}

func (this *blockIter) SeekToLast() {
	this.advance(this.blk.restarts[len(this.blk.restarts)-1], len(this.blk.data))
	return
}

func (this *blockIter) Seek(key []byte) {

}

func (this *blockIter) Next() {
	this.curptr = this.nextptr
	this.key, this.val, this.nextptr, this.err = this.blk.parse(this.curptr, this.key)
	return
}

func (this *blockIter) Prev() {
	if this.curptr <= 0 {
		this.key = nil
		return
	}
	restart_idx := rocksutil.FindLastIntLess(this.curptr, this.blk.restarts)
	if restart_idx == -1 {
		// 正常情况下这是不可能成立的, 除非有人恶意使坏!
		this.key = nil
		return
	}
	this.advance(this.blk.restarts[restart_idx], this.curptr)
	return
}

func (this *blockIter) Status() error {
	return this.err
}

func (this *blockIter) advance(offset, stop int) {
	preoff := offset
	this.key = nil
	for offset < stop {
		tihs.key, this.val, this.nextptr, this.err = this.blk.parse(offset, this.key)
		if this.err != nil {
			return
		}
		preoff = offset
		offset = this.nextptr
	}
	this.curptr = preoff
	return
}
