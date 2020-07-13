// vfs is a virtual file system implementation
package vfs

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"strings"
	"sync"
	"time"
)

var sep = "/"
var root = sep

// breadthFirstSearch finds a file within the file system based on path
// components array.
func breadthFirstSearch(files []*VFile, components []string) *VFile {

	for _, f := range files {
		if f.Name == components[0] {
			if len(components) == 1 {
				return f
			}
			return breadthFirstSearch(f.Children, components[1:])
		}
	}

	return nil

}

// pathComponents splits a path into separate components as a string array.
func pathComponents(p string) []string {
	start := []string{root}
	p = strings.Trim(p, sep)
	if p == "" {
		return start
	}
	return append(start, strings.Split(p, sep)...)
}

// find is just a convenient wrapper around breadthFirstSearch
func find(v *VFS, filename string) (*VFile, error) {

	file := breadthFirstSearch(v.Children, pathComponents(filename))

	if file == nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: filename,
			Err:  os.ErrNotExist,
		}
	}

	return file, nil

}

// VFS is the structure that represents the virtual file system.
type VFS struct {
	Children []*VFile
	mu       sync.RWMutex
}

// New creates a new VFS instance.
func New() *VFS {
	return &VFS{
		Children: []*VFile{
			NewFile(NewFileInfo(root, time.Now(), os.ModeDir|0755, 0)),
		},
	}
}

// Open retrieves a file by name
func (v *VFS) Open(filename string) (http.File, error) {

	v.mu.RLock()
	defer v.mu.RUnlock()

	if !path.IsAbs(filename) {
		return nil, errors.New("vfs: path must be absolute")
	}

	filename = path.Clean(filename)
	file, err := find(v, filename)
	if err != nil {
		return nil, err
	}

	file.Seek(0, io.SeekStart)

	return file, nil
}

// VFile structure that represents a virtual file
type VFile struct {
	At       int64
	Name     string
	Data     []byte
	ModTime  time.Time
	Mode     os.FileMode
	Children []*VFile
	mu       sync.RWMutex
}

// NewFile creates a new virtual file instance
func NewFile(info os.FileInfo, data ...byte) *VFile {
	return &VFile{
		At:       0,
		Name:     info.Name(),
		Data:     data,
		ModTime:  info.ModTime(),
		Mode:     info.Mode(),
		Children: []*VFile{},
	}
}

// IsDir is an intenal convenience used when building the file system tree and
// rendering the file system template.
func (f *VFile) IsDir() bool {
	return f.Mode&os.ModeDir != 0
}

// Append a file to a directory. Will not append files to files, only dirs
func (f *VFile) Append(file *VFile) {
	if f.IsDir() {
		f.Children = append(f.Children, file)
	}
}

// Close a file
func (f *VFile) Close() error {
	f.Seek(0, io.SeekStart)
	return nil
}

// Stat retrieves info about a file
func (f *VFile) Stat() (os.FileInfo, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	info := NewFileInfo(f.Name, f.ModTime, f.Mode, int64(len(f.Data)))
	return info, nil
}

// Readdir retreives file info about all files within a directory
func (f *VFile) Readdir(count int) ([]os.FileInfo, error) {

	f.mu.RLock()
	defer f.mu.RUnlock()

	res := make([]os.FileInfo, len(f.Children))

	for i, file := range f.Children {
		res[i], _ = file.Stat()
	}

	return res, nil
}

// Read writes file data to a byte array
func (f *VFile) Read(b []byte) (int, error) {

	f.mu.Lock()
	defer f.mu.Unlock()

	i := 0
	l := int64(len(f.Data))

	for f.At < l && i < len(b) {
		b[i] = f.Data[f.At]
		i++
		f.At++
	}

	if f.At >= l {
		return i, io.EOF
	}

	return i, nil
}

var outRangeErr = errors.New("offset outside byte range")
var invalidSeekType = errors.New("invalid seek constant")

// Seek moves the file pointer to a particular byte offset
func (f *VFile) Seek(offset int64, whence int) (int64, error) {

	f.mu.Lock()
	defer f.mu.Unlock()

	end := int64(len(f.Data)) - 1

	switch whence {
	case io.SeekStart:
		if offset < 0 || offset > end {
			return 0, outRangeErr
		}
		f.At = offset
	case io.SeekCurrent:
		newOffset := f.At + offset
		if newOffset < 0 || newOffset > end {
			return 0, outRangeErr
		}
		f.At = newOffset
	case io.SeekEnd:
		newOffset := end - offset
		if newOffset < 0 || newOffset > end {
			return 0, outRangeErr
		}
		f.At = newOffset
	default:
		return 0, invalidSeekType
	}

	return f.At, nil
}

// VFileInfo structure represents virtual file information
type VFileInfo struct {
	FileName    string
	FileModTime time.Time
	FileMode    os.FileMode
	FileSize    int64
}

func NewFileInfo(name string, modTime time.Time,
	mode os.FileMode, size int64) *VFileInfo {
	return &VFileInfo{
		FileName:    name,
		FileModTime: modTime,
		FileMode:    mode,
		FileSize:    size,
	}
}

// Name retrieves file's name
func (s *VFileInfo) Name() string {
	return s.FileName
}

// ModTime retrieves file's last modification time
func (s *VFileInfo) ModTime() time.Time {
	return s.FileModTime
}

// IsDir returns true if file node is a directory
func (s *VFileInfo) IsDir() bool {
	return s.FileMode&os.ModeDir != 0
}

// Sys simply retuns nil
func (s *VFileInfo) Sys() interface{} {
	return nil
}

// Size returns the interger byte size of file
func (s *VFileInfo) Size() int64 {
	return s.FileSize
}

// Mode retrieves file's unix permissions flags
func (s *VFileInfo) Mode() os.FileMode {
	return s.FileMode
}
