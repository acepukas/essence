package vfs

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"path"
	"path/filepath"
	"strings"
)

// VirtualFileSystem extended interface for convenience methods
type VirtualFileSystem interface {
	http.FileSystem
	Bytes(string) ([]byte, error)
	String(string) (string, error)
	ParseFiles(...string) (*template.Template, error)
	ParseGlob(string) (*template.Template, error)
	ParseFilesWithFuncMap(template.FuncMap, ...string) (*template.Template, error)
	ParseGlobWithFuncMap(template.FuncMap, string) (*template.Template, error)
}

// VFSX implements extended interface
type VFSX struct {
	http.FileSystem
}

// ParseFilesWithFuncMap takes a template function map and a list of file paths,
// parses them as templates and returns a *template.Template (HTML).
func (v *VFSX) ParseFilesWithFuncMap(fnMap template.FuncMap,
	filenames ...string) (*template.Template, error) {

	if len(filenames) == 0 {
		return nil, fmt.Errorf("template: no files named in call to ParseFiles")
	}

	var t *template.Template = nil

	for _, filename := range filenames {

		file, err := v.Open(filename)
		if err != nil {
			return nil, err
		}
		defer file.Close()

		b, err := ioutil.ReadAll(file)
		if err != nil {
			return nil, err
		}

		s := string(b)

		name := path.Base(filename)

		if t == nil {
			t = template.New(name).Funcs(fnMap)
		} else {
			t = t.New(name)
		}

		_, err = t.Parse(s)
		if err != nil {
			return nil, err
		}
	}

	return t, nil
}

// ParseFiles functions the same as ParseFilesWithFuncMap minus the map
func (v *VFSX) ParseFiles(filenames ...string) (*template.Template, error) {
	return v.ParseFilesWithFuncMap(nil, filenames...)
}

// walk walks the file system passing the file path to the supplied visitor func
func walk(v VirtualFileSystem, visitor func(string)) error {

	var walk func(http.File, []string) error

	walk = func(f http.File, components []string) error {

		stat, err := f.Stat()
		if err != nil {
			return err
		}

		if !stat.IsDir() {
			path := root + strings.Join(components, sep)
			visitor(path)
			return nil
		}

		stats, err := f.Readdir(0)
		if err != nil {
			return err
		}

		for _, stat := range stats {

			subComponents := append(components, stat.Name())
			path := root + strings.Join(subComponents, sep)

			child, err := v.Open(path)
			if err != nil {
				return err
			}
			defer child.Close()

			err = walk(child, subComponents)
			if err != nil {
				return err
			}

		}

		return nil
	}

	fsRoot, err := v.Open(root)
	if err != nil {
		return err
	}
	defer fsRoot.Close()

	return walk(fsRoot, []string{})
}

// ParseGlobWithFuncMap takes a function map and a glob pattern and then forwards
// the matching files list and the function map on to ParseFilesWithFuncMap.
func (v *VFSX) ParseGlobWithFuncMap(fnMap template.FuncMap,
	pattern string) (*template.Template, error) {

	matches := make([]string, 0)

	// check for bad pattern first
	_, err := filepath.Match(pattern, "")
	if err != nil {
		return nil, err
	}

	walk(v, func(path string) {
		if ok, _ := filepath.Match(pattern, path); ok {
			matches = append(matches, path)
		}
	})

	return v.ParseFilesWithFuncMap(fnMap, matches...)
}

// ParseGlob functions the same as ParseGlobWithFuncMap minus the map
func (v *VFSX) ParseGlob(pattern string) (*template.Template, error) {
	return v.ParseGlobWithFuncMap(nil, pattern)
}

// Bytes returns file contents as a byte slice
func (v *VFSX) Bytes(filename string) ([]byte, error) {

	file, err := v.Open(filename)
	if err != nil {
		return []byte{}, err
	}
	defer file.Close()

	return ioutil.ReadAll(file)
}

// String returns file contents as a string
func (v *VFSX) String(filename string) (string, error) {

	b, err := v.Bytes(filename)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
