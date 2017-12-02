package rockslog

const (
	/* record type 枚举值. */
	kZeroType = iota
	kFullType
	kFirstType
	kMiddleType
	kLastType
	kMaxRecordType = kLastType

	kBlockSize = 32768

	kHeaderSize = 4 + 1 + 2
)
