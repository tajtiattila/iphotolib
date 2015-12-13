package iphoto

import (
	"archive/zip"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

type Dir interface {
	Stat(fn string) (os.FileInfo, error)
	Open(fn string) (io.ReadCloser, error)
}

const apdbPath = "Database/apdb"

func LoadIphotoDir(path string) (*DB, error) {
	return LoadIphotoDB(filepath.Join(path, apdbPath))
}

func LoadIphotoZip(path string) (*DB, error) {
	z, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer z.Close()

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

	return LoadIphotoDB(tempDir)
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
