package gen

const prodTmplStr = `// +build !dev

// Code generated by github.com/acepukas/essence. DO NOT EDIT

package {{.Package}}

import (
	"time"

	"github.com/acepukas/essence/vfs"
)

var VFS = &vfs.VFS{
	Children: []*vfs.VFile{
		{{template "node" .FS.Children}}
	},
}

{{- define "node" -}}
	{{- range . -}}
		{{- if .IsDir -}}
		&vfs.VFile{
			Name: "{{.Name}}",
			ModTime: time.Unix({{.ModTime.Unix}}, 0),
			IsDir: true,
			Children: []*vfs.VFile{
				{{template "node" .Children}}
			},
		},
		{{else -}}
		&vfs.VFile{
			Name: "{{.Name}}",
			Data: []byte("{{- range .Data -}}{{printf "\\x%02X" .}}{{- end -}}"),
			ModTime: time.Unix({{.ModTime.Unix}}, 0),
		},
		{{end -}}
	{{- end -}}
{{- end -}}`

const devTmplStr = `// +build dev

// Code generated by github.com/acepukas/essence. DO NOT EDIT

package {{.Package}}

import "net/http"

var VFS = http.Dir("{{.SrcDir}}")`
