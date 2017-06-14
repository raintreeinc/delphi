package uses

import (
	"path/filepath"
	"strings"
)

func trimExt(name string) string {
	return name[:len(name)-len(filepath.Ext(name))]
}

func includeString(arr []string, item string) []string {
	citem := strings.ToLower(item)
	for i, use := range arr {
		cuse := strings.ToLower(use)
		if cuse == citem {
			return arr
		}
		if cuse > citem {
			arr = append(arr, "")
			copy(arr[i+1:], arr[i:])
			arr[i] = item
			return arr
		}
	}
	return append(arr, item)
}

func contains(value string, list []string) bool {
	for _, item := range list {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}
