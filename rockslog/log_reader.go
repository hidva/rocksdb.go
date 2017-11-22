package rockslog

type Reporter interface {
	Corruption(size uint64, err error)
}

type Reader struct {
}

func NewReader(path string, reporter Reporter, checksum bool) (*Reader, error) {

}

func (this *Reader) Close() error {

}

/* 若返回 nil, 则表明没有多余的内容了.

用户不应该修改返回值. 返回值一直有效直至下一次对 ReadRecord() 的调用. */
func (this *Reader) ReadRecord() []byte {

}
