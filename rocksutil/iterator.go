package rocksutil

/* Iterator, 其对应的底层数据模型类似与 std::map. 对 Key(), Value() 返回值的解释权由底层实现所有.

Iterator 并不是 goroutine 安全, 即不要在多个 goroutine 上不加任何同步的前提下并行访问 Iterator.

Iterator 在不需要的时候应该调用其 Close() 接口来释放资源.

当对 Iterator 的遍历结束之后, 应该通过 Status() 来判断结束的原因. 若 Status() 返回 nil, 则表明是因为没有多余
的数据了; 若 Status() 返回非 nil, 则表明是由于底层实现出了错导致遍历结束了. */
type Iterator interface {
	Close() error
	Valid() bool
	SeekToFirst()
	SeekToLast()
	Seek(key []byte)
	Next()
	Prev()
	Status() error
	/* Key(), Value() 返回迭代器当前位置对应的 key/value.

	调用者不应该修改返回值.

	返回值会一直有效直至下一次移动迭代器的位置 */
	Key() []byte
	Value() []byte
}
