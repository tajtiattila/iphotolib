package iphotolib

import (
	"io"
	"os"
	"time"
)

type Photo struct {
	Path     string
	Date     time.Time
	FileSize int64
	FileName string
	Name     string
	Desc     string
	Rating   int

	Event EventKey
	Place PlaceKey

	Hidden   bool
	Flagged  bool
	Original bool
	InTrash  bool

	// dir is user to open and stat image
	dir photoDir
}

func (p *Photo) Stat() (os.FileInfo, error) {
	return p.dir.Stat(p.Path)
}

func (p *Photo) Open() (io.ReadCloser, error) {
	return p.dir.Open("Masters/" + p.Path)
}

func (p *Photo) OpenThumb() (io.ReadCloser, error) {
	return p.dir.Open("Thumbnails/" + p.Path)
}
