package rocksutil

type CacheHandle interface {
	Value() interface{}
}

type Cache interface {
	EraseAll()
	Insert(key string, val interface{},
		charge int, deler func(key string, val interface{})) CacheHandle
	Lookup(key string) CacheHandle
	Release(handle CacheHandle)
	Erase(key string)
	NewId() int
}
