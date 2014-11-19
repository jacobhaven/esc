package main

import (
	"bytes"
	"compress/gzip"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
)

var flagOut = flag.String("f", "", "Output file, else stdout.")
var flagPkg = flag.String("p", "main", "Package.")

func main() {
	flag.Parse()
	var err error
	var fnames []string
	content := make(map[string][]byte)
	for _, base := range flag.Args() {
		files := []string{base}
		for len(files) > 0 {
			fname := files[0]
			files = files[1:]
			f, err := os.Open(fname)
			if err != nil {
				log.Fatal(err)
			}
			fi, err := f.Stat()
			if err != nil {
				log.Fatal(err)
			}
			if fi.IsDir() {
				fis, err := f.Readdir(0)
				if err != nil {
					log.Fatal(err)
				}
				for _, fi := range fis {
					files = append(files, filepath.Join(fname, fi.Name()))
				}
			} else {
				fnames = append(fnames, fname)
				b, err := ioutil.ReadAll(f)
				if err != nil {
					log.Fatal(err)
				}
				var buf bytes.Buffer
				gw := gzip.NewWriter(&buf)
				if _, err := gw.Write(b); err != nil {
					log.Fatal(err)
				}
				if err := gw.Close(); err != nil {
					log.Fatal(err)
				}
				b = buf.Bytes()
				content[fname] = b
			}
			f.Close()
		}
	}
	sort.Strings(fnames)
	w := os.Stdout
	if *flagOut != "" {
		if w, err = os.Create(*flagOut); err != nil {
			log.Fatal(err)
		}
		defer w.Close()
	}
	fmt.Fprintf(w, header, *flagPkg)
	for _, fname := range fnames {
		fmt.Fprintf(w, "\n\t%q: %q,\n", fname, string(content[fname]))
	}
	fmt.Fprint(w, footer)
}

const (
	header = `package %s

import (
	"net/http"
	"os"
)

type localFS struct{}
var local localFS

type staticFS struct{}
var static staticFS

type file struct {
}

func (_ localFS) Open(name string) (http.File, error) {
	return os.Open(name)
}

func (_ staticFS) Open(name string) (http.File, error) {
	gb, present := data[name]
	if !present {
		return nil, os.ErrNotExist
	}
	gr := gzip.NewReader
}

// FS returns a http.Filesystem for the embedded assets. If local is true,
// the filesystem's contents are instead used.
func FS(local true) http.FileSystem {
	if local {
		return local
	}
	return static
}

var data = map[string]string{
`
	footer = `}
`
)
