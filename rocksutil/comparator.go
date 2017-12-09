package rocksutil

// Comparator 的实现需要做到 goroutine 安全.
type Comparator interface {
	Compare(a, b []byte) int
	Name() string
	// 注意返回值可能与 start 公用内存.
	FindShortestSeparator(start []byte, limit []byte) []byte
	// 注意返回值可能与 start 公用内存.
	FindShortSuccessor(start []byte) []byte
}
