package gen

const prodTmplStr = `// +build !dev

// Code generated by github.com/acepukas/essence. DO NOT EDIT

package {{.Package}}

import (
	"time"
	"net/http"
	"html/template"
	"os"

	"github.com/acepukas/essence/vfs"
)

var fs vfs.VirtualFileSystem = &vfs.VFSX{
	FileSystem: &vfs.VFS{
		Children: []*vfs.VFile{
			{{template "node" .FS.Children}}
		},
	},
}

{{template "public_interface"}}

{{- define "common" -}}
			Name: "{{.Name}}",
			ModTime: time.Unix({{.ModTime.Unix}}, 0),
			Mode: os.FileMode(0{{printf "%o" .Mode}}),
{{- end -}}

{{- define "node" -}}
	{{- range . -}}
		{{- if .IsDir -}}
		&vfs.VFile{
			{{template "common" .}}
			Children: []*vfs.VFile{
				{{template "node" .Children}}
			},
		},
		{{else -}}
		&vfs.VFile{
			{{template "common" .}}
			Data: []byte("{{- encodeBytes .Data -}}"),
		},
		{{end -}}
	{{- end -}}
{{- end -}}`

const devTmplStr = `// +build dev

// Code generated by github.com/acepukas/essence. DO NOT EDIT

package {{.Package}}

import (
	"net/http"
	"html/template"

	"github.com/acepukas/essence/vfs"
)

var fs vfs.VirtualFileSystem = &vfs.VFSX{
	FileSystem: http.Dir("{{.SrcDir}}"),
}

{{template "public_interface"}}`

const publicInterfaceTmplStr = `{{- define "public_interface" -}}
func Instance() vfs.VirtualFileSystem {
	return fs
}

func Open(path string) (http.File, error) {
	return fs.Open(path)
}

func Bytes(path string) ([]byte, error) {
	return fs.Bytes(path)
}

func String(path string) (string, error) {
	return fs.String(path)
}

func ParseFiles(files ...string) (*template.Template, error) {
	return fs.ParseFiles(files...)
}

func ParseGlob(pattern string) (*template.Template, error) {
	return fs.ParseGlob(pattern)
}

func ParseFilesWithFuncMap(fnMap template.FuncMap, files ...string) (*template.Template, error) {
	return fs.ParseFilesWithFuncMap(fnMap, files...)
}

func ParseGlobWithFuncMap(fnMap template.FuncMap, pattern string) (*template.Template, error) {
	return fs.ParseGlobWithFuncMap(fnMap, pattern)
}{{- end -}}`
