package walk

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func IsDelphiFile(file string) bool {
	ext := strings.ToLower(filepath.Ext(file))
	return ext == ".pas" || ext == ".inc" || ext == ".dpr"
}

func istemp(name string) bool {
	return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "~") || strings.HasSuffix(name, "~")
}

func Globs(globs []string, filenames chan string, errors chan error, care func(file string) bool) {
	for _, glob := range globs {
		Glob(glob, filenames, errors, care)
	}
}

func Glob(glob string, filenames chan string, errors chan error, care func(file string) bool) {
	matches, err := filepath.Glob(glob)
	if err != nil {
		errors <- fmt.Errorf("GLOB %v: %v", glob, err)
		return
	}
	if care == nil {
		care = IsDelphiFile
	}

	for _, match := range matches {
		info, err := os.Lstat(match)
		if err != nil {
			errors <- err
		}
		name := filepath.Base(match)
		if istemp(name) && name != "." {
			continue
		}
		if info.IsDir() {
			err := filepath.Walk(glob, func(file string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}
				if !care(file) {
					return nil
				}
				if istemp(info.Name()) {
					return filepath.SkipDir
				}
				filenames <- file
				return nil
			})
			if err != nil {
				errors <- err
			}
		} else {
			if !care(name) {
				continue
			}
			filenames <- match
		}
	}
}
