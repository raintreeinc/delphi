package regex

import (
	"bytes"
	"io/ioutil"
	"regexp"
)

type File struct {
	Path     string
	Source   []byte
	Modified []byte
}

func LoadFile(path string) (*File, error) {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	return &File{path, src, src}, nil
}

func (file *File) Changed() bool {
	return !bytes.Equal(file.Source, file.Modified)
}

func (file *File) WriteChanges() error {
	return ioutil.WriteFile(file.Path, file.Modified, 0755)
}

func (file *File) CountRegular(re *regexp.Regexp, counter *Counter, ignoreCase, ignoreSpace bool) {
	re.ReplaceAllFunc(file.Source, func(match []byte) []byte {
		counter.Add(file.Path, string(match), ignoreCase, ignoreSpace)
		return match
	})
}

func (file *File) Replace(re *regexp.Regexp, replacement string) {
	file.Modified = re.ReplaceAll(file.Source, []byte(replacement))
}
