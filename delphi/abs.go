package delphi

import (
	"path/filepath"
	"strings"
)

func AbsPath(pathlist string) string {
	paths := strings.Split(pathlist, ";")
	for i, path := range paths {
		if abs, err := filepath.Abs(path); err == nil {
			paths[i] = abs
		}
	}
	return strings.Join(paths, ";")
}

func AbsPaths(pathlists []string) []string {
	result := make([]string, len(pathlists))
	for i, pathlist := range pathlists {
		result[i] = AbsPath(pathlist)
	}
	return result
}
