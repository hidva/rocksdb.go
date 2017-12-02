package rockslog

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"

	"github.com/pp-qq/rocksdb.go/rocksutil/crc32c"
)

const (
	kEOF = kMaxRecordType + 1
)

type Reporter interface {
	Corruption(size int, err error)
}

type Reader struct {
	file     *os.File
	reporter Reporter
	check    bool

	/* block 用来存放 log file 中 one block 的内容.

	[start, end) 定义了 block 内尚未被解析的缓冲. start 总是位于 record 开头. end 总是等于 block size.

	last_block 若为真, 则表明 block 是 log file 中最后一个 block, last block 要么长度为 0, 要么长度小于
	kBlockSize.
	*/
	block      [kBlockSize]byte
	start, end int
	last_block bool
}

func NewReader(path string, reporter Reporter, checksum bool) (*Reader, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	// open success, 注意关闭 file.

	return &Reader{file: file, reporter: reporter, check: checksum}, nil
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
	// 若 infragment 为 false, 表明尚未遇到 kFirstType, 此时 len(recordbuf) == 0.
	// 若 infragment 为 true, 表明已经遇到了 kFirstType, 此时 recordbuf 存放着相应的内容.
	infragment := false
	var recordbuf []byte

	for {
		switch recordtype, fragment := this.readPhysicalRecord(); recordtype {
		case kFullType:
			if infragment {
				this.reportCorruption(len(recordbuf), fmt.Errorf("partial record without end"))
				return nil
			}
			return fragment

		case kFirstType:
			if infragment {
				this.reportCorruption(len(recordbuf), fmt.Errorf("partial record without end"))
				return nil
			}
			recordbuf = append(recordbuf, fragment...)
			infragment = true
		case kMiddleType:
			if !infragment {
				const errmsg = "missing start of fragmented record"
				this.reportCorruption(len(fragment), fmt.Errorf(errmsg))
				return nil
			}
			recordbuf = append(recordbuf, fragment...)
		case kLastType:
			if !infragment {
				const errmsg = "missing start of fragmented record"
				this.reportCorruption(len(fragment), fmt.Errorf(errmsg))
				return nil
			}
			recordbuf = append(recordbuf, fragment...)
			return recordbuf
		case kEOF:
			if infragment {
				this.reportCorruption(len(recordbuf), fmt.Errorf("partial record without end"))
			}
			return nil
		default:
			const errmsg = "unknown record type"
			this.reportCorruption(len(fragment)+len(recordbuf), fmt.Errorf(errmsg))
			return nil
		}
	}
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

func (this *Reader) reportCorruption(size int, err error) {
	if this.reporter != nil {
		this.reporter.Corruption(size, err)
	}
	return
}

/* 若成功读取一个 record, 则返回 record type, record content, 其中 record type 可取值参见
log_format.go 中定义. 若由于 io error 或者文件内容被毁害导致无法读取一个 record, 则返回 kEOF, nil.
*/
func (this *Reader) readPhysicalRecord() (int, []byte) {
	// 注意兼容 rocksdb 中 PosixMmapFile. readPhysicalRecord() 不对 recordtype 进行过多地解读.
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
					this.last_block = true

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
				this.reportCorruption(restsize, fmt.Errorf("truncated record at end of file"))
				return kEOF, nil
			}
		}
		break
	}
	// restsize >= kHeaderSize, 如果这时候遇到 zero record type, 就认为是 PosixMmapFile 的填充, 就会
	// 认为文件已经结束了.

	bufstart := this.start
	checksum := binary.LittleEndian.Uint32(this.block[bufstart:this.end])
	bufstart += 4
	length := int(binary.LittleEndian.Uint16(this.block[bufstart:this.end]))
	bufstart += 2
	recordtype := this.block[bufstart]
	bufstart += 1
	if length > this.end-bufstart {
		this.reportCorruption(this.end-this.start, fmt.Errorf("bad record length"))
		return kEOF, nil
	}
	if recordtype == kZeroType {
		if checksum != 0 || length != 0 {
			// 此时这里可能是一个 recordtype 为 kZeroType 的合法 record.
			this.reportCorruption(kHeaderSize+length, fmt.Errorf("unknown record type"))
		}
		return kEOF, nil
	}
	if this.check &&
		checksum != crc32c.Mask(crc32c.Value(this.block[bufstart-1:bufstart+length])) {
		this.reportCorruption(length, fmt.Errorf("checksum mismatch"))
		return kEOF, nil
	}

	this.start = bufstart + length
	return int(recordtype), this.block[bufstart:this.start]
}
