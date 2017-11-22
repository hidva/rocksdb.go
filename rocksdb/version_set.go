package rocksdb

type version struct {
}

/* 按我理解, C++ rocksdb 中 VersionSet 之所以是一个 Version 集合, 是因为 old version 可能正在被 reader
thread 使用着, 所以其内存不能被释放. 在 golang 中, 我们就把对 old version 的管理交给 gc 吧, VersionSet 中
 始终只有一个 current version. */
type versionSet struct {
}
