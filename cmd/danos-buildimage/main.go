package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	bimg "jsouthworth.net/go/danos-buildimage"
)

var srcDir, destDir, pkgDir, version string
var noClean bool

func init() {
	flag.StringVar(&srcDir, "src", "", "source directory")
	flag.StringVar(&destDir, "dest", "..", "destination directory")
	flag.StringVar(&pkgDir, "pkg", "", "preferred package directory")
	flag.StringVar(&version, "version", "latest", "version of danos to build for")
	flag.BoolVar(&noClean, "no-clean", false, "don't delete the container when done")
}

func resolvePath(in string) string {
	if in == "" {
		return ""
	}
	out, err := filepath.Abs(in)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	return out
}

func main() {
	flag.Parse()
	b, err := bimg.MakeBuilder(
		bimg.SourceDirectory(resolvePath(srcDir)),
		bimg.DestinationDirectory(resolvePath(destDir)),
		bimg.PreferredPackageDirectory(resolvePath(pkgDir)),
		bimg.RemoveContainer(!noClean),
		bimg.Version(version),
	)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		flag.Usage()
		os.Exit(1)
	}
	err = b.Build()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
