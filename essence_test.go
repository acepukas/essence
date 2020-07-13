package main

import (
	"bytes"
	"html/template"
	"io"
	"testing"

	svfs "github.com/acepukas/essence/static_vfs"
)

func TestVFS(t *testing.T) {
	file, err := svfs.Open("/hello_essence.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, file)

	if buf.String() != "hello essence\n" {
		t.FailNow()
	}
}

func TestFileReRead(t *testing.T) {

	file1, err := svfs.Open("/hello_essence.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file1.Close()

	file2, err := svfs.Open("/hello_essence.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file2.Close()

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, file2)

	if buf.String() != "hello essence\n" {
		t.FailNow()
	}

}

func TestSeekStart(t *testing.T) {
	file, err := svfs.Open("/hello_essence.txt")
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()

	file.Seek(6, io.SeekStart)

	buf := new(bytes.Buffer)
	_, err = io.Copy(buf, file)

	if buf.String() != "essence\n" {
		t.Logf("%q\n", buf.String())
		t.FailNow()
	}
}

func TestString(t *testing.T) {
	s, err := svfs.String("/hello_essence.txt")
	if err != nil {
		t.Fatal(err)
	}
	if s != "hello essence\n" {
		t.FailNow()
	}
}

func TestParseFiles(t *testing.T) {
	tmpl, err := svfs.ParseFiles("/tmpl.tmpl", "/subdir/subtmpl.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	data := struct{ Message, InnerMessage string }{"a", "b"}
	buf := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(buf, "tmpl.tmpl", data)
	if err != nil {
		t.Fatal(err)
	}
	res := buf.String()
	if res != "subject: a - this is sub template: b\n\n" {
		t.FailNow()
	}
}

func TestParseFilesWithFuncs(t *testing.T) {
	fnMap := map[string]interface{}{
		"safe": func(s string) template.HTML {
			return template.HTML(s)
		},
	}
	tmpl, err := svfs.ParseFilesWithFuncMap(fnMap, "/functions.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	data := struct{ Data string }{"<p>paragraph</p>"}
	buf := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(buf, "functions.tmpl", data)
	if err != nil {
		t.Fatal(err)
	}
	res := buf.String()
	if res != "html: <p>paragraph</p>\n" {
		t.FailNow()
	}
}

func TestParseGlob(t *testing.T) {
	tmpl, err := svfs.ParseGlob("/subdir/*.tmpl")
	if err != nil {
		t.Fatal(err)
	}
	data := struct{ InnerMessage string }{"b"}
	buf := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(buf, "subtmpl.tmpl", data)
	if err != nil {
		t.Fatal(err)
	}
	res := buf.String()
	if res != "this is sub template: b\n" {
		t.FailNow()
	}
}
