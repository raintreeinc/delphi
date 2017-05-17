package delphi

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/raintreeinc/delphi/internal/walk"
)

func SearchPath() string {
	if path := os.Getenv("DELPHI_SEARCH"); path != "" {
		return path
	}
	return ""
}

func TempDir() string {
	if dir := os.Getenv("DELPHI_TEMP"); dir != "" {
		return dir
	}
	return os.TempDir()
}

func DCC() string {
	return "c:\\Program Files (x86)\\Borland\\Delphi7\\Bin\\dcc32.exe"
}

func SearchPathFromRoot(root string) []string {
	filenames := make(chan string, 8)
	errors := make(chan error, 8)
	go func() {
		walk.Glob(root, filenames, errors, walk.IsDelphiFile)
		close(filenames)
		close(errors)
	}()

	go func() {
		for range errors {
		}
	}()

	paths := []string{}
	for file := range filenames {
		ext := filepath.Ext(file)
		if strings.EqualFold(ext, ".pas") || strings.EqualFold(ext, ".inc") {
			dir := filepath.Dir(file)
			if !contains(dir, paths) {
				paths = append(paths, dir)
			}
		}
	}

	return paths
}

func contains(value string, list []string) bool {
	for _, item := range list {
		if strings.EqualFold(item, value) {
			return true
		}
	}
	return false
}
