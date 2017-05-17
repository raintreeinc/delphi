package walk

import (
	"path/filepath"
	"strings"
)

type GlobsFlag struct {
	NoDefault  bool
	Default    []string
	Additional []string
	Extensions []string
}

func (globs *GlobsFlag) IsEmpty() bool {
	return len(globs.Default) == 0 && len(globs.Additional) == 0 && len(globs.Extensions) == 0
}

func (globs *GlobsFlag) All() []string {
	if globs.NoDefault {
		return globs.Additional
	}
	return append(append([]string{}, globs.Default...), globs.Additional...)
}

func (globs *GlobsFlag) String() string {
	return strings.Join(globs.All(), ";")
}

func (globs *GlobsFlag) Set(value string) error {
	values := strings.Split(strings.Replace(value, ":", ";", -1), ";")
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}

		if strings.ContainsAny(value, "*[]?") {
			globs.Additional = append(globs.Additional, value)
		} else if value[0] == '.' {
			globs.Extensions = append(globs.Extensions, value)
		} else {
			globs.Additional = append(globs.Additional, value)
		}
	}
	return nil
}

func (globs *GlobsFlag) Matches(file string) bool {
	name := filepath.Base(file)
	ext := filepath.Ext(file)

	for _, mext := range globs.Extensions {
		if strings.EqualFold(ext, mext) {
			return true
		}
	}
	for _, glob := range globs.Default {
		if match, _ := filepath.Match(glob, name); match {
			return true
		}
	}

	for _, glob := range globs.Additional {
		if match, _ := filepath.Match(glob, name); match {
			return true
		}
	}

	return false
}
