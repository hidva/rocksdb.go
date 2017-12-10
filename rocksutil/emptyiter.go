package rocksutil

type EmptyIterator struct {
}

func (this EmptyIterator) Close() error {
	return nil
}

func (this EmptyIterator) Valid() bool {
	return false
}

func (this EmptyIterator) SeekToFirst() {
	return
}

func (this EmptyIterator) SeekToLast() {
	return
}

func (this EmptyIterator) Seek(key []byte) {
	return
}

func (this EmptyIterator) Next() {
	return
}

func (this EmptyIterator) Prev() {
	return
}

func (this EmptyIterator) Status() error {
	return nil
}

func (this EmptyIterator) Key() []byte {
	return nil
}

func (this EmptyIterator) Value() []byte {
	return nil
}

func NewEmptyIterator() EmptyIterator {
	return EmptyIterator{}
}
