package rockslog

import (
	"encoding/binary"
	"os"

	"github.com/pp-qq/rocksdb.go/rocksutil/crc32c"
)

const (
	// 不变量: kTrailerSize >= kHeaderSize
	kTrailerSize = 7
)

var g_trailer [kTrailerSize]byte

var g_recordtype_checksum = [...]uint32{
	0,
	crc32c.Value([]byte{kFullType}),
	crc32c.Value([]byte{kFirstType}),
	crc32c.Value([]byte{kMiddleType}),
	crc32c.Value([]byte{kLastType}),
}

/* Writer.

WriteRecord(), 当其正常返回时, 便保证了 record 已经送入了内核, 但可能尚未送入持久性设备.

Sync() 负责确保 WriteRecord() 写入的 record 送入持久性设备.
*/
type Writer struct {
	file *os.File

	// Writer 当前所用 block 的剩余长度.
	blocksize int
}

func min(a, b int) int {
	if a <= b {
		return a
	} else {
		return b
	}
}

func NewWriter(path string) (*Writer, error) {
	file, err := os.Create(path)
	if err != nil {
		return nil, err
	}

	return &Writer{file: file}, nil
}

func (this *Writer) WriteRecord(record []byte) error {
	// 当 len(record) 为 0 时, 也要写入!
	// record[recordptr:] 为尚未被写入的内容.
	recordptr := 0
	for {
		if this.blocksize <= kTrailerSize {
			if this.blocksize > 0 {
				_, err := this.file.Write(g_trailer[:this.blocksize])
				if err != nil {
					return err
				}
			}
			this.blocksize = kBlockSize
		}

		// 此时 this.blocksize > kHeaderSize
		fragment_size := min(len(record)-recordptr, this.blocksize-kHeaderSize)
		fragment_start := recordptr
		fragment_end := fragment_start + fragment_size
		// 此时 recordptr <= fragment_start <= fragment_end <= len(record)

		recordtype := 0
		if fragment_start == 0 {
			if fragment_end == len(record) {
				recordtype = kFullType
			} else {
				recordtype = kFirstType
			}
		} else {
			if fragment_end == len(record) {
				recordtype = kLastType
			} else {
				recordtype = kMiddleType
			}
		}

		err := this.writePhysicalRecord(recordtype, record[fragment_start:fragment_end])
		if err != nil {
			return err
		}
		recordptr = fragment_end
		this.blocksize -= (kHeaderSize + fragment_size)

		if recordptr >= len(record) {
			break
		}
	}
	return nil
}

func (this *Writer) Sync() error {
	return this.file.Sync() // 这里貌似等同于 fsync(), 而不是 fdatasync()==
}

func (this *Writer) Close() error {
	return this.file.Close()
}

func (this *Writer) writePhysicalRecord(recordtype int, fragment []byte) error {
	var tmpbuf [kHeaderSize]byte
	var err error

	checksum := g_recordtype_checksum[recordtype]
	checksum = crc32c.Mask(crc32c.Extend(checksum, fragment))
	binary.LittleEndian.PutUint32(tmpbuf[:], checksum)
	binary.LittleEndian.PutUint16(tmpbuf[4:], uint16(len(fragment)))
	tmpbuf[6] = byte(recordtype)
	_, err = this.file.Write(tmpbuf[:])
	if err != nil {
		return err
	}

	_, err = this.file.Write(fragment)
	return err
}
