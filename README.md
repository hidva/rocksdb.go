rocksdb.go 使用纯 golang 实现的 rocksdb, 与 rocksdb 100% 兼容.

目前与 [rocksdb-f67e15e](https://github.com/facebook/rocksdb/commit/f67e15e) 兼容. 随着对 rocksdb 源码的学习, rocksdb.go 也会随之提高兼容 rocksdb 的版本.

rocksdb benchmark:

```sh
$ ./db_bench --num=100000 --histogram=0 --compression_ratio=0.33 --value_size=100 --write_buffer_size=1048576
open         :   7845.879 micros/op;
writeseq     :      8.017 micros/op;  13.3 MB/s      
writeseq     :     15.145 micros/op;   7.1 MB/s      
writerandom  :     20.595 micros/op;   5.2 MB/s      
sync         :      1.907 micros/op;
tenth        :      0.954 micros/op;
tenth        :      0.954 micros/op;
writerandom  :   1865.993 micros/op;   0.1 MB/s    
nosync       :      3.099 micros/op;
normal       :      0.954 micros/op;
readseq      :      1.033 micros/op; 103.4 MB/s      
readrandom   :     29.711 micros/op;                 
compact      : 656408.072 micros/op;
readseq      :      0.673 micros/op; 158.7 MB/s      
readrandom   :      9.273 micros/op;                 
writebig     :   2224.159 micros/op;  42.9 MB/s   
```

rocksdb.go benchmark:

```
hh
```

## 实现细节

一些没地方放置的实现细节都会塞到这里.

-   去掉原 rocksdb env/port 模块. 按我理解 env/port 主要有如下几个功能: 
    
    方便移植; 这个直接使用 golang 标准库也能达到方便移植的目的, 所以不需要单独整个 env/port.
    
    可以很方便地 hook; 比如实现一个 SlowIOEnv, 限制 io 速率; 从而观察 leveldb 在低 io 下的性能等参数. 这个可以通过 docker/cgroups 等类似工具来控制. 所以也不需要单独整个 env/port
    
    所以去掉了 env/port 模块. (好吧我承认是我懒得写的缘故==

-   关于注释; 除非有必要, 否则 rocksdb.go 中大部分 struct/func 等都不会有注释性信息. 因为我在学习 rocksdb 时已经做了炒鸡多的注释说明, 参见 [rocksdb@study](https://github.com/pp-qq/rocksdb/tree/study)

-    关于单元测试, benchmark 就以后再补充吧. TDD? 不存在的.
