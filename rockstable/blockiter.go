package rockstable

import (
	"sort"

	"github.com/pp-qq/rocksdb.go/rocksutil"
)

type blockIter struct {
	cmp rocksutil.Comparator
	blk *block

	// 若 key 为 nil, 则表明 invalid; 否则表明 valid.
	// curptr 与 key 对应, 表明当前 key 在 block 中的 offset.
	// nextptr 是 next key 在 block 中的 offset.
	// 不变量: this.err != nil 蕴含着 this.key == nil
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
	// 很显然由于 sort.search() 不支持中途退出, 除非我们使用 panic()/recovery() 机制.
	// 所以这里当遇到错误之后仍会继续循环几次, 但循环就循环吧, 谁让你给我不合法数据来着.
	idx := sort.Search(len(this.blk.restarts), func(restartidx int) bool {
		if this.err != nil {
			return false
		}

		restart := this.blk.restarts[restartidx]
		this.key, this.val, this.nextptr, this.err = this.blk.parse(restart, nil)
		return this.err == nil && this.cmp.Compare(this.key, key) > 0
	})
	// 此时并不能肯定 this.key, this.val 等变量是与 idx 对应的. 因为 sort.Search() 并未保证 idx 是最后一次
	// 调用 func 的.
	if this.err != nil {
		return
	}
	if idx > 0 {
		idx--
	}

	this.curptr = this.blk.restarts[idx]
	this.key = nil
	for {
		this.key, this.val, this.nextptr, this.err = this.blk.parse(this.curptr, this.key)
		if this.err != nil {
			return
		}

		if this.cmp.Compare(this.key, key) >= 0 {
			return
		}
		this.curptr = this.nextptr
	}
	return
}

func (this *blockIter) Next() {
	if this.nextptr >= len(this.blk.data) {
		this.key = nil
		return
	}

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

func (this *blockIter) Key() []byte {
	return this.key
}

func (this *blockIter) Value() []byte {
	return this.val
}

func (this *blockIter) advance(offset, stop int) {
	preoff := offset
	this.key = nil
	for offset < stop {
		this.key, this.val, this.nextptr, this.err = this.blk.parse(offset, this.key)
		if this.err != nil {
			return
		}
		preoff = offset
		offset = this.nextptr
	}
	this.curptr = preoff
	return
}
