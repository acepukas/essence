// vfs is a virtual file system implementation
package vfs

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// breadthFirst finds a file within the file system based on path components array.
func breadthFirst(files []*VFile, components []string) *VFile {

	for _, f := range files {
		if f.Name == components[0] {
			if len(components) == 1 {
				return f
			}
			return breadthFirst(f.Children, components[1:])
		}
	}

	return nil

}

// pathComponents splits a path into separate components as a string array.
func pathComponents(p string) []string {
	pathSep := "/"
	start := []string{pathSep}
	p = strings.Trim(p, pathSep)
	if p == "" {
		return start
	}
	return append(start, strings.Split(p, pathSep)...)
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
			NewDir("/", time.Now()),
		},
	}
}

// Open retrieves a file by name
func (v *VFS) Open(filename string) (http.File, error) {

	v.mu.RLock()
	defer v.mu.RUnlock()

	filename = filepath.Clean(filename)

	file := breadthFirst(v.Children, pathComponents(filename))

	if file == nil {
		return nil, &os.PathError{
			Op:   "open",
			Path: filename,
			Err:  os.ErrNotExist,
		}
	}

	// if !file.isDir {
	// file.Seek(0, io.SeekStart)
	// }

	return file, nil
}

// VFile structure that represents a virtual file
type VFile struct {
	At       int64
	Name     string
	Data     []byte
	ModTime  time.Time
	IsDir    bool
	Children []*VFile
	mu       sync.RWMutex
}

// NewDir creates a new virtual directory file instance
func NewDir(name string, modTime time.Time) *VFile {
	return &VFile{
		At:       0,
		Name:     name,
		Data:     []byte{},
		ModTime:  modTime,
		IsDir:    true,
		Children: []*VFile{},
	}
}

// NewFile creates a new virtual file instance
func NewFile(name string, modTime time.Time, data []byte) *VFile {
	return &VFile{
		At:       0,
		Name:     name,
		Data:     data,
		ModTime:  modTime,
		IsDir:    false,
		Children: []*VFile{},
	}
}

// Append a file to a directory. Will not append files to files, only dirs
func (f *VFile) Append(file *VFile) {
	if f.IsDir {
		f.Children = append(f.Children, file)
	}
}

// Close a file
func (f *VFile) Close() error {
	return nil
}

// Stat retrieves info about a file
func (f *VFile) Stat() (os.FileInfo, error) {
	return &VFileInfo{f}, nil
}

// Readdir retreives file info about all files within a directory
func (f *VFile) Readdir(count int) ([]os.FileInfo, error) {

	f.mu.RLock()
	defer f.mu.RUnlock()

	numFiles := len(f.Children)

	res := make([]os.FileInfo, numFiles)

	i := 0
	for _, file := range f.Children {
		res[i], _ = file.Stat()
		i++
	}

	return res, nil
}

// Read writes file data to a byte array
func (f *VFile) Read(b []byte) (int, error) {

	f.mu.Lock()
	defer f.mu.Unlock()

	i := 0

	// NOTE: resetting the pointer here could cause problems as it makes seek
	//       pointless
	f.At = 0

	for f.At < int64(len(f.Data)) && i < len(b) {
		b[i] = f.Data[f.At]
		i++
		f.At++
	}

	return i, nil
}

// Seek moves the file pointer to a particular byte offset
func (f *VFile) Seek(offset int64, whence int) (int64, error) {

	f.mu.Lock()
	defer f.mu.Unlock()

	switch whence {
	case io.SeekStart:
		f.At = offset
	case io.SeekCurrent:
		f.At += offset
	case io.SeekEnd:
		f.At = int64(len(f.Data)) + offset
	default:
		return 0, fmt.Errorf("invalid seek constant")
	}

	return f.At, nil
}

// VFileInfo structure represents virtual file information
type VFileInfo struct {
	file *VFile
}

// Name retrieves file's name
func (s *VFileInfo) Name() string {

	s.file.mu.RLock()
	defer s.file.mu.RUnlock()

	return s.file.Name
}

// ModTime retrieves file's last modification time
func (s *VFileInfo) ModTime() time.Time {

	s.file.mu.RLock()
	defer s.file.mu.RUnlock()

	return s.file.ModTime
}

// IsDir returns true if file node is a directory
func (s *VFileInfo) IsDir() bool {

	s.file.mu.RLock()
	defer s.file.mu.RUnlock()

	return s.file.IsDir
}

// Sys simply retuns nil
func (s *VFileInfo) Sys() interface{} {
	return nil
}

// Size returns the interger byte size of file
func (s *VFileInfo) Size() int64 {

	s.file.mu.RLock()
	defer s.file.mu.RUnlock()

	size := int64(len(s.file.Data))

	return size
}

// Mode retrieves file's unix permissions flags
func (s *VFileInfo) Mode() os.FileMode {
	if s.IsDir() {
		return os.FileMode(0755)
	}
	return os.FileMode(0644)
}
