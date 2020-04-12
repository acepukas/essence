package main

import (
	"flag"
	"log"

	"github.com/acepukas/essence/gen"
)

func main() {

	var (
		packageName string
		srcDir      string
	)

	flag.StringVar(&packageName, "package-name", "essence",
		"package name of generated file sytem")

	flag.StringVar(&srcDir, "src-dir", "./static", "source directory")

	flag.Parse()

	err := gen.Generate(packageName, srcDir)
	if err != nil {
		log.Fatal(err)
	}

}
