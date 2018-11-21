package main

import (
	"flag"
	"log"

	"github.com/fellou89/gogen/specific"
)

var (
	pkg          = flag.String("pkg", "", "generic package")
	out          = flag.String("out-dir", "", "directory to store the specific package")
	verb         = flag.String("verb", "", "REST verb to use identifying templates")
	specificType = flag.String("specific-type", "", "what specific type to use instead of interface{}")
	skipTests    = flag.Bool("skip-tests", false, "whether to skip generating test files")
)

func main() {
	flag.Parse()

	if *pkg == "" {
		flag.Usage()
		log.Fatal("missing generic package")
	}

	if *specificType == "" {
		flag.Usage()
		log.Fatal("missing specific type")
	}

	if err := specific.Process(*pkg, *out, *verb, *specificType, func(opts *specific.Options) {
		opts.SkipTestFiles = *skipTests
	}); err != nil {
		log.Fatal(err)
	}
}
