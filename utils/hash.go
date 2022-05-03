package utils

import (
	"crypto/md5"
	"sort"
	"unsafe"
)

func ByteArrayToInt(arr []byte) int64 {
	val := int64(0)
	size := len(arr)
	for i := 0; i < size; i++ {
		*(*uint8)(unsafe.Pointer(uintptr(unsafe.Pointer(&val)) + uintptr(i))) = arr[i]
	}
	return val
}

func GetHash(id string) int64 {
	data := []byte(id)
	hash := md5.Sum(data)
	return ByteArrayToInt(hash[:])
}

func GetHashFromKeys(keys []string) int64 {
	concatKey := ""
	for _, key := range keys {
		concatKey += key
	}
	return GetHash(concatKey)
}
func Sort(int64Values []int64) []int64 {
	out := make([]int64, len(int64Values))
	int64AsIntValues := make([]int, len(int64Values))

	for i, val := range int64Values {
		int64AsIntValues[i] = int(val)
	}

	sort.Ints(int64AsIntValues)

	for i, val := range int64AsIntValues {
		out[i] = int64(val)
	}

	return out
}

func Int64ToInt(int64Values []int64) []int {
	int64AsIntValues := make([]int, len(int64Values))

	for i, val := range int64Values {
		int64AsIntValues[i] = int(val)
	}
	return int64AsIntValues
}

func IsInNodeHash(nodeHashes []int64, hash int64) bool {
	for _, nodeHash := range nodeHashes {
		if nodeHash == hash {
			return true
		}
	}
	return false
}
