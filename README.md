# danos-buildimage

This tool aids in building DANOS iso images using containers. It uses the docker API and should work with any compatibile container engine.

## Installation

Binaries are included for the Release, the appropriate binary for your operating system can be placed in your PATH.

From source:

```
$ git clone https://github.com/jsouthworth/danos-buildimage
$ cd danos-buildimage
$ go install jsouthworth.net/go/danos-buildimage/cmd/danos-buildimage
```

## Usage

```
$ danos-buildimage -h
Usage of danos-buildimage:
  -dest string
    	destination directory (default ".")
  -pkg string
    	preferred package directory
  -version string
    	version of danos to build for (default "latest")
```

