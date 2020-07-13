# Essence

Essence is a virtual file system implementation for embedding binary file data within a go binary executable file.

The purpose is to bundle file content together into the binary executable for easier deployment as well as having the content served from memory rather than first retrieved from disk.

This module contains a package called vfs that is referenced within application code when interacting with the virtual file system. This is a dependency of the generated code but application code should never have to interact with this package directly.

## Installation

Install the binary as any other golang application:

    go install github.com/acepukas/essence

After which the `essence` binary will be on your `$GOPATH`.

## Usage

You can place code generation directives in your application code. For example:

    //go:generate essence -package-name=static_vfs -src-dir=./static

The command line flags demonstrated here are the default values if they are not given. In this particular example the `static` directory (provided it resides within the currently working directory) will be scanned recursively for files and subdirectories. A new directory will be created called `static_vfs` (in the current directory where essence was run) which will contain the generated code that comprises the virtual file system as well as the development stub that uses the on disk file system.

Within your application you can then import the generated package (`static_vfs` in this example):

```go
import (

  ...

  vfs "github.com/username/project_name/static_vfs"

)
```

Adjust the module path as you need to for your project. It's assumed that the package alias `vfs` is being used within the rest of this README doc.

Then refer to the virtual file system with

```go
file, err := vfs.Open("/path/to/file")
```

When static content changes you must rerun the `go generate` command each time.

## Public Interface

The virtual file system uses an extended interface beyond the `http.FileSystem` interface, which it embeds. These are

```go
String(path string) (string, error)
Bytes(path string) ([]byte, error)
ParseFiles(paths ...string) (*template.Template, error)
ParseGlob(pattern string) (*template.Template, error)
ParseFilesWithFuncMap(template.FuncMap, paths ...string) (*template.Template, error)
ParseGlobWithFuncMap(template.FuncMap, pattern string) (*template.Template, error)
```

Example:

```go
fileStr, err := vfs.String("/path/to/file")
```

This API is the public interface that wraps around a single instance of a virtual file system within the generated package code.

In order to use this wrapper virtual file system directly you can acquire the instance with

```go
fs := vfs.Instance()
```

At which point you could then pass the instance to `http.FileServer()` for example since the virtual file server already implements the  `http.FileSystem` interface.

## Build

When building the web application binary you can specify on the command line the "dev" tag to use the on disk file system while in development:

    go build -tags=dev

## Development

In order to ensure that tests function properly it's necessary to generate the virtual file system code as the generated files are not tracked with version control. In the root of the essence module, run

    go run main.go

with no arguments and the `static_vfs` directory and files contained within will be generated, at which point tests can be run.
