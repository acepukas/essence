# Essence

Essence is a virtual file system implementation for serving static content from a go web application. The content is converted to binary and stored within the go web application binary executable file.

The purpose is to bundle content together into the binary executable for easier deployment as well as having the content served from memory rather than first retrieved from disk.

This module contains a package called vfs that is referenced within application code when interacting with the virtual file system. You should never have to interact with this package directly as all references to it will be part of the generated code that essence creates when building the virtual file system code.

## Installation

Install the binary as any other golang application:

    go install github.com/acepukas/essence

After which the `essence` binary will be on your `$GOPATH`.

## Usage

You can place code generation directives in your application code. For example:

    //go:generate essence -package-name=assets_vfs -src-dir=./assets

This directive will call the essence binary and generate the necessary code when the web application binary executable is built.

When building the web application binary you can specify on the command line the "dev" tag to use the on disk file system while in development:

    go build -tags dev
