package gen

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/acepukas/essence/vfs"
)

// reEncodeJson is used for stripping json files of white space for better compression
func reEncodeJson(reader io.Reader) (*bytes.Buffer, error) {

	errFmt := "reencode json: %v\n"

	jsonMap := make(map[string]interface{})

	err := json.NewDecoder(reader).Decode(&jsonMap)
	if err != nil {
		return nil, fmt.Errorf(errFmt, err)
	}

	buf := new(bytes.Buffer)

	err = json.NewEncoder(buf).Encode(jsonMap)
	if err != nil {
		return nil, fmt.Errorf(errFmt, err)
	}

	return buf, nil

}

var fnMap = map[string]interface{}{
	"encodeBytes": func(bts []byte) string {
		strBldr := new(strings.Builder)
		for _, b := range bts {
			strBldr.WriteString(fmt.Sprintf("\\x%02X", b))
		}
		return strBldr.String()
	},
}

// vfsSpec is a utility structure containing necessary fields for write the file system to binary.
type vfsSpec struct {
	Package string
	SrcDir  string
	FS      *vfs.VFS
}

// writeFile creates the file that will contain the binary vfs or stub for development phase.
func writeFile(tmplStr, suffix string, spec *vfsSpec) error {

	errFmt := "write file: %v\n"

	if suffix != "" {
		suffix = "_" + suffix
	}

	filename := fmt.Sprintf("%s/%s%s.go", spec.Package, spec.Package, suffix)

	out, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}
	defer out.Close()

	tmpl, err := template.New("").Funcs(fnMap).Parse(tmplStr)
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	tmpl, err = tmpl.Parse(publicInterfaceTmplStr)
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	buf := new(bytes.Buffer)
	err = tmpl.Execute(buf, spec)
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	data, err := format.Source(buf.Bytes())
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	if err := ioutil.WriteFile(filename, data, os.FileMode(0644)); err != nil {
		return fmt.Errorf(errFmt, err)
	}

	pwd, err := filepath.Abs(".")
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	fmt.Printf("ESSENCE: file written: %s/%s\n", pwd, filename)

	return nil
}

// buildTree recursively builds a mirrored virtual file system from the provided path. Requires virtual file array.
func buildTree(path string, vfileStack []*vfs.VFile) error {

	errFmt := "build tree: %v\n"

	stats, err := ioutil.ReadDir(path)
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	for _, node := range stats {

		nodePath := fmt.Sprintf("%s/%s", path, node.Name())

		if node.IsDir() {
			vDir := vfs.NewFile(node)
			// append virtual directory to last element in vfs
			vfileStack[len(vfileStack)-1].Append(vDir)
			// virtual directory becomes new last
			// element of vfs within recursive call
			err := buildTree(nodePath, append(vfileStack, vDir))
			if err != nil {
				return fmt.Errorf(errFmt, err)
			}

			continue
		}

		fileBytes, err := ioutil.ReadFile(nodePath)
		if err != nil {
			return fmt.Errorf(errFmt, err)
		}

		buf := bytes.NewBuffer(fileBytes)

		// strip json of white space to save space in generated binary
		if strings.HasSuffix(node.Name(), ".json") {
			buf, err = reEncodeJson(buf)
			if err != nil {
				return fmt.Errorf(errFmt, err)
			}
		}

		vfile := vfs.NewFile(node, buf.Bytes()...)
		vfileStack[len(vfileStack)-1].Append(vfile)

		fmt.Printf("ESSENCE: embedded file: %s/%s\n", path, node.Name())
	}

	return nil
}

// Generate virtual file system. Requires package name that will be used when generating executable code and the "real" file system directory.
func Generate(packageName, srcDir string) error {

	errFmt := "generate: %v\n"

	absPath, err := filepath.Abs(srcDir)
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	fs := vfs.New()

	err = buildTree(absPath, fs.Children)
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	err = os.MkdirAll(packageName, os.FileMode(0755))
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	spec := vfsSpec{packageName, absPath, fs}

	err = writeFile(prodTmplStr, "", &spec)
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	err = writeFile(devTmplStr, "dev", &spec)
	if err != nil {
		return fmt.Errorf(errFmt, err)
	}

	return nil

}
