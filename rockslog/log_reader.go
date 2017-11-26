package rockslog

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"unsafe"

	"github.com/pp-qq/rocksdb.go/rocksutil/crc32c"
)

const (
	kEOF = kMaxRecordType + 1
)

type Reporter interface {
	Corruption(size uint64, err error)
}

type Reader struct {
	file     *os.File
	reporter Reporter
	checksum bool

	/* block 用来存放 log file 中 one block 的内容.

	[start, end) 定义了 block 内尚未被解析的缓冲. start 总是位于 record 开头. end 总是等于 block size.

	last_block 若为真, 则表明 block 是 log file 中最后一个 block, last block 要么长度为 0, 要么长度小于
	kBlockSize.
	*/
	block      [kBlockSize]byte
	start, end uint
	last_block bool
}

func NewReader(path string, reporter Reporter, checksum bool) (*Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	// open success, 注意关闭 file.

	return &Reader{file: file, reporter: reporter, checksum: checksum}, nil
}

func (this *Reader) Close() error {
	return this.file.Close()
}

/* 若返回 nil, 则表明没有多余的内容了.

用户不应该修改返回值. 返回值一直有效直至下一次对 ReadRecord() 的调用.

当返回 nil 时, 若 Corruption 未被调用过, 则表明确实是由于没有 record 了; 若 Corruption() 被调用过, 则表明是
由于出错而导致 ReadRecord() 返回 nil. 这点与 rocksdb 不同, rocksdb 在出错时会略过出错的 record 继续往下读
取. */
func (this *Reader) ReadRecord() []byte {
	// checksum 为 0 时的特殊处理.
}

func isZeros(data []byte) bool {
	for _, d := range data {
		if d != 0 {
			return false
		}
	}
	return true
}

func (this *Reader) blockBuffer() []byte {
	return this.block[this.start:this.end]
}

func (this *Reader) zeroBlock() bool {
	return isZeros(this.blockBuffer())
}

func (this *Reader) reportCorruption(size uint64, err error) {
	if this.reporter != nil {
		this.report.Corruption(size, err)
	}
	return
}

/* 若成功读取一个 record, 则返回 record type, record content, 其中 record type 可取值参见
log_format.go 中定义. 若由于 io error 或者文件内容被毁害导致无法读取一个 record, 则返回 kEOF, nil.
*/
func (this *Reader) readPhysicalRecord() (int, []byte) {
	// 注意兼容 rocksdb 中 PosixMmapFile
	var err error
	for {
		restsize := this.end - this.start
		if restsize <= kHeaderSize {
			if !this.last_block {
				if !this.zeroBlock() {
					// log format spec: Any leftover bytes here form the trailer, which must
					// consist entirely of zero bytes.
					this.reportCorruption(restsize, fmt.Errorf("bad block trailer"))
					return kEOF, nil
				}

				this.end, err = this.file.Read(this.block[:])
				if err != nil && err != io.EOF {
					// 后续对 readPhysicalRecord() 的调用将返回 kEOF.
					this.end = 0
					this.start = this.end
					last_block = true

					this.reportCorruption(kBlockSize, err)
					return kEOF, nil
				}
				this.start = 0
				if this.end < kBlockSize {
					// rocksdb 中未考虑 this.end < 剩余文件大小, 或许这种情况并不会发生, 所以这里也不会考虑
					this.last_block = true
					continue
				}
				break
			} else if restsize >= kHeaderSize {
				break
			} else if this.zeroBlock() {
				return kEOF, nil
			} else {
				this.reportCorruption(restsize, fmt.Errof("truncated record at end of file"))
				return kEOF, nil
			}
		}
	}
	// restsize >= kHeaderSize, 如果这时候遇到 zero record type, 就认为是 PosixMmapFile 的填充, 就会
	// 认为文件已经结束了.
	// 未完待续

}
