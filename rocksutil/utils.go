package rocksutil

import (
	"sort"
)

/*
在 vals 中找到并返回最后一个 < val 的元素的下标. 若不存在, 则返回 -1.

此时假设 vals 增序排列.
*/
func FindLastIntLess(val int, vals []int) int {
	return sort.SearchInts(vals, val) - 1
}
