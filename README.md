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


