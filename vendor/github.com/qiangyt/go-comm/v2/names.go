package comm

import "strings"

func NameOfKey(key string) string {
	indexOfName := strings.LastIndex(key, ".")
	if indexOfName <= 0 {
		return key
	}
	return key[indexOfName+1:]
}
