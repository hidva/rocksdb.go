package rocksutil

import (
	"container/list"
	"sync"
)

type lruCacheHandle struct {
	key    string
	val    interface{}
	charge int
	deler  func(key string, val interface{})

	ref int
}

func (this *lruCacheHandle) Value() interface{} {
	return this.val
}

type lruCache struct {
	capacity int // 只读

	mux sync.Mutex

	// 最近一次被访问的放在 list 表头位置. handles 中存放的是 *lruCacheHandle 类型.
	handles *list.List
	elems   map[string]*list.Element
	usage   int

	id int
}

func (this *lruCache) EraseAll() {
	this.mux.Lock()
	defer this.mux.Unlock()

	for this.handles.Len() > 0 {
		elem := this.handles.Front()
		key := elem.Value.(*lruCacheHandle).key
		this.erase(key, elem)
	}
	return
}

func (this *lruCache) Insert(key string, val interface{},
	charge int, deler func(key string, val interface{})) CacheHandle {

	this.mux.Lock()
	defer this.mux.Unlock()

	if elem, ok := this.elems[key]; ok {
		this.removeElem(elem)
	}

	handle := &lruCacheHandle{key: key, val: val, charge: charge, deler: deler, ref: 2}
	this.elems[key] = this.handles.PushFront(handle)
	this.usage += charge

	this.shrinkToCapacity()
	return handle
}

func (this *lruCache) Lookup(key string) CacheHandle {
	this.mux.Lock()
	defer this.mux.Unlock()

	elem, ok := this.elems[key]
	if !ok {
		return nil
	}

	this.handles.MoveToFront(elem)
	handle := elem.Value.(*lruCacheHandle)
	handle.ref++
	return handle
}

func (this *lruCache) Release(handle CacheHandle) {
	this.mux.Lock()
	defer this.mux.Unlock()

	this.unref(handle.(*lruCacheHandle))
	return
}

func (this *lruCache) Erase(key string) {
	this.mux.Lock()
	defer this.mux.Unlock()

	elem, ok := this.elems[key]
	if !ok {
		return
	}

	this.erase(key, elem)
	return
}

func (this *lruCache) NewId() int {
	this.mux.Lock()
	defer this.mux.Unlock()

	this.id++
	return this.id
}

func (this *lruCache) unref(handle *lruCacheHandle) {
	handle.ref--
	if handle.ref <= 0 {
		this.usage -= handle.charge
		handle.deler(handle.key, handle.val)
	}
	return
}

func (this *lruCache) removeElem(elem *list.Element) {
	this.unref(this.handles.Remove(elem).(*lruCacheHandle))
	return
}

func (this *lruCache) shrinkToCapacity() {
	for this.usage > this.capacity && this.handles.Len() > 0 {
		handle := this.handles.Remove(this.handles.Back()).(*lruCacheHandle)
		key := handle.key
		this.unref(handle)
		delete(this.elems, key)
	}
	return
}

func (this *lruCache) erase(key string, elem *list.Element) {
	this.removeElem(elem)
	delete(this.elems, key)
	return
}

func NewLRUCache(capacity int) Cache {
	return &lruCache{
		capacity: capacity,
		handles:  list.New(),
		elems:    make(map[string]*list.Element),
	}
}
