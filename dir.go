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

func Open(path string) (*DB, error) {
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

func openIphotoDir(path string) (*DB, error) {
	db := &DB{
		dir: prefixPhotoDir(path),
	}
	err := readIphotoDB(db, filepath.Join(path, apdbPath))
	if err != nil {
		return nil, err
	}
	return db, nil
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

func openIphotoZip(path string) (*DB, error) {
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

	db := &DB{
		dir: newZipPhotoDir(z, root),
	}
	if err := readIphotoDB(db, tempDir); err != nil {
		return nil, err
	}
	ok = true
	return db, nil
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
