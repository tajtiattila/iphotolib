// Copyright (c) 2015 Attila Tajti
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.  IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

// Package iphotolib provides access to Apple® iPhoto® databases.
// An iPhoto Library may be accessed either directly
// (typically on OSX) or it can be packed within a zip file.
package iphotolib

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"

	"golang.org/x/text/unicode/norm"
)

const apdbPath = "Database/apdb"

// Open opens the database for reading. The provided path
// should be either a zip file or point to an iPhoto Library folder.
// Open reads all data from the internal sqlite databases
// and provides access to images and thumbnails.
func Open(path string) (*Lib, error) {
	fi, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if fi.IsDir() {
		return openIphotoDir(path)
	}
	return openIphotoZip(path)
}

type photoDir interface {
	Stat(fn string) (os.FileInfo, error)
	Open(fn string) (io.ReadCloser, error)
	Close() error
}

func openIphotoDir(path string) (*Lib, error) {
	lib := &Lib{
		dir: prefixPhotoDir(path),
	}
	err := readIphotoDB(lib, filepath.Join(path, apdbPath))
	if err != nil {
		return nil, err
	}
	return lib, nil
}

type prefixPhotoDir string

func (d prefixPhotoDir) Stat(fn string) (os.FileInfo, error) {
	return os.Stat(filepath.Join(string(d), fn))
}

func (d prefixPhotoDir) Open(fn string) (io.ReadCloser, error) {
	return os.Open(filepath.Join(string(d), fn))
}

func (d prefixPhotoDir) Close() error {
	return nil
}

func openIphotoZip(path string) (*Lib, error) {
	z, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	ok := false
	defer func() {
		if !ok {
			z.Close()
		}
	}()

	root, err := findZipDBRoot(&z.Reader)
	if err != nil {
		return nil, err
	}
	prefix := root + "/" + apdbPath + "/"

	dbf := make([]*zip.File, 0, 3)
	// find DB files
	dbfn := []string{"Library.apdb", "Properties.apdb", "Faces.db"}
	for _, f := range z.File {
		s := filepath.ToSlash(f.Name)
		if !strings.HasPrefix(s, prefix) {
			continue
		}
		for _, n := range dbfn {
			if s[len(prefix):] == n {
				dbf = append(dbf, f)
			}
		}
	}
	if len(dbf) != 3 {
		// TODO(tajti): check individual names and report better error?
		return nil, fmt.Errorf("iphoto: can't find all db files")
	}

	// copy db files to temp dir so sqlite can access them
	tempDir, err := ioutil.TempDir("", "iphoto")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(tempDir)
	for _, f := range dbf {
		if err := extractFile(tempDir, f); err != nil {
			return nil, err
		}
	}

	lib := &Lib{
		dir: newZipPhotoDir(z, root),
	}
	if err := readIphotoDB(lib, tempDir); err != nil {
		return nil, err
	}
	ok = true
	return lib, nil
}

type zipPhotoDir struct {
	*zip.ReadCloser
	// m maps lowercase path using slashes to files
	m map[string]*zip.File
}

func newZipPhotoDir(z *zip.ReadCloser, base string) *zipPhotoDir {
	base = zipPath(base)
	if base != "" {
		// clean removes last slash
		base += "/"
	}
	m := make(map[string]*zip.File)
	for _, f := range z.File {
		s := zipPath(f.Name)
		if strings.HasPrefix(s, base) {
			m[s[len(base):]] = f
		}
	}
	return &zipPhotoDir{z, m}
}

func (d *zipPhotoDir) Stat(fn string) (os.FileInfo, error) {
	zf := d.m[zipPath(fn)]
	if zf == nil {
		return nil, os.ErrNotExist
	}
	return zf.FileInfo(), nil
}

func (d *zipPhotoDir) Open(fn string) (io.ReadCloser, error) {
	zf := d.m[zipPath(fn)]
	if zf == nil {
		return nil, os.ErrNotExist
	}
	return zf.Open()
}

func zipPath(p string) string {
	return filepath.ToSlash(filepath.Clean(strings.ToLower(norm.NFC.String(p))))
}

func extractFile(destdir string, f *zip.File) error {
	r, err := f.Open()
	if err != nil {
		return err
	}
	defer r.Close()

	w, err := os.Create(filepath.Join(destdir, filepath.Base(f.Name)))
	if err != nil {
		return err
	}
	defer w.Close()

	_, err = io.Copy(w, r)
	return err
}

func findZipDBRoot(z *zip.Reader) (string, error) {
	// find directory with Database/apdb in it
	for _, f := range z.File {
		s := filepath.ToSlash(f.Name)
		if i := strings.Index(s, apdbPath); i >= 0 {
			if i != 0 && s[i-1] != '/' {
				continue
			}
			if i+len(apdbPath) < len(s) && s[i+len(apdbPath)] != '/' {
				continue
			}
			return path.Clean(s[:i]), nil
		}
	}
	return "", fmt.Errorf("iphoto: can't find '%s' within zip", apdbPath)
}
