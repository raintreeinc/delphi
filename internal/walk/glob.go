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

func istemp(file string) bool {
	name := filepath.Base(file)
	return strings.HasPrefix(name, ".") || strings.HasPrefix(name, "~") || strings.HasSuffix(name, "~")
}

func Globs(globs []string, filenames chan string, errors chan error, care func(file string) bool) {
	for _, glob := range globs {
		Glob(glob, filenames, errors, care)
	}
}

func Glob(glob string, filenames chan string, errors chan error, care func(file string) bool) {
	var matches []string
	if _, err := os.Lstat(glob); err == nil {
		matches = []string{glob}
	} else {
		var err error
		matches, err = filepath.Glob(glob)
		if err != nil {
			errors <- fmt.Errorf("GLOB %v: %v", glob, err)
			return
		}
	}

	if care == nil {
		care = IsDelphiFile
	}

	for _, match := range matches {
		info, err := os.Lstat(match)
		if err != nil {
			errors <- err
		}

		if istemp(match) && match != glob {
			continue
		}

		if info.IsDir() {
			err = filepath.Walk(match, func(file string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				name := filepath.Base(file)
				istmp := istemp(file)
				if info.IsDir() && istmp {
					if name == "." || name == ".." {
						return nil
					}
					return filepath.SkipDir
				}
				if !care(file) {
					return nil
				}
				if istmp {
					return nil
				}
				filenames <- file
				return nil
			})
			if err != nil {
				errors <- err
			}
		} else {
			if !care(match) {
				continue
			}
			filenames <- match
		}
	}
}
