package walk

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func istemp(name string) bool {
	return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "~") || strings.HasSuffix(name, "~")
}

func Globs(globs []string, filenames chan string, errors chan error) {
	for _, glob := range globs {
		Glob(glob, filenames, errors)
	}
}

func Glob(glob string, filenames chan string, errors chan error) {
	matches, err := filepath.Glob(glob)
	if err != nil {
		errors <- fmt.Errorf("GLOB %v: %v", glob, err)
		return
	}

	for _, match := range matches {
		info, err := os.Lstat(match)
		if err != nil {
			errors <- err
		}
		if info.IsDir() {
			err := filepath.Walk(glob, func(path string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if ext := strings.ToLower(filepath.Ext(path)); ext != ".pas" && ext != ".inc" {
					return nil
				}
				if istemp(info.Name()) {
					return filepath.SkipDir
				}
				filenames <- path
				return nil
			})
			if err != nil {
				errors <- err
			}
		} else {
			filenames <- match
		}
	}
}
