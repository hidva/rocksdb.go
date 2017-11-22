package rockslog

const (
	kZeroType = iota
	kFullType
	kFirstType
	kMiddleType
	kLastType
	kMaxRecordType = kLastType

	kBlockSize = 32768

	kHeaderSize = 4 + 1 + 2
)
