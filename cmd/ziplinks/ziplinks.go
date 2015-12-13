package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

// show links in a zip file
func main() {
	for _, a := range os.Args[1:] {
		showZipLinks(a)
	}
}

func showZipLinks(fn string) {
	z, err := zip.OpenReader(fn)
	if err != nil {
		log.Println(err)
		return
	}
	defer z.Close()

	var buf bytes.Buffer
	for _, f := range z.File {
		showZipLink(f, &buf)
	}
}

func showZipLink(f *zip.File, tmp *bytes.Buffer) {
	if f.Mode()&os.ModeSymlink == 0 {
		return
	}

	fr, err := f.Open()
	if err != nil {
		log.Println(err)
		return
	}
	defer fr.Close()

	tmp.Reset()
	_, err = io.Copy(tmp, &io.LimitedReader{fr, 4096})
	if err != nil {
		log.Println(err)
		return
	}

	linkstr := tmp.String()

	fmt.Println(f.Name, "->", linkstr)
}
